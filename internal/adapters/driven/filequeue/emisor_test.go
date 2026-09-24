package filequeue

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"jokai-client/internal/domain"
)

func nuevoBot(t *testing.T, nombre string) domain.Bot {
	t.Helper()
	bot, err := domain.NuevoBot(nombre, domain.AmbienteProd(), `C:\bots\`+nombre, `C:\bots\`+nombre+`\main.py`)
	if err != nil {
		t.Fatal(err)
	}
	return bot
}

func testSolicitud(t *testing.T, bot domain.Bot) domain.Solicitud {
	t.Helper()
	return testSolicitudConID(t, bot, "550e8400-e29b-41d4-a716-446655440000")
}

func testSolicitudConID(t *testing.T, bot domain.Bot, id string) domain.Solicitud {
	t.Helper()
	sol, err := domain.NuevaSolicitud(
		id,
		bot,
		domain.Usuario{WindowsCompleto: `JOKAI\ana`, Rol: domain.RolNegocio},
		"BOTPC01",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatal(err)
	}
	return sol
}

func TestEmisorPublicaComando(t *testing.T) {
	raiz := t.TempDir()
	e := NuevoEmisor(raiz)
	sol := testSolicitud(t, nuevoBot(t, "facturacion"))

	if err := e.SolicitarEjecucion(context.Background(), sol); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	carpeta := filepath.Join(raiz, domain.CarpetaInbox, "BOTPC01")
	entries, err := os.ReadDir(carpeta)
	if err != nil {
		t.Fatalf("no se creó el inbox de la máquina: %v", err)
	}
	var vistos []string
	for _, en := range entries {
		vistos = append(vistos, en.Name())
		if en.IsDir() || filepath.Ext(en.Name()) != domain.SufijoComando {
			continue
		}
		if !domain.EsUUIDv4(en.Name()[0 : len(en.Name())-len(domain.SufijoComando)]) {
			t.Errorf("el comando no tiene nombre UUID v4: %s", en.Name())
		}
		datos, err := os.ReadFile(filepath.Join(carpeta, en.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var comando struct {
			VersionFormato  int    `json:"version_formato"`
			ID              string `json:"id"`
			BotNombre       string `json:"bot_nombre"`
			BotAmbiente     string `json:"bot_ambiente"`
			BotRutaEntrada  string `json:"bot_ruta_entrada"`
			Maquina         string `json:"maquina"`
			Usuario         string `json:"usuario"`
			Fecha           string `json:"fecha"`
		}
		if err := json.Unmarshal(datos, &comando); err != nil {
			t.Fatalf("el comando no es JSON válido: %v", err)
		}
		if comando.VersionFormato != domain.VersionFormatoActual {
			t.Errorf("version_formato = %d", comando.VersionFormato)
		}
		if comando.BotNombre != "facturacion" || comando.BotAmbiente != "prod" {
			t.Errorf("datos del bot incorrectos: %+v", comando)
		}
		if comando.Maquina != "BOTPC01" || comando.Usuario != `JOKAI\ana` {
			t.Errorf("máquina/usuario incorrectos: %+v", comando)
		}
	}
	if len(vistos) != 1 {
		t.Fatalf("se esperaba 1 comando, hay: %v", vistos)
	}
}

func TestEmisorIgnoraTemp(t *testing.T) {
	raiz := t.TempDir()
	e := NuevoEmisor(raiz)
	bot := nuevoBot(t, "facturacion")
	sol1 := testSolicitud(t, bot)
	sol2 := testSolicitudConID(t, bot, "550e8400-e29b-41d4-a716-446655440222")

	if err := e.SolicitarEjecucion(context.Background(), sol1); err != nil {
		t.Fatal(err)
	}
	carpeta := filepath.Join(raiz, domain.CarpetaInbox, "BOTPC01")
	// simular un .tmp a medio escribir
	tmp := filepath.Join(carpeta, "550e8400-e29b-41d4-a716-446655440111"+domain.SufijoComando+domain.SufijoTemp)
	if err := os.WriteFile(tmp, []byte(`{"incompleto"`), 0o644); err != nil {
		t.Fatal(err)
	}

	// el emisor debe publicar la segunda solicitud sin tocar el .tmp ni el .cmd previo
	if err := e.SolicitarEjecucion(context.Background(), sol2); err != nil {
		t.Fatal(err)
	}
	// debe haber 2 comandos .cmd finales y el .tmp intacto
	entries, _ := os.ReadDir(carpeta)
	cmds := 0
	tmps := 0
	tmpIntacto := false
	for _, en := range entries {
		switch filepath.Ext(en.Name()) {
		case domain.SufijoComando:
			cmds++
		case domain.SufijoTemp:
			tmps++
			if en.Name() == filepath.Base(tmp) {
				tmpIntacto = true
			}
		}
	}
	if cmds != 2 {
		t.Errorf("debe haber 2 comandos .cmd, hay %d", cmds)
	}
	if !tmpIntacto {
		t.Errorf("el .tmp a medio escribir debe conservarse intacto, hay %d .tmp", tmps)
	}
}

func TestEmisorRaizInexistente(t *testing.T) {
	e := NuevoEmisor(`C:\ruta\que\no\existe`)
	sol := testSolicitud(t, nuevoBot(t, "facturacion"))

	err := e.SolicitarEjecucion(context.Background(), sol)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if eu.Causa == "" || eu.Sugerencia == "" {
		t.Error("el error debe llevar causa y sugerencia")
	}
}