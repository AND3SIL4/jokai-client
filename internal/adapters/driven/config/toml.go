package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	toml "github.com/pelletier/go-toml/v2"

	"jokai-client/internal/domain"
)

type crudo struct {
	Fileserver struct {
		Raiz string `toml:"raiz"`
		DB   string `toml:"db"`
	} `toml:"fileserver"`
	Descubrimiento struct {
		ArchivoEntrada    *string `toml:"archivo_entrada"`
		ProfundidadMaxima *int    `toml:"profundidad_maxima"`
	} `toml:"descubrimiento"`
	Ambientes struct {
		Dev struct {
			RutaBots string `toml:"ruta_bots"`
		} `toml:"dev"`
		Prod struct {
			RutaBots string `toml:"ruta_bots"`
		} `toml:"prod"`
	} `toml:"ambientes"`
	UI struct {
		IntervaloRefresco string `toml:"intervalo_refresco"`
	} `toml:"ui"`
	Protocolo crudoProtocolo `toml:"protocolo"`
}

type crudoProtocolo struct {
	Snapshot             *string `toml:"snapshot"`
	TimeoutEjecucion     string  `toml:"timeout_ejecucion"`
	LatidoAgente         string  `toml:"latido_agente"`
	UmbralAgenteInactivo string  `toml:"umbral_agente_inactivo"`
	RetencionArchivos    string  `toml:"retencion_archivos"`
	RetencionHistorial   string  `toml:"retencion_historial"`
}

func Cargar(ruta string) (domain.Configuracion, error) {
	datos, err := os.ReadFile(ruta)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.Configuracion{}, domain.NuevoErrorUsuario(
				fmt.Sprintf("no se encontró el archivo de configuración '%s'", ruta),
				fmt.Sprintf("copie 'config.example.toml' como '%s', ajústelo a sus rutas y vuelva a ejecutar", filepath.Base(ruta)),
			)
		}
		return domain.Configuracion{}, domain.NuevoErrorUsuario(
			fmt.Sprintf("no se pudo leer el archivo de configuración '%s'", ruta),
			"verifique que tenga permisos de lectura sobre el archivo",
		)
	}

	var c crudo
	if err := toml.Unmarshal(datos, &c); err != nil {
		var decErr *toml.DecodeError
		detalle := err.Error()
		if errors.As(err, &decErr) {
			linea, columna := decErr.Position()
			detalle = fmt.Sprintf("error de sintaxis TOML en la línea %d, columna %d", linea, columna)
		}
		return domain.Configuracion{}, domain.NuevoErrorUsuario(
			fmt.Sprintf("el archivo de configuración '%s' no es válido: %s", ruta, detalle),
			"revise la sintaxis del archivo contra config.example.toml",
		)
	}

	archivoEntrada := domain.ArchivoEntradaDefecto
	if c.Descubrimiento.ArchivoEntrada != nil && *c.Descubrimiento.ArchivoEntrada != "" {
		archivoEntrada = *c.Descubrimiento.ArchivoEntrada
	}
	profundidad := domain.ProfundidadMaximaDefecto
	if c.Descubrimiento.ProfundidadMaxima != nil {
		profundidad = *c.Descubrimiento.ProfundidadMaxima
	}
	intervalo := domain.IntervaloRefrescoDefecto
	if c.UI.IntervaloRefresco != "" {
		intervalo, err = time.ParseDuration(c.UI.IntervaloRefresco)
		if err != nil {
			return domain.Configuracion{}, domain.NuevoErrorUsuario(
				"el campo 'ui.intervalo_refresco' no es una duración válida",
				"use valores como '2s' o '500ms'",
			)
		}
	}
	rutaBD := c.Fileserver.DB
	if rutaBD == "" {
		rutaBD = "jokai.db"
	}

	protocolo, err := parsearProtocolo(c.Protocolo)
	if err != nil {
		return domain.Configuracion{}, err
	}

	return domain.NuevaConfiguracion(
		c.Fileserver.Raiz,
		rutaBD,
		domain.DescubrimientoConfig{
			ArchivoEntrada:    archivoEntrada,
			ProfundidadMaxima: profundidad,
		},
		protocolo,
		c.Ambientes.Dev.RutaBots,
		c.Ambientes.Prod.RutaBots,
		intervalo,
	)
}

func parsearDuracioProtocolo(valor string, defecto time.Duration, campo string) (time.Duration, error) {
	if valor == "" {
		return defecto, nil
	}
	duracion, err := time.ParseDuration(valor)
	if err != nil {
		return 0, domain.NuevoErrorUsuario(
			fmt.Sprintf("el campo '%s' no es una duración válida", campo),
			"use valores como '4h', '60s' o '5min'",
		)
	}
	return duracion, nil
}

func parsearProtocolo(c crudoProtocolo) (domain.ProtocoloConfig, error) {
	snapshot := domain.SnapshotDefecto
	if c.Snapshot != nil && *c.Snapshot != "" {
		snapshot = *c.Snapshot
	}
	timeout, err := parsearDuracioProtocolo(c.TimeoutEjecucion, domain.TimeoutEjecucionDefecto, "protocolo.timeout_ejecucion")
	if err != nil {
		return domain.ProtocoloConfig{}, err
	}
	latido, err := parsearDuracioProtocolo(c.LatidoAgente, domain.LatidoAgenteDefecto, "protocolo.latido_agente")
	if err != nil {
		return domain.ProtocoloConfig{}, err
	}
	umbral, err := parsearDuracioProtocolo(c.UmbralAgenteInactivo, domain.UmbralAgenteInactivoDefecto, "protocolo.umbral_agente_inactivo")
	if err != nil {
		return domain.ProtocoloConfig{}, err
	}
	retencionArchivos, err := parsearDuracioProtocolo(c.RetencionArchivos, domain.RetencionArchivosDefecto, "protocolo.retencion_archivos")
	if err != nil {
		return domain.ProtocoloConfig{}, err
	}
	retencionHistorial, err := parsearDuracioProtocolo(c.RetencionHistorial, domain.RetencionHistorialDefecto, "protocolo.retencion_historial")
	if err != nil {
		return domain.ProtocoloConfig{}, err
	}
	cfg := domain.ProtocoloConfig{
		Snapshot:             snapshot,
		TimeoutEjecucion:     timeout,
		LatidoAgente:         latido,
		UmbralAgenteInactivo: umbral,
		RetencionArchivos:    retencionArchivos,
		RetencionHistorial:   retencionHistorial,
	}
	if err := cfg.EsValida(); err != nil {
		return domain.ProtocoloConfig{}, err
	}
	return cfg, nil

}
