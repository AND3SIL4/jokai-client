package filequeue

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"jokai-client/internal/adapters/driven/identificadores"
	"jokai-client/internal/domain"
)

// T021 (HU4 AC1): N publicaciones simultáneas contra la misma raíz producen N
// comandos .cmd con UUIDs distintos, ninguno pisado, todos legibles y válidos.
func TestEmisorConcurrencia(t *testing.T) {
	raiz := t.TempDir()
	e := NuevoEmisor(raiz)
	gen := identificadores.Nuevo()
	bot := nuevoBot(t, "facturacion")
	usuario := domain.Usuario{WindowsCompleto: `JOKAI\ana`, Rol: domain.RolNegocio}

	const N = 20
	errs := make(chan error, N)
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := gen.NuevoID()
			if err != nil {
				errs <- err
				return
			}
			sol, err := domain.NuevaSolicitud(id, bot, usuario, "BOTPC01", time.Now().UTC())
			if err != nil {
				errs <- err
				return
			}
			if err := e.SolicitarEjecucion(context.Background(), sol); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("una publicación concurrente falló: %v", err)
	}

	carpeta := filepath.Join(raiz, domain.CarpetaInbox, "BOTPC01")
	entradas, err := os.ReadDir(carpeta)
	if err != nil {
		t.Fatal(err)
	}
	vistos := make(map[string]bool, N)
	cmds := 0
	for _, en := range entradas {
		if filepath.Ext(en.Name()) != domain.SufijoComando {
			continue
		}
		cmds++
		base := strings.TrimSuffix(en.Name(), domain.SufijoComando)
		if !domain.EsUUIDv4(base) {
			t.Errorf("el nombre del comando no es un UUID v4: %s", en.Name())
			continue
		}
		if vistos[base] {
			t.Errorf("dos publicaciones pisaron el mismo UUID: %s", base)
			continue
		}
		vistos[base] = true

		datos, err := os.ReadFile(filepath.Join(carpeta, en.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var c struct {
			VersionFormato int    `json:"version_formato"`
			ID             string `json:"id"`
		}
		if err := json.Unmarshal(datos, &c); err != nil {
			t.Errorf("el comando %s no es JSON legible: %v", en.Name(), err)
			continue
		}
		if c.ID != base {
			t.Errorf("el id interno (%s) no coincide con el nombre (%s)", c.ID, base)
		}
		if c.VersionFormato != domain.VersionFormatoActual {
			t.Errorf("version_formato = %d en %s", c.VersionFormato, en.Name())
		}
	}
	if cmds != N {
		t.Errorf("se esperaban %d comandos .cmd, hay %d", N, cmds)
	}
}

// T022 (HU4 AC2): un .cmd.tmp incompleto NO es publicación (solo cuenta el
// .cmd final) y el emisor publica con O_CREATE|O_EXCL sobre el .tmp + rename.
func TestEmisorAtomicidadTmp(t *testing.T) {
	raiz := t.TempDir()
	carpeta := filepath.Join(raiz, domain.CarpetaInbox, "BOTPC01")
	if err := os.MkdirAll(carpeta, 0o755); err != nil {
		t.Fatal(err)
	}

	// .tmp a medio escribir de una publicación interrumpida: no es publicación.
	tmpHuerfano := filepath.Join(carpeta, "550e8400-e29b-41d4-a716-446655440099"+domain.SufijoComando+domain.SufijoTemp)
	if err := os.WriteFile(tmpHuerfano, []byte(`{"incompleto`), 0o644); err != nil {
		t.Fatal(err)
	}
	entradas, _ := os.ReadDir(carpeta)
	for _, en := range entradas {
		if strings.HasSuffix(en.Name(), domain.SufijoComando) {
			t.Fatalf("un .tmp nunca debe exponerse como publicación: %s", en.Name())
		}
	}

	e := NuevoEmisor(raiz)
	bot := nuevoBot(t, "facturacion")
	id := "550e8400-e29b-41d4-a716-446655440100"
	sol := testSolicitudConID(t, bot, id)
	if err := e.SolicitarEjecucion(context.Background(), sol); err != nil {
		t.Fatalf("la publicación falló: %v", err)
	}

	// El .tmp propio desapareció con el rename; el huérfano queda intacto.
	if _, err := os.Stat(filepath.Join(carpeta, id+domain.SufijoComando+domain.SufijoTemp)); !os.IsNotExist(err) {
		t.Error("el .tmp propio debe desaparecer tras el rename")
	}
	datosHuerfano, err := os.ReadFile(tmpHuerfano)
	if err != nil || string(datosHuerfano) != `{"incompleto` {
		t.Error("el .tmp huérfano debe conservarse intacto")
	}
}

// T022 (HU4): con O_CREATE|O_EXCL, un .tmp preexistente para el mismo UUID
// hace fallar la publicación sin pisar el archivo previo.
func TestEmisorTmpExistenteNoSePisa(t *testing.T) {
	raiz := t.TempDir()
	carpeta := filepath.Join(raiz, domain.CarpetaInbox, "BOTPC01")
	if err := os.MkdirAll(carpeta, 0o755); err != nil {
		t.Fatal(err)
	}
	id := "550e8400-e29b-41d4-a716-446655440111"
	tmp := filepath.Join(carpeta, id+domain.SufijoComando+domain.SufijoTemp)
	basura := `{"publicacion a medias"`
	if err := os.WriteFile(tmp, []byte(basura), 0o644); err != nil {
		t.Fatal(err)
	}

	e := NuevoEmisor(raiz)
	sol := testSolicitudConID(t, nuevoBot(t, "facturacion"), id)
	err := e.SolicitarEjecucion(context.Background(), sol)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario por el tmp tomado (O_EXCL), se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "comando temporal") {
		t.Errorf("la causa debe nombrar el tmp: %q", eu.Causa)
	}

	conservado, errLectura := os.ReadFile(tmp)
	if errLectura != nil || string(conservado) != basura {
		t.Error("el .tmp preexistente no debe pisarse")
	}
	if _, err := os.Stat(filepath.Join(carpeta, id+domain.SufijoComando)); !os.IsNotExist(err) {
		t.Error("la publicación fallida no debe dejar un .cmd final")
	}
}

// T023 (HU4): ningún camino pisa un UUID existente ni abre/trunca el .cmd final.
func TestEmisorNoPisaCmdExistente(t *testing.T) {
	raiz := t.TempDir()
	e := NuevoEmisor(raiz)
	bot := nuevoBot(t, "facturacion")
	sol := testSolicitud(t, bot)

	if err := e.SolicitarEjecucion(context.Background(), sol); err != nil {
		t.Fatal(err)
	}
	cmd := filepath.Join(raiz, domain.CarpetaInbox, sol.Maquina, sol.ID+domain.SufijoComando)
	antes, err := os.ReadFile(cmd)
	if err != nil {
		t.Fatal(err)
	}

	err = e.SolicitarEjecucion(context.Background(), sol)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("republicar el mismo UUID debe fallar con ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "ya existe") {
		t.Errorf("la causa debe indicar el comando existente: %q", eu.Causa)
	}
	if !strings.Contains(eu.Sugerencia, "reintente") {
		t.Errorf("la sugerencia debe indicar reintento: %q", eu.Sugerencia)
	}

	despues, err := os.ReadFile(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if string(antes) != string(despues) {
		t.Error("el .cmd final jamás se abre/trunca: el contenido cambió")
	}
	if len(despues) == 0 {
		t.Error("el .cmd final no debe quedar vacío")
	}
}