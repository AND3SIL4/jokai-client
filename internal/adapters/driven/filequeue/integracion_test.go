package filequeue

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"jokai-client/internal/adapters/driven/identificadores"
	"jokai-client/internal/adapters/driven/snapshotlector"
	"jokai-client/internal/domain"
)

// T020: fija el contrato HU3 a nivel de archivos sin jokai-server.
// El agente y el consolidador son simulados por el test según el contrato
// (contracts/protocolo-fileserver.md §2-§8); el emisor y el LectorSnapshot son reales.

type cmdJSON struct {
	VersionFormato int    `json:"version_formato"`
	ID             string `json:"id"`
	BotNombre      string `json:"bot_nombre"`
	BotAmbiente    string `json:"bot_ambiente"`
	BotRutaEntrada string `json:"bot_ruta_entrada"`
	Maquina        string `json:"maquina"`
	Usuario        string `json:"usuario"`
	Fecha          string `json:"fecha"`
}

// publicarArchivo reproduce la publicación atómica del contrato: .tmp → rename.
func publicarArchivo(t *testing.T, ruta string, datos []byte) {
	t.Helper()
	if err := os.WriteFile(ruta+domain.SufijoTemp, datos, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(ruta+domain.SufijoTemp, ruta); err != nil {
		t.Fatal(err)
	}
}

func publicarST(t *testing.T, raiz, id, solicitudID string, estado domain.EstadoEjecucion, inicio string, fin *string, codigo *int, mensaje *string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(raiz, domain.CarpetaEstado), 0o755); err != nil {
		t.Fatal(err)
	}
	reporte := map[string]any{
		"version_formato": domain.VersionFormatoActual,
		"id":              id,
		"solicitud_id":    solicitudID,
		"estado":          estado,
		"inicio":          inicio,
		"fin":             fin,
		"codigo":          codigo,
		"mensaje":         mensaje,
	}
	datos, err := json.Marshal(reporte)
	if err != nil {
		t.Fatal(err)
	}
	publicarArchivo(t, filepath.Join(raiz, domain.CarpetaEstado, id+domain.SufijoEstado), datos)
}

func marshalarSnapshot(t *testing.T, comando cmdJSON, ejecucionID, inicio string, estadoSolicitud, estadoEjecucion domain.EstadoEjecucion, fin *string, codigo *int, mensaje *string) []byte {
	t.Helper()
	generado := "2026-09-23T10:02:31Z"
	snap := map[string]any{
		"version_formato": domain.VersionFormatoActual,
		"generado_en":     generado,
		"maquinas": []any{
			map[string]any{
				"nombre":        comando.Maquina,
				"ambiente":      comando.BotAmbiente,
				"ultimo_latido": generado,
				"bots":          []string{comando.BotNombre},
			},
		},
		"solicitudes": []any{
			map[string]any{
				"id":           comando.ID,
				"bot_nombre":   comando.BotNombre,
				"bot_ambiente": comando.BotAmbiente,
				"maquina":      comando.Maquina,
				"usuario":      comando.Usuario,
				"fecha":        comando.Fecha,
				"estado":       estadoSolicitud,
			},
		},
		"ejecuciones": []any{
			map[string]any{
				"id":           ejecucionID,
				"solicitud_id": comando.ID,
				"estado":       estadoEjecucion,
				"inicio":       inicio,
				"fin":          fin,
				"codigo":       codigo,
				"mensaje":      mensaje,
			},
		},
	}
	datos, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	return datos
}

// drenar simula el consolidador: toma el .st y mueve los archivos a procesados/.
func drenar(t *testing.T, raiz, maquina, solicitudID, ejecucionID string) {
	t.Helper()
	destino := filepath.Join(raiz, domain.CarpetaProcesados, time.Now().UTC().Format("2006/01/02"))
	if err := os.MkdirAll(destino, 0o755); err != nil {
		t.Fatal(err)
	}
	st := filepath.Join(raiz, domain.CarpetaEstado, ejecucionID+domain.SufijoEstado)
	stTomando := st + domain.SufijoTomando
	if err := os.Rename(st, stTomando); err != nil {
		t.Fatalf("el take del consolidador sobre el .st falló: %v", err)
	}
	cmdTomando := filepath.Join(raiz, domain.CarpetaInbox, maquina, solicitudID+domain.SufijoComando+domain.SufijoTomando)
	for _, origen := range []string{cmdTomando, stTomando} {
		if err := os.Rename(origen, filepath.Join(destino, filepath.Base(origen))); err != nil {
			t.Fatalf("no se pudo mover %s a procesados: %v", origen, err)
		}
	}
}

