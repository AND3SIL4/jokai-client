package identificadores

import (
	"testing"

	"jokai-client/internal/domain"
)

func TestNuevoIDUUIDv4(t *testing.T) {
	g := Nuevo()
	for i := 0; i < 100; i++ {
		id, err := g.NuevoID()
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if !domain.EsUUIDv4(id) {
			t.Fatalf("el id %q no es un UUID v4 válido", id)
		}
		if id[14] != '4' {
			t.Errorf("el id %q debe tener versión 4 en la posición 14", id)
		}
	}
}

func TestNuevoIDUnicos(t *testing.T) {
	g := Nuevo()
	vistos := map[string]bool{}
	for i := 0; i < 100; i++ {
		id, _ := g.NuevoID()
		if vistos[id] {
			t.Fatalf("id duplicado: %q", id)
		}
		vistos[id] = true
	}
}