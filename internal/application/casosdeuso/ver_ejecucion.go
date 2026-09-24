package casosdeuso

import (
	"context"
	"fmt"

	"jokai-client/internal/application"
	"jokai-client/internal/domain"
)

type VerEjecucion struct {
	snapshot application.LectorSnapshot
}

func NewVerEjecucion(snapshot application.LectorSnapshot) *VerEjecucion {
	return &VerEjecucion{snapshot: snapshot}
}

func (uc *VerEjecucion) Ejecutar(ctx context.Context, id string) (EjecucionConSolicitud, error) {
	snapshot, err := uc.snapshot.Leer(ctx)
	if err != nil {
		return EjecucionConSolicitud{}, err
	}

	var ejecucion *domain.Ejecucion
	for i := range snapshot.Ejecuciones {
		if snapshot.Ejecuciones[i].ID == id {
			ejecucion = &snapshot.Ejecuciones[i]
			break
		}
	}
	if ejecucion == nil {
		return EjecucionConSolicitud{}, domain.NuevoErrorUsuario(
			fmt.Sprintf("no se encontró la ejecución con id '%s'", id),
			"revise el listado de ejecuciones e introduzca un id válido",
		)
	}

	for _, s := range snapshot.Solicitudes {
		if s.ID == ejecucion.SolicitudID {
			return EjecucionConSolicitud{Ejecucion: *ejecucion, Solicitud: s}, nil
		}
	}
	return EjecucionConSolicitud{}, domain.NuevoErrorUsuario(
		fmt.Sprintf("la ejecución '%s' no tiene solicitud asociada en el snapshot", id),
		"revise la consistencia del snapshot publicado por el consolidador",
	)
}