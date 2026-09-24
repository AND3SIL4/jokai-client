package domain

import "testing"

func TestNuevoBotValido(t *testing.T) {
	bot, err := NuevoBot("facturacion", AmbienteDev(), `C:\bots\dev\facturacion`, `C:\bots\dev\facturacion\main.py`)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if bot.Nombre != "facturacion" {
		t.Errorf("Nombre = %q, se esperaba 'facturacion'", bot.Nombre)
	}
	if bot.Ambiente != AmbienteDev() {
		t.Errorf("Ambiente = %q, se esperaba dev", bot.Ambiente)
	}
}

func TestNuevoBotConRutaUNC(t *testing.T) {
	_, err := NuevoBot("nomina", AmbienteProd(), `\\fileserver\bots\prod\nomina`, `\\fileserver\bots\prod\nomina\src\main.py`)
	if err != nil {
		t.Fatalf("las rutas UNC deben ser válidas: %v", err)
	}
}

func TestNuevoBotEntradaFueraDeRaiz(t *testing.T) {
	_, err := NuevoBot("x", AmbienteDev(), `C:\bots\dev\x`, `C:\bots\dev\otro\main.py`)
	if err == nil {
		t.Fatal("la entrada fuera de la carpeta del bot debe dar error")
	}
}

func TestNuevoBotRutaRelativa(t *testing.T) {
	if _, err := NuevoBot("x", AmbienteDev(), "bots/dev/x", `C:\bots\dev\x\main.py`); err == nil {
		t.Error("la carpeta relativa debe dar error")
	}
	if _, err := NuevoBot("x", AmbienteDev(), `C:\bots\dev\x`, "main.py"); err == nil {
		t.Error("la entrada relativa debe dar error")
	}
}

func TestNuevoBotNombreVacio(t *testing.T) {
	if _, err := NuevoBot("  ", AmbienteDev(), `C:\bots\dev\x`, `C:\bots\dev\x\main.py`); err == nil {
		t.Error("el nombre vacío debe dar error")
	}
}

func TestNuevoBotAmbienteInvalido(t *testing.T) {
	if _, err := NuevoBot("x", Ambiente{}, `C:\bots\dev\x`, `C:\bots\dev\x\main.py`); err == nil {
		t.Error("el ambiente inválido debe dar error")
	}
}
