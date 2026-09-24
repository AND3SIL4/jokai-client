package domain

import (
	"path/filepath"
	"strings"
	"time"
)

const (
	ArchivoEntradaDefecto    = "main.py"
	ProfundidadMaximaDefecto = 5
	IntervaloRefrescoDefecto  = 2 * time.Second

	SnapshotDefecto            = "estado.json"
	TimeoutEjecucionDefecto    = 4 * time.Hour
	LatidoAgenteDefecto        = 60 * time.Second
	UmbralAgenteInactivoDefecto = 5 * time.Minute
	RetencionArchivosDefecto   = 30 * 24 * time.Hour
	RetencionHistorialDefecto  = 90 * 24 * time.Hour
)

type DescubrimientoConfig struct {
	ArchivoEntrada    string
	ProfundidadMaxima int
}

type ProtocoloConfig struct {
	Snapshot             string
	TimeoutEjecucion     time.Duration
	LatidoAgente         time.Duration
	UmbralAgenteInactivo time.Duration
	RetencionArchivos    time.Duration
	RetencionHistorial   time.Duration
}

type Configuracion struct {
	RaizFileserver    string
	RutaBD            string
	Descubrimiento    DescubrimientoConfig
	Protocolo         ProtocoloConfig
	RutaBotsDev       string
	RutaBotsProd      string
	IntervaloRefresco time.Duration
}

func (c Configuracion) RutaBots(a Ambiente) string {
	if a == AmbienteProd() {
		return c.RutaBotsProd
	}
	return c.RutaBotsDev
}

func NuevaConfiguracion(
	raizFileserver string,
	rutaBDRelativa string,
	descubrimiento DescubrimientoConfig,
	protocolo ProtocoloConfig,
	rutaBotsDev, rutaBotsProd string,
	intervaloRefresco time.Duration,
) (Configuracion, error) {
	if !filepath.IsAbs(raizFileserver) {
		return Configuracion{}, NuevoErrorUsuario(
			"el campo 'fileserver.raiz' debe ser una ruta absoluta (por ejemplo '\\\\servidor\\compartida' o 'C:\\carpeta')",
			"corrija la ruta en config.toml",
		)
	}
	if rutaBDRelativa == "" || filepath.IsAbs(rutaBDRelativa) || strings.Contains(rutaBDRelativa, "..") {
		return Configuracion{}, NuevoErrorUsuario(
			"el campo 'fileserver.db' debe ser una ruta relativa dentro del fileserver, sin '..'",
			"use un valor como 'jokai.db' o 'datos\\jokai.db'",
		)
	}
	if descubrimiento.ArchivoEntrada == "" ||
		filepath.Base(descubrimiento.ArchivoEntrada) != descubrimiento.ArchivoEntrada ||
		strings.ContainsAny(descubrimiento.ArchivoEntrada, `/\`) {
		return Configuracion{}, NuevoErrorUsuario(
			"el campo 'descubrimiento.archivo_entrada' debe ser un nombre de archivo sin rutas",
			"use un valor como 'main.py'",
		)
	}
	if descubrimiento.ProfundidadMaxima < 0 {
		return Configuracion{}, NuevoErrorUsuario(
			"el campo 'descubrimiento.profundidad_maxima' debe ser 0 o más (0 = sin límite)",
			"use 5 como valor típico",
		)
	}
	if !filepath.IsAbs(rutaBotsDev) {
		return Configuracion{}, NuevoErrorUsuario(
			"el campo 'ambientes.dev.ruta_bots' debe ser una ruta absoluta",
			"indique la carpeta de bots de desarrollo en config.toml",
		)
	}
	if !filepath.IsAbs(rutaBotsProd) {
		return Configuracion{}, NuevoErrorUsuario(
			"el campo 'ambientes.prod.ruta_bots' debe ser una ruta absoluta",
			"indique la carpeta de bots de producción en config.toml",
		)
	}
	if intervaloRefresco <= 0 {
		return Configuracion{}, NuevoErrorUsuario(
			"el campo 'ui.intervalo_refresco' debe ser una duración positiva",
			"use valores como '2s' o '500ms'",
		)
	}
	if err := protocolo.EsValida(); err != nil {
		return Configuracion{}, err
	}
	return Configuracion{
		RaizFileserver:    raizFileserver,
		RutaBD:            filepath.Join(raizFileserver, rutaBDRelativa),
		Descubrimiento:    descubrimiento,
		Protocolo:         protocolo,
		RutaBotsDev:       rutaBotsDev,
		RutaBotsProd:      rutaBotsProd,
		IntervaloRefresco: intervaloRefresco,
	}, nil
}

func (p ProtocoloConfig) EsValida() error {
	if p.Snapshot == "" || filepath.Base(p.Snapshot) != p.Snapshot || strings.Contains(p.Snapshot, "/") || strings.Contains(p.Snapshot, `\`) {
		return NuevoErrorUsuario(
			"el campo 'protocolo.snapshot' debe ser un nombre de archivo sin rutas",
			"use un valor como 'estado.json'",
		)
	}
	if p.TimeoutEjecucion <= 0 || p.LatidoAgente <= 0 ||
		p.UmbralAgenteInactivo <= 0 || p.RetencionArchivos <= 0 || p.RetencionHistorial <= 0 {
		return NuevoErrorUsuario(
			"los campos de 'protocolo' (timeout_ejecucion, latido_agente, umbral_agente_inactivo, retencion_archivos, retencion_historial) deben ser duraciones positivas",
			"use valores como '4h', '60s', '5min', '2160h'",
		)
	}
	return nil
}
