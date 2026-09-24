package sqlite

import (
	"context"

	"jokai-client/internal/domain"
)

type LectorUsuarios struct{}

func NuevoLectorUsuarios() *LectorUsuarios {
	return &LectorUsuarios{}
}

func (l *LectorUsuarios) RolDe(ctx context.Context, windowsCompleto string) (domain.Rol, error) {
	return domain.RolSinMapear, nil
}
