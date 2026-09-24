package tui

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"jokai-client/internal/domain"
)

const maxAdvertenciasVisibles = 3

func (m *Modelo) vistaListado() string {
	var b strings.Builder

	b.WriteString(m.vistaEncabezado())
	b.WriteString("\n\n")

	b.WriteString(m.vistaEstado())
	b.WriteString("\n\n")

	if len(m.bots) == 0 && !m.escaneando && m.errorEscaneo == nil {
		b.WriteString("No se encontraron bots en este ambiente.\n")
	}

	for i, bot := range m.bots {
		cursor := "  "
		estilo := m.est.bot
		if i == m.cursor {
			cursor = m.est.cursor.Render("› ")
			estilo = estilo.Bold(true)
		}
		rel, err := filepath.Rel(bot.RutaRaiz, bot.RutaEntrada)
		if err != nil {
			rel = bot.RutaEntrada
		}
		b.WriteString(cursor)
		b.WriteString(estilo.Render(bot.Nombre))
		b.WriteString(" ")
		b.WriteString(m.est.entrada.Render(rel))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.vistaAdvertencias())
	b.WriteString("\n")
	b.WriteString(m.est.ayuda.Render(m.ayuda.ShortHelpView(m.teclas.ayudaListado())))
	return m.est.app.Render(b.String())
}

func (m *Modelo) vistaEncabezado() string {
	var b strings.Builder
	b.WriteString(m.est.titulo.Render("jokai — orquestador de bots"))
	b.WriteString("\n")

	usuario := m.usuario.WindowsNombre
	if usuario == "" {
		usuario = "identificando…"
	}
	b.WriteString(m.est.usuario.Render(fmt.Sprintf("Usuario: %s (%s)", usuario, m.usuario.Rol.Etiqueta())))

	estAmbiente := m.est.ambienteDev
	if m.ambiente == domain.AmbienteProd() {
		estAmbiente = m.est.ambienteProd
	}
	b.WriteString("    ")
	b.WriteString(estAmbiente.Render(fmt.Sprintf("Ambiente: %s (%s)", m.ambiente.Etiqueta(), m.ambiente)))
	return b.String()
}

func (m *Modelo) vistaEstado() string {
	if m.escaneando {
		return m.est.estado.Render(fmt.Sprintf("%s escaneando bots del ambiente %s…", m.spinner.View(), m.ambiente.Etiqueta()))
	}
	if m.errorEscaneo != nil {
		return m.est.errorVista.Render("Error al escanear: " + m.errorEscaneo.Error())
	}
	return m.est.estado.Render(fmt.Sprintf("%d bots encontrados", len(m.bots)))
}

func (m *Modelo) vistaAdvertencias() string {
	if len(m.advertencias) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(m.est.advertencia.Render(fmt.Sprintf("Avisos del escaneo (%d):", len(m.advertencias))))
	b.WriteString("\n")
	visibles := m.advertencias
	if len(visibles) > maxAdvertenciasVisibles {
		visibles = visibles[:maxAdvertenciasVisibles]
	}
	for _, adv := range visibles {
		b.WriteString(m.est.advertencia.Render("  • " + adv))
		b.WriteString("\n")
	}
	if len(m.advertencias) > maxAdvertenciasVisibles {
		b.WriteString(m.est.advertencia.Render(fmt.Sprintf("  … y %d más", len(m.advertencias)-maxAdvertenciasVisibles)))
		b.WriteString("\n")
	}
	return b.String()
}

func (m *Modelo) vistaDetalle() string {
	if m.cursor < 0 || m.cursor >= len(m.bots) {
		return m.vistaListado()
	}
	bot := m.bots[m.cursor]

	var b strings.Builder
	b.WriteString(m.est.titulo.Render("Detalle del bot"))
	b.WriteString("\n\n")

	fila := func(clave, valor string) {
		b.WriteString(m.est.clave.Render(clave + ": "))
		b.WriteString(m.est.valor.Render(valor))
		b.WriteString("\n")
	}
	fila("Nombre", bot.Nombre)
	fila("Ambiente", bot.Ambiente.Etiqueta()+" ("+bot.Ambiente.String()+")")
	fila("Carpeta", bot.RutaRaiz)
	fila("Archivo de entrada", bot.RutaEntrada)

	b.WriteString("\n")
	if m.cargandoSolicitud {
		b.WriteString(m.est.estado.Render(fmt.Sprintf("%s publicando la solicitud…", m.spinner.View())))
		b.WriteString("\n")
	} else if m.errorSolicitud != nil {
		b.WriteString(m.vistaErrorEjecucion(m.errorSolicitud))
		b.WriteString("\n")
	} else if m.solicitudPublicada != nil {
		b.WriteString(m.est.estado.Render(fmt.Sprintf("Solicitud %s publicada en la máquina %s.", m.solicitudPublicada.ID, m.solicitudPublicada.Maquina)))
		b.WriteString("\n")
		b.WriteString(m.est.ayuda.Render("Siga su avance con l."))
		b.WriteString("\n")
	}
	b.WriteString(m.est.ayuda.Render(m.ayuda.ShortHelpView(m.teclas.ayudaDetalle())))
	return m.est.app.Render(b.String())
}

