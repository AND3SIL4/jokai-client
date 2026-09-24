package tui

import "charm.land/bubbles/v2/key"

type teclas struct {
	arriba      key.Binding
	abajo       key.Binding
	detalle     key.Binding
	regresar    key.Binding
	refrescar   key.Binding
	ambiente    key.Binding
	ejecuciones key.Binding
	solicitar   key.Binding
	salir       key.Binding
}

func nuevasTeclas() teclas {
	return teclas{
		arriba:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "mover arriba")),
		abajo:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "mover abajo")),
		detalle:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "ver detalle")),
		regresar:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "regresar")),
		refrescar: key.NewBinding(key.WithKeys("f5", "r"), key.WithHelp("F5/r", "refrescar")),
		ambiente:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "cambiar ambiente")),
		ejecuciones: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "ejecuciones")),
		solicitar:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "solicitar ejecución")),
		salir:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "salir")),
	}
}

func (t teclas) ayudaListado() []key.Binding {
	return []key.Binding{t.arriba, t.abajo, t.detalle, t.ejecuciones, t.refrescar, t.ambiente, t.salir}
}

func (t teclas) ayudaDetalle() []key.Binding {
	return []key.Binding{t.solicitar, t.ejecuciones, t.regresar, t.ambiente, t.salir}
}

func (t teclas) ayudaEjecuciones() []key.Binding {
	return []key.Binding{t.arriba, t.abajo, t.detalle, t.refrescar, t.regresar, t.ambiente, t.salir}
}

func (t teclas) ayudaEjecucionDetalle() []key.Binding {
	return []key.Binding{t.regresar, t.refrescar, t.ambiente, t.salir}
}
