package domain

import (
	"errors"
	"testing"
	"time"
)

func TestEsUUIDv4(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"550e8400-e29b-51d4-a716-446655440000", false}, // versión 5
		{"550e8400e29b41d4a716446655440000", false},     // sin guiones
		{"550e8400-e29b-41d4-a716-44665544000", false},  // corto
		{"550e8400-e29b-41d4-a716-44665544000g", false}, // no hex
		{"", false},
	}
	for _, c := range casos {
		if got := EsUUIDv4(c.entrada); got != c.esperado {
			t.Errorf("EsUUIDv4(%q) = %v, se esperaba %v", c.entrada, got, c.esperado)
		}
	}
}

func TestEstadoEjecucion(t *testing.T) {
	if !EstadoSolicitada.EsValido() || !EstadoEnProceso.EsValido() ||
		!EstadoExitosa.EsValido() || !EstadoFallida.EsValido() || !EstadoCancelada.EsValido() {
		t.Error("todos los estados definidos deben ser válidos")
	}
	if EstadoEjecucion("X").EsValido() {
		t.Error("un estado desconocido no debe ser válido")
	}
	if !EstadoExitosa.EsFinal() || !EstadoFallida.EsFinal() || !EstadoCancelada.EsFinal() {
		t.Error("EXITOSA, FALLIDA y CANCELADA deben ser finales")
	}
	if EstadoSolicitada.EsFinal() || EstadoEnProceso.EsFinal() {
		t.Error("SOLICITADA y EN_PROCESO no deben ser finales")
	}
	if EstadoEnProceso.Etiqueta() != "En proceso" || EstadoFallida.Etiqueta() != "Fallida" {
		t.Error("las etiquetas de estado deben estar en español")
	}
}

func TestNuevaSolicitud(t *testing.T) {
	bot, err := NuevoBot("facturacion", AmbienteProd(), `C:\bots\facturacion`, `C:\bots\facturacion\main.py`)
	if err != nil {
		t.Fatal(err)
	}
	usuario := Usuario{WindowsCompleto: `JOKAI\ana`, Rol: RolNegocio}
	base := func() Solicitud {
		s, err := NuevaSolicitud("550e8400-e29b-41d4-a716-446655440000", bot, usuario, "BOTPC01", time.Now().UTC())
		if err != nil {
			t.Fatal(err)
		}
		return s
	}

	if s := base(); s.VersionFormato != VersionFormatoActual {
		t.Errorf("VersionFormato = %d, se esperaba %d", s.VersionFormato, VersionFormatoActual)
	}

	if _, err := NuevaSolicitud("no-es-uuid", bot, usuario, "BOTPC01", time.Now()); err == nil {
		t.Error("id no UUID debe fallar")
	}
	if _, err := NuevaSolicitud("550e8400-e29b-41d4-a716-446655440000", Bot{}, usuario, "BOTPC01", time.Now()); err == nil {
		t.Error("bot inválido debe fallar")
	}
	if _, err := NuevaSolicitud("550e8400-e29b-41d4-a716-446655440000", bot, Usuario{WindowsCompleto: " "}, "BOTPC01", time.Now()); err == nil {
		t.Error("usuario vacío debe fallar")
	}
	if _, err := NuevaSolicitud("550e8400-e29b-41d4-a716-446655440000", bot, usuario, " ", time.Now()); err == nil {
		t.Error("máquina vacía debe fallar")
	}

	var eu *ErrorUsuario
	_, err = NuevaSolicitud("no-es-uuid", bot, usuario, "BOTPC01", time.Now())
	if !errors.As(err, &eu) || eu.Causa == "" || eu.Sugerencia == "" {
		t.Error("el error debe ser ErrorUsuario con Causa y Sugerencia")
	}
}