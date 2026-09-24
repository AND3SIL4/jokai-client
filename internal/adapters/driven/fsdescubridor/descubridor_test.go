package fsdescubridor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"jokai-client/internal/domain"
)

func crearArbol(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	touch := func(ruta ...string) {
		t.Helper()
		completa := filepath.Join(append([]string{raiz}, ruta...)...)
		if err := os.MkdirAll(filepath.Dir(completa), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(completa, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkdir := func(ruta ...string) {
		t.Helper()
		completa := filepath.Join(append([]string{raiz}, ruta...)...)
		if err := os.MkdirAll(completa, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	mkdir("dev", "facturacion")
	touch("dev", "facturacion", "main.py")
	touch("dev", "facturacion", "README.md")

	mkdir("dev", "inventario", "src")
	touch("dev", "inventario", "src", "main.py")

	mkdir("dev", "sin-main")
	touch("dev", "sin-main", "otro.py")

	mkdir("dev", "anidado", "a", "b", "c")
	touch("dev", "anidado", "a", "b", "c", "main.py")

	mkdir("prod", "nomina")
	touch("prod", "nomina", "main.py")
	return raiz
}

func nombres(bots []domain.Bot) []string {
	out := make([]string, 0, len(bots))
	for _, b := range bots {
		out = append(out, b.Nombre)
	}
	return out
}

func TestDescubrirBotsBasicos(t *testing.T) {
	raiz := crearArbol(t)
	d := Nuevo(filepath.Join(raiz, "dev"), filepath.Join(raiz, "prod"), "main.py", 5)

	bots, advertencias, err := d.Descubrir(context.Background(), domain.AmbienteDev())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(advertencias) != 0 {
		t.Errorf("advertencias inesperadas: %v", advertencias)
	}
	obtenidos := nombres(bots)
	esperados := []string{"anidado", "facturacion", "inventario"}
	if len(obtenidos) != len(esperados) {
		t.Fatalf("bots = %v, se esperaban exactamente %v", obtenidos, esperados)
	}
	for i, nombre := range esperados {
		if obtenidos[i] != nombre {
			t.Errorf("bots[%d] = %q, se esperaba %q", i, obtenidos[i], nombre)
		}
	}
}

func TestDescubrirBotsSinMainPySeOmite(t *testing.T) {
	raiz := crearArbol(t)
	d := Nuevo(filepath.Join(raiz, "dev"), filepath.Join(raiz, "prod"), "main.py", 5)

	bots, _, err := d.Descubrir(context.Background(), domain.AmbienteDev())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	for _, b := range bots {
		if b.Nombre == "sin-main" {
			t.Error("la carpeta sin main.py no debe aparecer en el catálogo")
		}
	}
}

func TestDescubrirBotsAmbienteProd(t *testing.T) {
	raiz := crearArbol(t)
	d := Nuevo(filepath.Join(raiz, "dev"), filepath.Join(raiz, "prod"), "main.py", 5)

	bots, _, err := d.Descubrir(context.Background(), domain.AmbienteProd())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	obtenidos := nombres(bots)
	if len(obtenidos) != 1 || obtenidos[0] != "nomina" {
		t.Errorf("bots = %v, se esperaba solo 'nomina'", obtenidos)
	}
}

func TestDescubrirBotsMasSuperficialGana(t *testing.T) {
	raiz := t.TempDir()
	amb := filepath.Join(raiz, "dev")
	botDir := filepath.Join(amb, "doble")
	sub := filepath.Join(botDir, "src")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "main.py"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(botDir, "main.py"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	d := Nuevo(amb, amb, "main.py", 5)
	bots, _, err := d.Descubrir(context.Background(), domain.AmbienteDev())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(bots) != 1 || bots[0].Nombre != "doble" {
		t.Fatalf("se esperaba 1 bot 'doble', hubo %v", nombres(bots))
	}
	if bots[0].RutaEntrada != filepath.Join(botDir, "main.py") {
		t.Errorf("RutaEntrada = %q, debía ganar la entrada más superficial", bots[0].RutaEntrada)
	}
}

func TestDescubrirBotsProfundidadMaxima(t *testing.T) {
	raiz := t.TempDir()
	amb := filepath.Join(raiz, "dev")
	botDir := filepath.Join(amb, "profundo")
	if err := os.MkdirAll(filepath.Join(botDir, "a", "b", "c"), 0o755); err != nil {
		t.Fatal(err)
	}
	rutaMain := filepath.Join(botDir, "a", "b", "c", "main.py")
	if err := os.WriteFile(rutaMain, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	d2 := Nuevo(amb, amb, "main.py", 2)
	bots, _, err := d2.Descubrir(context.Background(), domain.AmbienteDev())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(bots) != 0 {
		t.Errorf("con profundidad 2 el main.py del nivel 3 no debe encontrarse, bots = %v", nombres(bots))
	}

	d3 := Nuevo(amb, amb, "main.py", 3)
	bots, _, err = d3.Descubrir(context.Background(), domain.AmbienteDev())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(bots) != 1 || bots[0].RutaEntrada != rutaMain {
		t.Errorf("con profundidad 3 el bot debe encontrarse: %v", nombres(bots))
	}

	dIlimitado := Nuevo(amb, amb, "main.py", 0)
	bots, _, err = dIlimitado.Descubrir(context.Background(), domain.AmbienteDev())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(bots) != 1 {
		t.Error("con profundidad 0 (sin límite) el bot debe encontrarse")
	}
}

func TestDescubrirBotsRutaInexistente(t *testing.T) {
	d := Nuevo(`C:\ruta\que\no\existe`, `C:\otra\no`, "main.py", 5)

	_, _, err := d.Descubrir(context.Background(), domain.AmbienteDev())
	if err == nil {
		t.Fatal("la ruta inexistente debe dar error")
	}
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if eu.Causa == "" || eu.Sugerencia == "" {
		t.Error("el error debe llevar causa y sugerencia")
	}
}

func TestDescubrirBotsAmbienteSinRuta(t *testing.T) {
	d := Nuevo(`C:\x`, `C:\y`, "main.py", 5)
	if _, _, err := d.Descubrir(context.Background(), domain.Ambiente{}); !errors.Is(err, domain.ErrAmbienteInvalido) {
		t.Errorf("se esperaba ErrAmbienteInvalido, se obtuvo: %v", err)
	}
}
