package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"jokai-client/internal/domain"
)

func escribirConfig(t *testing.T, contenido string) string {
	t.Helper()
	ruta := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(ruta, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
	return ruta
}

const configValida = `
[fileserver]
raiz = '\\FILESERVER\jokai'
db = 'jokai.db'

[descubrimiento]
archivo_entrada = 'main.py'
profundidad_maxima = 5

[ambientes.dev]
ruta_bots = '\\FILESERVER\bots-dev'

[ambientes.prod]
ruta_bots = '\\FILESERVER\bots-prod'

[ui]
intervalo_refresco = '2s'
`

func TestCargarConfigValida(t *testing.T) {
	ruta := escribirConfig(t, configValida)
	cfg, err := Cargar(ruta)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if cfg.RaizFileserver != `\\FILESERVER\jokai` {
		t.Errorf("RaizFileserver = %q", cfg.RaizFileserver)
	}
	if cfg.RutaBD != `\\FILESERVER\jokai\jokai.db` {
		t.Errorf("RutaBD = %q", cfg.RutaBD)
	}
	if cfg.Descubrimiento.ArchivoEntrada != "main.py" {
		t.Errorf("ArchivoEntrada = %q", cfg.Descubrimiento.ArchivoEntrada)
	}
	if cfg.Descubrimiento.ProfundidadMaxima != 5 {
		t.Errorf("ProfundidadMaxima = %d", cfg.Descubrimiento.ProfundidadMaxima)
	}
	if cfg.RutaBots(domain.AmbienteDev()) != `\\FILESERVER\bots-dev` {
		t.Error("RutaBots(dev) incorrecta")
	}
	if cfg.RutaBots(domain.AmbienteProd()) != `\\FILESERVER\bots-prod` {
		t.Error("RutaBots(prod) incorrecta")
	}
	if cfg.IntervaloRefresco != 2*time.Second {
		t.Errorf("IntervaloRefresco = %v", cfg.IntervaloRefresco)
	}
}

func TestCargarAplicaDefectos(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = '\\FS\jokai'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'

[ambientes.prod]
ruta_bots = '\\FS\bots-prod'
`)
	cfg, err := Cargar(ruta)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if cfg.Descubrimiento.ArchivoEntrada != domain.ArchivoEntradaDefecto {
		t.Errorf("ArchivoEntrada por defecto = %q", cfg.Descubrimiento.ArchivoEntrada)
	}
	if cfg.Descubrimiento.ProfundidadMaxima != domain.ProfundidadMaximaDefecto {
		t.Errorf("ProfundidadMaxima por defecto = %d", cfg.Descubrimiento.ProfundidadMaxima)
	}
	if cfg.IntervaloRefresco != domain.IntervaloRefrescoDefecto {
		t.Errorf("IntervaloRefresco por defecto = %v", cfg.IntervaloRefresco)
	}
	if cfg.RutaBD == "" {
		t.Error("RutaBD debe tener un valor por defecto (jokai.db)")
	}
}

func TestCargarArchivoInexistente(t *testing.T) {
	_, err := Cargar(filepath.Join(t.TempDir(), "no-existe.toml"))
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if eu.Causa == "" || eu.Sugerencia == "" {
		t.Error("el error debe indicar la ruta buscada y cómo crear el archivo")
	}
}

func TestCargarTOMLInvalido(t *testing.T) {
	ruta := escribirConfig(t, "[fileserver\nraiz = 'x'")
	_, err := Cargar(ruta)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
}

func TestCargarCampoRelativo(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = 'carpeta-relativa'
db = 'jokai.db'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'

[ambientes.prod]
ruta_bots = '\\FS\bots-prod'
`)
	_, err := Cargar(ruta)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "fileserver.raiz") {
		t.Errorf("la causa debe nombrar el campo 'fileserver.raiz': %q", eu.Causa)
	}
}