func (m *Modelo) vistaError() string {
	var b strings.Builder
	b.WriteString(m.est.errorVista.Render(m.est.titulo.Render("Error de configuración")))
	b.WriteString("\n\n")

	causa := ""
	sugerencia := ""
	var eu *domain.ErrorUsuario
	if errors.As(m.errorFatal, &eu) {
		causa = eu.Causa
		sugerencia = eu.Sugerencia
	} else if m.errorFatal != nil {
		causa = m.errorFatal.Error()
	}

	b.WriteString(m.est.errorVista.Render(causa))
	b.WriteString("\n")
	if sugerencia != "" {
		b.WriteString("\n")
		b.WriteString(m.est.usuario.Render("Sugerencia: " + sugerencia))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(m.est.ayuda.Render("Presione q para salir."))
	return m.est.app.Render(b.String())
}

func (m *Modelo) vistaErrorEjecucion(err error) string {
	var eu *domain.ErrorUsuario
	if errors.As(err, &eu) {
		salida := m.est.errorVista.Render(eu.Causa)
		if eu.Sugerencia != "" {
			salida += "\n" + m.est.usuario.Render("Sugerencia: " + eu.Sugerencia)
		}
		return salida
	}
	return m.est.errorVista.Render(err.Error())
}

func (m *Modelo) vistaEjecuciones() string {
	var b strings.Builder

	b.WriteString(m.est.titulo.Render("jokai — ejecuciones"))
	b.WriteString("\n\n")

	if m.cargandoEjecuciones {
		b.WriteString(m.est.estado.Render(fmt.Sprintf("%s consultando el snapshot del consolidador…", m.spinner.View())))
		b.WriteString("\n\n")
	}
	if m.errorEjecuciones != nil && !m.cargandoEjecuciones {
		b.WriteString(m.vistaErrorEjecucion(m.errorEjecuciones))
		b.WriteString("\n\n")
	}
	if len(m.ejecuciones) == 0 && !m.cargandoEjecuciones && m.errorEjecuciones == nil {
		b.WriteString("No hay ejecuciones publicadas en el snapshot.\n\n")
	}

	for i, e := range m.ejecuciones {
		cursor := "  "
		estilo := m.est.bot
		if i == m.cursorEjecuciones {
			cursor = m.est.cursor.Render("› ")
			estilo = estilo.Bold(true)
		}
		linea := fmt.Sprintf("%s  %s  %s  %s",
			e.Solicitud.Fecha.Local().Format("2006-01-02 15:04"),
			e.Solicitud.BotNombre,
			e.Ejecucion.Estado.Etiqueta(),
			e.Ejecucion.ID,
		)
		b.WriteString(cursor)
		b.WriteString(estilo.Render(linea))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.est.ayuda.Render(m.ayuda.ShortHelpView(m.teclas.ayudaEjecuciones())))
	return m.est.app.Render(b.String())
}

func (m *Modelo) vistaEjecucionDetalle() string {
	var b strings.Builder
	b.WriteString(m.est.titulo.Render("Detalle de la ejecución"))
	b.WriteString("\n\n")

	if m.cargandoDetalle {
		b.WriteString(m.est.estado.Render(fmt.Sprintf("%s consultando el snapshot…", m.spinner.View())))
	} else if m.detalleEjecucion == nil {
		if m.errorEjecuciones != nil {
			b.WriteString(m.vistaErrorEjecucion(m.errorEjecuciones))
		} else {
			b.WriteString("Seleccione una ejecución del listado.")
		}
	} else {
		d := m.detalleEjecucion
		fila := func(clave, valor string) {
			b.WriteString(m.est.clave.Render(clave + ": "))
			b.WriteString(m.est.valor.Render(valor))
			b.WriteString("\n")
		}
		fila("Ejecución", d.Ejecucion.ID)
		fila("Estado", d.Ejecucion.Estado.Etiqueta())
		fila("Bot", d.Solicitud.BotNombre)
		fila("Ambiente", d.Solicitud.BotAmbiente.Etiqueta()+" ("+d.Solicitud.BotAmbiente.String()+")")
		fila("Máquina", d.Solicitud.Maquina)
		fila("Usuario", d.Solicitud.Usuario)
		fila("Solicitud", d.Solicitud.ID)
		fila("Solicitada", d.Solicitud.Fecha.Local().Format("2006-01-02 15:04:05"))
		if d.Ejecucion.Inicio != nil {
			fila("Inicio", d.Ejecucion.Inicio.Local().Format("2006-01-02 15:04:05"))
		}
		if d.Ejecucion.Fin != nil {
			fila("Fin", d.Ejecucion.Fin.Local().Format("2006-01-02 15:04:05"))
		}
		if d.Ejecucion.Codigo != nil {
			fila("Código de salida", fmt.Sprintf("%d", *d.Ejecucion.Codigo))
		}
		if d.Ejecucion.Mensaje != nil {
			fila("Mensaje", *d.Ejecucion.Mensaje)
		}
	}

	b.WriteString("\n")
	b.WriteString(m.est.ayuda.Render(m.ayuda.ShortHelpView(m.teclas.ayudaEjecucionDetalle())))
	return m.est.app.Render(b.String())
}
