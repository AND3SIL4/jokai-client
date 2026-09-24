package casosdeuso

import (
	"context"
	"fmt"

	"jokai-client/internal/application"
	"jokai-client/internal/domain"
)

type SolicitarEjecucion struct {
	Emisor     application.EmisorComandos
	Snapshot   application.LectorSnapshot
	Generador  application.GeneradorIdentificadores
	Reloj      application.Reloj
}

func NewSolicitarEjecucion(
	emisor application.EmisorComandos,
	snapshot application.LectorSnapshot,
	generador application.GeneradorIdentificadores,
	reloj application.Reloj,
) *SolicitarEjecucion {
	return &SolicitarEjecucion{Emisor: emisor, Snapshot: snapshot, Generador: generador, Reloj: reloj}
}

func (uc *SolicitarEjecucion) Ejecutar(ctx context.Context, bot domain.Bot, usuario domain.Usuario) (domain.Solicitud, error) {
	if usuario.Rol.SoloLectura() {
		return domain.Solicitud{}, domain.NuevoErrorUsuario(
			"su rol no permite solicitar ejecuciones (solo lectura)",
			"contacte a soporte para obtener permisos de solicitud",
		)
	}

	snapshot, err := uc.Snapshot.Leer(ctx)
	if err != nil {
		return domain.Solicitud{}, err
	}

	maquina, err := uc.resolverMaquina(bot, snapshot)
	if err != nil {
		return domain.Solicitud{}, err
	}

	id, err := uc.Generador.NuevoID()
	if err != nil {
		return domain.Solicitud{}, domain.NuevoErrorUsuario(
			"no se pudo generar el identificador de la solicitud",
			"reintente la solicitud",
		)
	}

	solicitud, err := domain.NuevaSolicitud(id, bot, usuario, maquina, uc.Reloj.Ahora())
	if err != nil {
		return domain.Solicitud{}, err
	}

	if err := uc.Emisor.SolicitarEjecucion(ctx, solicitud); err != nil {
		return domain.Solicitud{}, err
	}

	return solicitud, nil
}

func (uc *SolicitarEjecucion) resolverMaquina(bot domain.Bot, snapshot domain.Snapshot) (string, error) {
	for _, maquina := range snapshot.Maquinas {
		if maquina.Ambiente != bot.Ambiente {
			continue
		}
		for _, nombre := range maquina.Bots {
			if nombre == bot.Nombre {
				return maquina.Nombre, nil
			}
		}
	}
	return "", domain.NuevoErrorUsuario(
		fmt.Sprintf("el bot '%s' no está asignado a ninguna máquina del ambiente %s", bot.Nombre, bot.Ambiente.String()),
		"publique la asignación del bot en el snapshot del consolidador (config del agente)",
	)
}