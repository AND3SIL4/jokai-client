package casosdeuso

import (
	"context"
	"sort"

	"jokai-client/internal/application"
	"jokai-client/internal/domain"
)

type EjecucionConSolicitud struct {
	Ejecucion domain.Ejecucion
	Solicitud domain.SolicitudSnapshot
}

type ListarEjecuciones struct {
	snapshot application.LectorSnapshot
}

func NewListarEjecuciones(snapshot application.LectorSnapshot) *ListarEjecuciones {
	return &ListarEjecuciones{snapshot: snapshot}
}

func (uc *ListarEjecuciones) Ejecutar(ctx context.Context) ([]EjecucionConSolicitud, error) {
	return uc.listar(ctx, "", domain.Ambiente{}, false)
}

func (uc *ListarEjecuciones) EjecutarPorBot(ctx context.Context, nombre string, ambiente domain.Ambiente) ([]EjecucionConSolicitud, error) {
	return uc.listar(ctx, nombre, ambiente, true)
}

func (uc *ListarEjecuciones) listar(ctx context.Context, botNombre string, ambiente domain.Ambiente, filtrado bool) ([]EjecucionConSolicitud, error) {
	snapshot, err := uc.snapshot.Leer(ctx)
	if err != nil {
		return nil, err
	}

	porSolicitud := make(map[string]domain.SolicitudSnapshot, len(snapshot.Solicitudes))
	for _, s := range snapshot.Solicitudes {
		porSolicitud[s.ID] = s
	}

	resultado := make([]EjecucionConSolicitud, 0, len(snapshot.Ejecuciones))
	for _, e := range snapshot.Ejecuciones {
		solicitud, ok := porSolicitud[e.SolicitudID]
		if !ok {
			continue
		}
		if filtrado && (solicitud.BotNombre != botNombre || solicitud.BotAmbiente != ambiente) {
			continue
		}
		resultado = append(resultado, EjecucionConSolicitud{Ejecucion: e, Solicitud: solicitud})
	}

	sort.SliceStable(resultado, func(i, j int) bool {
		return resultado[i].Solicitud.Fecha.After(resultado[j].Solicitud.Fecha)
	})
	return resultado, nil
}