package tui

import (
	"jokai-client/internal/application/casosdeuso"
	"jokai-client/internal/domain"
)

type msjBotsDescubiertos struct {
	ambiente     domain.Ambiente
	bots         []domain.Bot
	advertencias []string
}

type msjErrorEscaneo struct {
	ambiente domain.Ambiente
	err      error
}

type msjUsuarioIdentificado struct {
	usuario domain.Usuario
	err     error
}

type msjEjecucionesListadas struct {
	ejecuciones []casosdeuso.EjecucionConSolicitud
}

type msjErrorEjecuciones struct {
	err error
}

type msjEjecucionCargada struct {
	detalle casosdeuso.EjecucionConSolicitud
	err     error
}

type msjTickEjecuciones struct{}

type msjSolicitudPublicada struct {
	solicitud domain.Solicitud
	err       error
}