package domain

import "testing"

func TestParseAmbiente(t *testing.T) {
	casos := []struct {
		entrada    string
		esperado   Ambiente
		conError   bool
	}{
		{entrada: "dev", esperado: AmbienteDev()},
		{entrada: "prod", esperado: AmbienteProd()},
		{entrada: "DEV", conError: true},
		{entrada: "", conError: true},
		{entrada: "qa", conError: true},
	}
	for _, c := range casos {
		obtenido, err := ParseAmbiente(c.entrada)
		if c.conError {
			if err == nil {
				t.Errorf("ParseAmbiente(%q): se esperaba error, no hubo", c.entrada)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseAmbiente(%q): error inesperado: %v", c.entrada, err)
			continue
		}
		if obtenido != c.esperado {
			t.Errorf("ParseAmbiente(%q) = %q, se esperaba %q", c.entrada, obtenido, c.esperado)
		}
	}
}

func TestAmbienteEtiqueta(t *testing.T) {
	if AmbienteDev().Etiqueta() != "desarrollo" {
		t.Error("la etiqueta de dev debe ser 'desarrollo'")
	}
	if AmbienteProd().Etiqueta() != "producción" {
		t.Error("la etiqueta de prod debe ser 'producción'")
	}
	if (Ambiente{}).Etiqueta() != "desconocido" {
		t.Error("la etiqueta del ambiente vacío debe ser 'desconocido'")
	}
}

func TestAmbienteEsValido(t *testing.T) {
	if !AmbienteDev().EsValido() || !AmbienteProd().EsValido() {
		t.Error("dev y prod deben ser válidos")
	}
	if (Ambiente{}).EsValido() {
		t.Error("el ambiente vacío no debe ser válido")
	}
}

func TestAmbienteAlternativo(t *testing.T) {
	if AmbienteDev().Alternativo() != AmbienteProd() {
		t.Error("el alternativo de dev debe ser prod")
	}
	if AmbienteProd().Alternativo() != AmbienteDev() {
		t.Error("el alternativo de prod debe ser dev")
	}
}
