package casosdeuso

import (
	"context"
	"errors"
	"testing"
	"time"

	"jokai-client/internal/domain"
)

func snapshotConDatos() domain.Snapshot {
	fecha := time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC)
	return domain.Snapshot{
		VersionFormato: domain.VersionFormatoActual,
		Solicitudes: []domain.SolicitudSnapshot{
			{ID: "a1", BotNombre: "nomina", BotAmbiente: domain.AmbienteProd(), Maquina: "PC2", Usuario: `JOKAI\ana`, Fecha: fecha.Add(2 * time.Hour), Estado: domain.EstadoEnProceso},
			{ID: "a2", BotNombre: "facturacion", BotAmbiente: domain.AmbienteProd(), Maquina: "PC1", Usuario: `JOKAI\ana`, Fecha: fecha.Add(time.Hour), Estado: domain.EstadoExitosa},
			{ID: "a3", BotNombre: "inventario", BotAmbiente: domain.AmbienteDev(), Maquina: "PC1", Usuario: `JOKAI\lui`, Fecha: fecha, Estado: domain.EstadoSolicitada},
		},
		Ejecuciones: []domain.Ejecucion{
			{ID: "e2", SolicitudID: "a2", Estado: domain.EstadoExitosa, Fin: ptrTime(fecha.Add(1*time.Hour + 30*time.Minute)), Codigo: intPtr(0)},
			{ID: "e1", SolicitudID: "a1", Estado: domain.EstadoEnProceso, Inicio: ptrTime(fecha.Add(2 * time.Hour))},
			{ID: "e0", SolicitudID: "a0", Estado: domain.EstadoSolicitada},
		},
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func intPtr(v int) *int { return &v }

func nuevoListar(snapshot *snapshotFake) *ListarEjecuciones {
	return NewListarEjecuciones(snapshot)
}

func TestListarEjecucionesOrdenDescendente(t *testing.T) {
	uc := nuevoListar(&snapshotFake{snapshot: snapshotConDatos()})
	resultado, err := uc.Ejecutar(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(resultado) != 2 {
		t.Fatalf("se esperaban 2 ejecuciones con solicitud, hay %d", len(resultado))
	}
	// orden descendente por fecha de la solicitud
	if resultado[0].Solicitud.ID != "a1" || resultado[1].Solicitud.ID != "a2" {
		t.Errorf("orden incorrecto: %q > %q", resultado[0].Solicitud.ID, resultado[1].Solicitud.ID)
	}
}

func TestListarEjecucionesSoloConSolicitud(t *testing.T) {
	uc := nuevoListar(&snapshotFake{snapshot: snapshotConDatos()})
	resultado, err := uc.Ejecutar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range resultado {
		if r.Ejecucion.SolicitudID == "a0" {
			t.Error("no deben listarse ejecuciones sin solicitud asociada en el snapshot")
		}
	}
}

func TestListarEjecucionesFiltroPorBot(t *testing.T) {
	uc := NewListarEjecuciones(&snapshotFake{snapshot: snapshotConDatos()})
	resultado, err := uc.EjecutarPorBot(context.Background(), "facturacion", domain.AmbienteProd())
	if err != nil {
		t.Fatal(err)
	}
	if len(resultado) != 1 {
		t.Fatalf("se esperaba 1 ejecución de facturacion, hay %d", len(resultado))
	}
	if resultado[0].Solicitud.BotNombre != "facturacion" {
		t.Errorf("bot = %q", resultado[0].Solicitud.BotNombre)
	}
}

func TestListarEjecucionesErrorSnapshot(t *testing.T) {
	uc := nuevoListar(&snapshotFake{err: domain.NuevoErrorUsuario("sin snapshot", "verifique")})
	_, err := uc.Ejecutar(context.Background())
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
}