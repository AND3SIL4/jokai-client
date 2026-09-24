package casosdeuso

import (
	"context"
	"errors"
	"strings"
	"testing"

	"jokai-client/internal/domain"
)

func TestVerEjecucionDetalle(t *testing.T) {
	uc := NewVerEjecucion(&snapshotFake{snapshot: snapshotConDatos()})
	detalle, err := uc.Ejecutar(context.Background(), "e2")
	if err != nil {
		t.Fatal(err)
	}
	if detalle.Ejecucion.ID != "e2" || detalle.Ejecucion.Estado != domain.EstadoExitosa {
		t.Errorf("detalle incorrecto: %+v", detalle.Ejecucion)
	}
	if detalle.Solicitud.ID != "a2" {
		t.Errorf("solicitud asociada incorrecta: %+v", detalle.Solicitud)
	}
	if detalle.Ejecucion.Codigo == nil || *detalle.Ejecucion.Codigo != 0 {
		t.Errorf("código de salida = %v", detalle.Ejecucion.Codigo)
	}
}

func TestVerEjecucionInexistente(t *testing.T) {
	uc := NewVerEjecucion(&snapshotFake{snapshot: snapshotConDatos()})
	_, err := uc.Ejecutar(context.Background(), "no-existe")
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "no se encontró") {
		t.Errorf("causa = %q", eu.Causa)
	}
	if eu.Sugerencia == "" {
		t.Error("debe haber sugerencia")
	}
}

func TestVerEjecucionSnapshotAusente(t *testing.T) {
	uc := NewVerEjecucion(&snapshotFake{err: domain.NuevoErrorUsuario("sin snapshot", "espere")})
	_, err := uc.Ejecutar(context.Background(), "e2")
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
}

func TestVerEjecucionSinSolicitud(t *testing.T) {
	snap := snapshotConDatos()
	snap.Solicitudes = snap.Solicitudes[:1] // quitar a2
	uc := NewVerEjecucion(&snapshotFake{snapshot: snap})
	_, err := uc.Ejecutar(context.Background(), "e2")
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "solicitud") {
		t.Errorf("la causa debe mencionar la solicitud: %q", eu.Causa)
	}
}