func TestCicloContratoHU3(t *testing.T) {
	raiz := t.TempDir()
	emisor := NuevoEmisor(raiz)
	lector := snapshotlector.NuevoLector(raiz, domain.SnapshotDefecto)

	// Lado TUI (HU1): publicación del comando en inbox/<maquina>/.
	sol := testSolicitud(t, nuevoBot(t, "facturacion"))
	if err := emisor.SolicitarEjecucion(context.Background(), sol); err != nil {
		t.Fatalf("la publicación del comando falló: %v", err)
	}
	cmd := filepath.Join(raiz, domain.CarpetaInbox, sol.Maquina, sol.ID+domain.SufijoComando)

	// Lado agente (AC1): take exactamente una vez — .cmd → .cmd.tomando.
	tomando := cmd + domain.SufijoTomando
	if err := os.Rename(cmd, tomando); err != nil {
		t.Fatalf("el take del agente falló: %v", err)
	}
	if _, err := os.Stat(cmd); !os.IsNotExist(err) {
		t.Fatal("tras el take no debe existir el .cmd (ningún agente puede tomarlo dos veces)")
	}
	datosCmd, err := os.ReadFile(tomando)
	if err != nil {
		t.Fatal(err)
	}
	var comando cmdJSON
	if err := json.Unmarshal(datosCmd, &comando); err != nil {
		t.Fatalf("el comando no es JSON válido: %v", err)
	}
	if comando.ID != sol.ID || comando.Maquina != sol.Maquina ||
		comando.BotNombre != "facturacion" || comando.BotAmbiente != "prod" ||
		comando.Usuario != sol.Usuario.WindowsCompleto ||
		comando.VersionFormato != domain.VersionFormatoActual {
		t.Fatalf("el comando no refleja la solicitud: %+v", comando)
	}

	// Lado agente (AC1): reporta EN_PROCESO como .st atómico en estado/.
	gen := identificadores.Nuevo()
	ejecucionID, err := gen.NuevoID()
	if err != nil {
		t.Fatal(err)
	}
	inicio := "2026-09-23T10:00:01Z"
	publicarST(t, raiz, ejecucionID, sol.ID, domain.EstadoEnProceso, inicio, nil, nil, nil)

	// Lado consolidador: drena y publica el snapshot (AC3).
	publicarArchivo(t, filepath.Join(raiz, domain.SnapshotDefecto),
		marshalarSnapshot(t, comando, ejecucionID, inicio, domain.EstadoEnProceso, domain.EstadoEnProceso, nil, nil, nil))

	// Lado TUI (HU2): el LectorSnapshot refleja el estado en curso.
	snap, err := lector.Leer(context.Background())
	if err != nil {
		t.Fatalf("el snapshot no es legible por el TUI: %v", err)
	}
	if len(snap.Solicitudes) != 1 || len(snap.Ejecuciones) != 1 {
		t.Fatalf("el snapshot debe reflejar 1 solicitud y 1 ejecución: %+v", snap)
	}
	s := snap.Solicitudes[0]
	if s.ID != sol.ID || s.Maquina != sol.Maquina || s.Usuario != sol.Usuario.WindowsCompleto ||
		s.BotNombre != "facturacion" || s.Estado != domain.EstadoEnProceso {
		t.Fatalf("la solicitud del snapshot no coincide: %+v", s)
	}
	if want := sol.Fecha.UTC().Truncate(time.Second); !s.Fecha.Equal(want) {
		t.Errorf("fecha redondeada por el protocolo: %v ≠ %v", s.Fecha, want)
	}
	e := snap.Ejecuciones[0]
	if e.ID != ejecucionID || e.SolicitudID != sol.ID || e.Estado != domain.EstadoEnProceso {
		t.Fatalf("la ejecución del snapshot no coincide: %+v", e)
	}
	if e.Inicio == nil || !e.Inicio.Equal(time.Date(2026, 9, 23, 10, 0, 1, 0, time.UTC)) {
		t.Errorf("inicio de la ejecución no mapeado: %+v", e.Inicio)
	}

	// Lado agente (AC2): publica el resultado final con código.
	fin := "2026-09-23T10:02:30Z"
	cero := 0
	publicarST(t, raiz, ejecucionID, sol.ID, domain.EstadoExitosa, inicio, &fin, &cero, nil)

	// Lado consolidador: actualiza el snapshot y drena los archivos a procesados/.
	publicarArchivo(t, filepath.Join(raiz, domain.SnapshotDefecto),
		marshalarSnapshot(t, comando, ejecucionID, inicio, domain.EstadoExitosa, domain.EstadoExitosa, &fin, &cero, nil))
	drenar(t, raiz, sol.Maquina, sol.ID, ejecucionID)

	// Lado TUI: estado final consultable y carpetas drenadas.
	snap2, err := lector.Leer(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	e2 := snap2.Ejecuciones[0]
	if e2.Estado != domain.EstadoExitosa || !e2.Estado.EsFinal() {
		t.Errorf("la ejecución debe ser final EXITOSA: %+v", e2)
	}
	if e2.Codigo == nil || *e2.Codigo != 0 {
		t.Errorf("código de salida = %+v", e2.Codigo)
	}
	if e2.Fin == nil {
		t.Error("el fin de la ejecución debe estar presente")
	}
	entradasInbox, err := os.ReadDir(filepath.Join(raiz, domain.CarpetaInbox, sol.Maquina))
	if err != nil || len(entradasInbox) != 0 {
		t.Errorf("el inbox debe quedar drenado (err=%v, quedan=%d)", err, len(entradasInbox))
	}
	entradasEstado, err := os.ReadDir(filepath.Join(raiz, domain.CarpetaEstado))
	if err != nil || len(entradasEstado) != 0 {
		t.Errorf("estado/ debe quedar drenado (err=%v, quedan=%d)", err, len(entradasEstado))
	}
}