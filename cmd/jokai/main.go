package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"jokai-client/internal/adapters/driven/config"
	"jokai-client/internal/adapters/driven/filequeue"
	"jokai-client/internal/adapters/driven/fsdescubridor"
	"jokai-client/internal/adapters/driven/identificadores"
	"jokai-client/internal/adapters/driven/snapshotlector"
	"jokai-client/internal/adapters/driven/sqlite"
	"jokai-client/internal/adapters/driving/tui"
	"jokai-client/internal/application"
	"jokai-client/internal/application/casosdeuso"
	"jokai-client/internal/domain"
)

func main() {
	rutaConfig := flag.String("config", "config.toml", "ruta del archivo de configuración TOML")
	flag.Parse()

	cfg, err := config.Cargar(*rutaConfig)
	var errorFatal error
	if err != nil {
		errorFatal = err
		cfg = domain.Configuracion{}
	}

	descubridor := fsdescubridor.Nuevo(
		cfg.RutaBotsDev,
		cfg.RutaBotsProd,
		cfg.Descubrimiento.ArchivoEntrada,
		cfg.Descubrimiento.ProfundidadMaxima,
	)
	lectorUsuarios := sqlite.NuevoLectorUsuarios()
	lectorSnapshot := snapshotlector.NuevoLector(cfg.RaizFileserver, cfg.Protocolo.Snapshot)
	emisor := filequeue.NuevoEmisor(cfg.RaizFileserver)
	generador := identificadores.Nuevo()

	identificar := casosdeuso.NewIdentificarUsuario(lectorUsuarios)
	descubrir := casosdeuso.NewDescubrirBots(descubridor)
	solicitar := casosdeuso.NewSolicitarEjecucion(emisor, lectorSnapshot, generador, application.RelojSistema{})
	listar := casosdeuso.NewListarEjecuciones(lectorSnapshot)
	ver := casosdeuso.NewVerEjecucion(lectorSnapshot)

	modelo := tui.Nuevo(cfg, identificar, descubrir, solicitar, listar, ver, errorFatal)

	if _, err := tea.NewProgram(modelo).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error al ejecutar la interfaz:", err)
		os.Exit(1)
	}
}