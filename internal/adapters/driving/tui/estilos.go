package tui

import "charm.land/lipgloss/v2"

type estilos struct {
	app          lipgloss.Style
	titulo       lipgloss.Style
	usuario      lipgloss.Style
	ambienteDev  lipgloss.Style
	ambienteProd lipgloss.Style
	estado       lipgloss.Style
	cursor       lipgloss.Style
	bot          lipgloss.Style
	entrada      lipgloss.Style
	advertencia  lipgloss.Style
	errorVista   lipgloss.Style
	ayuda        lipgloss.Style
	clave        lipgloss.Style
	valor        lipgloss.Style
}

func nuevosEstilos() estilos {
	return estilos{
		app:          lipgloss.NewStyle().Padding(1, 2),
		titulo:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		usuario:      lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		ambienteDev:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")),
		ambienteProd: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203")),
		estado:       lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		cursor:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		bot:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		entrada:      lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		advertencia:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		errorVista:   lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		ayuda:        lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		clave:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")),
		valor:        lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	}
}
