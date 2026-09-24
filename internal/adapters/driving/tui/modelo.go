package tui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/spinner"

	"jokai-client/internal/application/casosdeuso"
	"jokai-client/internal/domain"
)

type pantalla int

const (
	pantallaListado pantalla = iota
	pantallaDetalle
	pantallaError
	pantallaEjecuciones
	pantallaEjecucionDetalle
)

type Modelo struct {
	cfg                  domain.Configuracion
	identificar          *casosdeuso.IdentificarUsuario
	descubrir            *casosdeuso.DescubrirBots
	solicitar            *casosdeuso.SolicitarEjecucion
	listar               *casosdeuso.ListarEjecuciones
	ver                  *casosdeuso.VerEjecucion
	errorFatal           error
	pantalla             pantalla
	usuario              domain.Usuario
	usuarioListo         bool
	ambiente             domain.Ambiente
	bots                 []domain.Bot
	cursor               int
	escaneando           bool
	spinner              spinner.Model
	advertencias         []string
	errorEscaneo         error
	ejecuciones          []casosdeuso.EjecucionConSolicitud
	cursorEjecuciones    int
	detalleEjecucion     *casosdeuso.EjecucionConSolicitud
	cargandoEjecuciones  bool
	cargandoDetalle      bool
	errorEjecuciones     error
	cargandoSolicitud    bool
	solicitudPublicada   *domain.Solicitud
	errorSolicitud       error
	ancho                int
	alto                 int
	teclas               teclas
	ayuda                help.Model
	est                  estilos
}

func Nuevo(
	cfg domain.Configuracion,
	identificar *casosdeuso.IdentificarUsuario,
	descubrir *casosdeuso.DescubrirBots,
	solicitar *casosdeuso.SolicitarEjecucion,
	listar *casosdeuso.ListarEjecuciones,
	ver *casosdeuso.VerEjecucion,
	errorFatal error,
) *Modelo {
	m := &Modelo{
		cfg:         cfg,
		identificar: identificar,
		descubrir:   descubrir,
		solicitar:   solicitar,
		listar:      listar,
		ver:         ver,
		errorFatal:  errorFatal,
		ambiente:    domain.AmbienteDev(),
		spinner:     spinner.New(spinner.WithSpinner(spinner.Line)),
		teclas:      nuevasTeclas(),
		ayuda:       help.New(),
		est:         nuevosEstilos(),
	}
	if errorFatal != nil {
		m.pantalla = pantallaError
	}
	return m
}

func (m *Modelo) Init() tea.Cmd {
	if m.errorFatal != nil {
		return nil
	}
	return tea.Batch(m.cmdIdentificar(), m.cmdEscanear())
}

