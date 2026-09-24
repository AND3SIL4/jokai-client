package domain

import (
	"strings"
	"time"
)

type Solicitud struct {
	ID            string
	Bot           Bot
	Usuario       Usuario
	Maquina       string
	Fecha         time.Time
	VersionFormato int
}

func NuevaSolicitud(id string, bot Bot, usuario Usuario, maquina string, fecha time.Time) (Solicitud, error) {
	if !EsUUIDv4(id) {
		return Solicitud{}, NuevoErrorUsuario(
			"el identificador de la solicitud no es un UUID v4 válido",
			"reintente la solicitud",
		)
	}
	if !bot.EsValido() {
		return Solicitud{}, NuevoErrorUsuario(
			"la solicitud debe referenciar un bot válido (nombre, ambiente y rutas absolutas)",
			"seleccione un bot del catálogo y vuelva a solicitar",
		)
	}
	if strings.TrimSpace(usuario.WindowsCompleto) == "" {
		return Solicitud{}, NuevoErrorUsuario(
			"la solicitud debe registrar el usuario solicitante",
			"verifique que el sistema detecte su usuario de Windows",
		)
	}
	if strings.TrimSpace(maquina) == "" {
		return Solicitud{}, NuevoErrorUsuario(
			"la solicitud debe indicar la máquina destino del bot",
			"verifique que el bot esté asignado a una máquina en el snapshot",
		)
	}
	return Solicitud{
		ID:             id,
		Bot:            bot,
		Usuario:        usuario,
		Maquina:        maquina,
		Fecha:          fecha,
		VersionFormato: VersionFormatoActual,
	}, nil
}