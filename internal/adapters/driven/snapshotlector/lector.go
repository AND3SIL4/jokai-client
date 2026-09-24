package snapshotlector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"jokai-client/internal/domain"
)

type Lector struct {
	raiz     string
	nombre   string
}

func NuevoLector(raiz, nombre string) *Lector {
	return &Lector{raiz: raiz, nombre: nombre}
}

type snapshotJSON struct {
	VersionFormato int              `json:"version_formato"`
	GeneradoEn     string           `json:"generado_en"`
	Maquinas       []maquinaJSON    `json:"maquinas"`
	Solicitudes    []solicitudJSON  `json:"solicitudes"`
	Ejecuciones    []ejecucionJSON  `json:"ejecuciones"`
}

type maquinaJSON struct {
	Nombre       string   `json:"nombre"`
	Ambiente     string   `json:"ambiente"`
	UltimoLatido *string  `json:"ultimo_latido"`
	Bots         []string `json:"bots"`
}

type solicitudJSON struct {
	ID          string `json:"id"`
	BotNombre   string `json:"bot_nombre"`
	BotAmbiente string `json:"bot_ambiente"`
	Maquina     string `json:"maquina"`
	Usuario     string `json:"usuario"`
	Fecha       string `json:"fecha"`
	Estado      string `json:"estado"`
}

type ejecucionJSON struct {
	ID          string  `json:"id"`
	SolicitudID string  `json:"solicitud_id"`
	Estado      string  `json:"estado"`
	Inicio      *string `json:"inicio"`
	Fin         *string `json:"fin"`
	Codigo      *int    `json:"codigo"`
	Mensaje     *string `json:"mensaje"`
}

func (l *Lector) Leer(ctx context.Context) (domain.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return domain.Snapshot{}, err
	}
	ruta := filepath.Join(l.raiz, l.nombre)
	datos, err := os.ReadFile(ruta)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.Snapshot{}, domain.NuevoErrorUsuario(
				fmt.Sprintf("el consolidador aún no ha publicado el snapshot (%s)", ruta),
				"espere a que el consolidador de jokai-server publique 'estado.json' en la raíz del fileserver",
			)
		}
		return domain.Snapshot{}, domain.NuevoErrorUsuario(
			fmt.Sprintf("no se pudo leer el snapshot '%s'", ruta),
			"verifique los permisos de lectura del fileserver",
		)
	}

	var crudo snapshotJSON
	if err := json.Unmarshal(datos, &crudo); err != nil {
		return domain.Snapshot{}, domain.NuevoErrorUsuario(
			fmt.Sprintf("el snapshot '%s' está malformado", ruta),
			"verifique que jokai-server publique un estado.json JSON válido",
		)
	}

	if crudo.VersionFormato > domain.VersionFormatoActual {
		return domain.Snapshot{}, domain.NuevoErrorUsuario(
			fmt.Sprintf("el snapshot usa una versión de protocolo nueva (%d) que el consolidador soporta, pero este cliente solo soporta hasta la %d", crudo.VersionFormato, domain.VersionFormatoActual),
			"actualice jokai-client a una versión que soporte el protocolo nuevo del consolidador",
		)
	}
	if crudo.VersionFormato < 1 {
		return domain.Snapshot{}, domain.NuevoErrorUsuario(
			"el snapshot declara una versión de protocolo inválida",
			"verifique que jokai-server publique un estado.json correcto",
		)
	}

	snapshot := domain.Snapshot{
		VersionFormato: crudo.VersionFormato,
		GeneradoEn:     timeFromString(crudo.GeneradoEn),
		Maquinas:       make([]domain.MaquinaSnapshot, 0, len(crudo.Maquinas)),
		Solicitudes:    make([]domain.SolicitudSnapshot, 0, len(crudo.Solicitudes)),
		Ejecuciones:    make([]domain.Ejecucion, 0, len(crudo.Ejecuciones)),
	}

	for _, m := range crudo.Maquinas {
		ambiente, err := domain.ParseAmbiente(m.Ambiente)
		if err != nil {
			return domain.Snapshot{}, estadoInvalido("el snapshot declara un ambiente de máquina inválido", m.Ambiente)
		}
		snapshot.Maquinas = append(snapshot.Maquinas, domain.MaquinaSnapshot{
			Nombre:       m.Nombre,
			Ambiente:     ambiente,
			UltimoLatido: ptrTimeFromString(m.UltimoLatido),
			Bots:         m.Bots,
		})
	}

	for _, s := range crudo.Solicitudes {
		ambiente, err := domain.ParseAmbiente(s.BotAmbiente)
		if err != nil {
			return domain.Snapshot{}, estadoInvalido("el snapshot declara un ambiente de bot inválido", s.BotAmbiente)
		}
		estado := domain.EstadoEjecucion(s.Estado)
		if !estado.EsValido() {
			return domain.Snapshot{}, estadoInvalido("el snapshot declara un estado de solicitud inválido", s.Estado)
		}
		snapshot.Solicitudes = append(snapshot.Solicitudes, domain.SolicitudSnapshot{
			ID:          s.ID,
			BotNombre:   s.BotNombre,
			BotAmbiente: ambiente,
			Maquina:     s.Maquina,
			Usuario:     s.Usuario,
			Fecha:       timeFromString(s.Fecha),
			Estado:      estado,
		})
	}

	for _, e := range crudo.Ejecuciones {
		estado := domain.EstadoEjecucion(e.Estado)
		if !estado.EsValido() {
			return domain.Snapshot{}, estadoInvalido("el snapshot declara un estado de ejecución inválido", e.Estado)
		}
		snapshot.Ejecuciones = append(snapshot.Ejecuciones, domain.Ejecucion{
			ID:          e.ID,
			SolicitudID: e.SolicitudID,
			Estado:      estado,
			Inicio:      ptrTimeFromString(e.Inicio),
			Fin:         ptrTimeFromString(e.Fin),
			Codigo:      e.Codigo,
			Mensaje:     e.Mensaje,
		})
	}

	return snapshot, nil
}

func estadoInvalido(causa, valor string) error {
	return domain.NuevoErrorUsuario(
		fmt.Sprintf("%s: %q", causa, valor),
		"verifique que jokai-server publique un estado.json consistente",
	)
}

func timeFromString(valor string) time.Time {
	if valor == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, valor)
	if err != nil {
		return time.Time{}
	}
	return t
}

func ptrTimeFromString(valor *string) *time.Time {
	if valor == nil || *valor == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *valor)
	if err != nil {
		return nil
	}
	return &t
}