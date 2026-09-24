package application

import (
	"context"
	"time"

	"jokai-client/internal/domain"
)

type DescubridorBots interface {
	Descubrir(ctx context.Context, ambiente domain.Ambiente) ([]domain.Bot, []string, error)
}

type LectorUsuarios interface {
	RolDe(ctx context.Context, windowsCompleto string) (domain.Rol, error)
}

type EmisorComandos interface {
	SolicitarEjecucion(ctx context.Context, solicitud domain.Solicitud) error
}

type Reloj interface {
	Ahora() time.Time
}

type RelojSistema struct{}

func (RelojSistema) Ahora() time.Time { return time.Now() }

type LectorSnapshot interface {
	Leer(ctx context.Context) (domain.Snapshot, error)
}

type GeneradorIdentificadores interface {
	NuevoID() (string, error)
}