func TestCargarFaltaAmbiente(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = '\\FS\jokai'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'
`)
	_, err := Cargar(ruta)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "ambientes.prod") {
		t.Errorf("la causa debe nombrar el campo faltante 'ambientes.prod': %q", eu.Causa)
	}
}

func TestCargarProtocoloValido(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = '\\FS\jokai'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'

[ambientes.prod]
ruta_bots = '\\FS\bots-prod'

[protocolo]
snapshot               = 'snap.json'
timeout_ejecucion      = '2h'
latido_agente          = '30s'
umbral_agente_inactivo = '3m'
retencion_archivos     = '24h'
retencion_historial    = '48h'
`)
	cfg, err := Cargar(ruta)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if cfg.Protocolo.Snapshot != "snap.json" {
		t.Errorf("Snapshot = %q", cfg.Protocolo.Snapshot)
	}
	if cfg.Protocolo.TimeoutEjecucion != 2*time.Hour {
		t.Errorf("TimeoutEjecucion = %v", cfg.Protocolo.TimeoutEjecucion)
	}
	if cfg.Protocolo.LatidoAgente != 30*time.Second {
		t.Errorf("LatidoAgente = %v", cfg.Protocolo.LatidoAgente)
	}
	if cfg.Protocolo.UmbralAgenteInactivo != 3*time.Minute {
		t.Errorf("UmbralAgenteInactivo = %v", cfg.Protocolo.UmbralAgenteInactivo)
	}
	if cfg.Protocolo.RetencionArchivos != 24*time.Hour {
		t.Errorf("RetencionArchivos = %v", cfg.Protocolo.RetencionArchivos)
	}
	if cfg.Protocolo.RetencionHistorial != 48*time.Hour {
		t.Errorf("RetencionHistorial = %v", cfg.Protocolo.RetencionHistorial)
	}
}

func TestCargarProtocoloDefectos(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = '\\FS\jokai'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'

[ambientes.prod]
ruta_bots = '\\FS\bots-prod'
`)
	cfg, err := Cargar(ruta)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if cfg.Protocolo.Snapshot != domain.SnapshotDefecto {
		t.Errorf("Snapshot por defecto = %q", cfg.Protocolo.Snapshot)
	}
	if cfg.Protocolo.TimeoutEjecucion != domain.TimeoutEjecucionDefecto {
		t.Errorf("TimeoutEjecucion por defecto = %v", cfg.Protocolo.TimeoutEjecucion)
	}
	if cfg.Protocolo.LatidoAgente != domain.LatidoAgenteDefecto {
		t.Errorf("LatidoAgente por defecto = %v", cfg.Protocolo.LatidoAgente)
	}
	if cfg.Protocolo.UmbralAgenteInactivo != domain.UmbralAgenteInactivoDefecto {
		t.Errorf("UmbralAgenteInactivo por defecto = %v", cfg.Protocolo.UmbralAgenteInactivo)
	}
}

func TestCargarProtocoloSnapshotConRuta(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = '\\FS\jokai'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'

[ambientes.prod]
ruta_bots = '\\FS\bots-prod'

[protocolo]
snapshot = 'sub\estado.json'
`)
	_, err := Cargar(ruta)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "protocolo.snapshot") {
		t.Errorf("la causa debe nombrar el campo: %q", eu.Causa)
	}
}

func TestCargarProtocoloDuracionInvalida(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = '\\FS\jokai'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'

[ambientes.prod]
ruta_bots = '\\FS\bots-prod'

[protocolo]
timeout_ejecucion = 'rapido'
`)
	_, err := Cargar(ruta)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "protocolo.timeout_ejecucion") {
		t.Errorf("la causa debe nombrar el campo: %q", eu.Causa)
	}
}

func TestCargarProtocoloDuracionCero(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = '\\FS\jokai'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'

[ambientes.prod]
ruta_bots = '\\FS\bots-prod'

[protocolo]
latido_agente = '0s'
`)
	_, err := Cargar(ruta)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "duraciones positivas") {
		t.Errorf("la causa debe indicar duraciones positivas: %q", eu.Causa)
	}
}

func TestCargarIntervaloInvalido(t *testing.T) {
	ruta := escribirConfig(t, `
[fileserver]
raiz = '\\FS\jokai'

[ambientes.dev]
ruta_bots = '\\FS\bots-dev'

[ambientes.prod]
ruta_bots = '\\FS\bots-prod'

[ui]
intervalo_refresco = 'lento'
`)
	_, err := Cargar(ruta)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "intervalo_refresco") {
		t.Errorf("la causa debe nombrar el campo: %q", eu.Causa)
	}
}