func (m *Modelo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ancho = msg.Width
		m.alto = msg.Height
		m.ayuda.SetWidth(msg.Width)
	case tea.KeyPressMsg:
		return m.manejarTecla(msg)
	case spinner.TickMsg:
		if m.escaneando || m.cargandoEjecuciones || m.cargandoDetalle || m.cargandoSolicitud {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case msjUsuarioIdentificado:
		if msg.err != nil {
			m.usuario = domain.Usuario{Rol: domain.RolSinMapear}
		} else {
			m.usuario = msg.usuario
		}
		m.usuarioListo = true
	case msjBotsDescubiertos:
		if msg.ambiente != m.ambiente {
			return m, nil
		}
		m.bots = msg.bots
		m.advertencias = msg.advertencias
		m.escaneando = false
		m.errorEscaneo = nil
		if m.cursor >= len(m.bots) {
			m.cursor = len(m.bots) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	case msjErrorEscaneo:
		if msg.ambiente != m.ambiente {
			return m, nil
		}
		m.escaneando = false
		m.errorEscaneo = msg.err
		m.bots = nil
		m.advertencias = nil
	case msjEjecucionesListadas:
		m.ejecuciones = msg.ejecuciones
		m.cargandoEjecuciones = false
		m.errorEjecuciones = nil
		if m.cursorEjecuciones >= len(m.ejecuciones) {
			m.cursorEjecuciones = len(m.ejecuciones) - 1
		}
		if m.cursorEjecuciones < 0 {
			m.cursorEjecuciones = 0
		}
	case msjErrorEjecuciones:
		m.cargandoEjecuciones = false
		m.errorEjecuciones = msg.err
		m.ejecuciones = nil
	case msjEjecucionCargada:
		m.cargandoDetalle = false
		if msg.err != nil {
			m.errorEjecuciones = msg.err
			m.detalleEjecucion = nil
		} else {
			detalle := msg.detalle
			m.detalleEjecucion = &detalle
			m.errorEjecuciones = nil
		}
	case msjTickEjecuciones:
		if m.pantalla == pantallaEjecuciones || m.pantalla == pantallaEjecucionDetalle {
			return m, tea.Batch(m.cmdListarEjecuciones(), m.cmdTickEjecuciones())
		}
	case msjSolicitudPublicada:
		m.cargandoSolicitud = false
		if msg.err != nil {
			m.errorSolicitud = msg.err
			m.solicitudPublicada = nil
		} else {
			solicitud := msg.solicitud
			m.solicitudPublicada = &solicitud
			m.errorSolicitud = nil
		}
	}
	return m, nil
}

func (m *Modelo) View() tea.View {
	switch m.pantalla {
	case pantallaError:
		return tea.NewView(m.vistaError())
	case pantallaDetalle:
		return tea.NewView(m.vistaDetalle())
	case pantallaEjecuciones:
		return tea.NewView(m.vistaEjecuciones())
	case pantallaEjecucionDetalle:
		return tea.NewView(m.vistaEjecucionDetalle())
	default:
		return tea.NewView(m.vistaListado())
	}
}

func (m *Modelo) manejarTecla(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.pantalla {
	case pantallaError:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil

	case pantallaListado:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.bots)-1 {
				m.cursor++
			}
		case "tab":
			m.conmutarAmbiente()
			return m, tea.Batch(m.cmdEscanear(), m.spinner.Tick)
		case "f5", "r":
			if !m.escaneando {
				m.escaneando = true
				m.bots = nil
				m.advertencias = nil
				return m, tea.Batch(m.cmdEscanear(), m.spinner.Tick)
			}
		case "enter":
			if len(m.bots) > 0 && !m.escaneando {
				m.solicitudPublicada = nil
				m.errorSolicitud = nil
				m.pantalla = pantallaDetalle
			}
		case "l":
			return m, m.abrirEjecuciones()
		}
		return m, nil

	case pantallaDetalle:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.pantalla = pantallaListado
		case "tab":
			m.conmutarAmbiente()
			m.pantalla = pantallaListado
			return m, tea.Batch(m.cmdEscanear(), m.spinner.Tick)
		case "l":
			return m, m.abrirEjecuciones()
		case "s":
			if len(m.bots) > 0 && m.cursor >= 0 && m.cursor < len(m.bots) &&
				m.usuarioListo && !m.cargandoSolicitud {
				m.cargandoSolicitud = true
				return m, tea.Batch(m.cmdSolicitar(), m.spinner.Tick)
			}
		}
		return m, nil

	case pantallaEjecuciones:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursorEjecuciones > 0 {
				m.cursorEjecuciones--
			}
		case "down", "j":
			if m.cursorEjecuciones < len(m.ejecuciones)-1 {
				m.cursorEjecuciones++
			}
		case "f5", "r":
			if !m.cargandoEjecuciones {
				m.cargandoEjecuciones = true
				m.ejecuciones = nil
				return m, tea.Batch(m.cmdListarEjecuciones(), m.spinner.Tick)
			}
		case "enter":
			if len(m.ejecuciones) > 0 && !m.cargandoEjecuciones {
				m.pantalla = pantallaEjecucionDetalle
				m.cargandoDetalle = true
				m.detalleEjecucion = nil
				m.errorEjecuciones = nil
				return m, tea.Batch(m.cmdVerEjecucion(m.ejecuciones[m.cursorEjecuciones].Ejecucion.ID), m.spinner.Tick)
			}
		case "esc":
			m.pantalla = pantallaListado
		case "tab":
			m.conmutarAmbiente()
			m.pantalla = pantallaListado
			return m, tea.Batch(m.cmdEscanear(), m.spinner.Tick)
		}
		return m, nil

	case pantallaEjecucionDetalle:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.pantalla = pantallaEjecuciones
		case "f5", "r":
			if m.detalleEjecucion != nil && !m.cargandoDetalle {
				m.cargandoDetalle = true
				return m, tea.Batch(m.cmdVerEjecucion(m.detalleEjecucion.Ejecucion.ID), m.spinner.Tick)
			}
		case "tab":
			m.conmutarAmbiente()
			m.pantalla = pantallaListado
			return m, tea.Batch(m.cmdEscanear(), m.spinner.Tick)
		}
		return m, nil
	}
	return m, nil
}

func (m *Modelo) conmutarAmbiente() {
	m.ambiente = m.ambiente.Alternativo()
	m.bots = nil
	m.advertencias = nil
	m.errorEscaneo = nil
	m.cursor = 0
	m.escaneando = true
}

func (m *Modelo) abrirEjecuciones() tea.Cmd {
	m.pantalla = pantallaEjecuciones
	m.cargandoEjecuciones = true
	m.ejecuciones = nil
	m.errorEjecuciones = nil
	m.cursorEjecuciones = 0
	return tea.Batch(m.cmdListarEjecuciones(), m.spinner.Tick, m.cmdTickEjecuciones())
}

func (m *Modelo) cmdIdentificar() tea.Cmd {
	uc := m.identificar
	return func() tea.Msg {
		u, err := uc.Ejecutar(context.Background())
		return msjUsuarioIdentificado{usuario: u, err: err}
	}
}

func (m *Modelo) cmdEscanear() tea.Cmd {
	uc := m.descubrir
	ambiente := m.ambiente
	return func() tea.Msg {
		res, err := uc.Ejecutar(context.Background(), ambiente)
		if err != nil {
			return msjErrorEscaneo{ambiente: ambiente, err: err}
		}
		return msjBotsDescubiertos{ambiente: ambiente, bots: res.Bots, advertencias: res.Advertencias}
	}
}

func (m *Modelo) cmdListarEjecuciones() tea.Cmd {
	uc := m.listar
	return func() tea.Msg {
		ejecuciones, err := uc.Ejecutar(context.Background())
		if err != nil {
			return msjErrorEjecuciones{err: err}
		}
		return msjEjecucionesListadas{ejecuciones: ejecuciones}
	}
}

func (m *Modelo) cmdVerEjecucion(id string) tea.Cmd {
	uc := m.ver
	return func() tea.Msg {
		detalle, err := uc.Ejecutar(context.Background(), id)
		return msjEjecucionCargada{detalle: detalle, err: err}
	}
}

func (m *Modelo) cmdSolicitar() tea.Cmd {
	uc := m.solicitar
	bot := m.bots[m.cursor]
	usuario := m.usuario
	return func() tea.Msg {
		solicitud, err := uc.Ejecutar(context.Background(), bot, usuario)
		return msjSolicitudPublicada{solicitud: solicitud, err: err}
	}
}

func (m *Modelo) cmdTickEjecuciones() tea.Cmd {
	intervalo := m.cfg.IntervaloRefresco
	return tea.Tick(intervalo, func(time.Time) tea.Msg {
		return msjTickEjecuciones{}
	})
}