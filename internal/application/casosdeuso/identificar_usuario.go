package casosdeuso

import (
	"context"
	"os"
	"os/user"
	"strings"

	"jokai-client/internal/application"
	"jokai-client/internal/domain"
)

type IdentificarUsuario struct {
	lector    application.LectorUsuarios
	usuarioSO func() (nombre, completo string, err error)
}

func NewIdentificarUsuario(lector application.LectorUsuarios) *IdentificarUsuario {
	return &IdentificarUsuario{lector: lector, usuarioSO: usuarioDeWindows}
}

func usuarioDeWindows() (nombre, completo string, err error) {
	u, err := user.Current()
	if err != nil {
		nombre = os.Getenv("USERNAME")
		if nombre == "" {
			return "", "", err
		}
		return nombre, nombre, nil
	}
	completo = u.Username
	nombre = completo
	if i := strings.LastIndex(completo, `\`); i >= 0 && i+1 < len(completo) {
		nombre = completo[i+1:]
	}
	return nombre, completo, nil
}

func (uc *IdentificarUsuario) Ejecutar(ctx context.Context) (domain.Usuario, error) {
	nombre, completo, err := uc.usuarioSO()
	if err != nil {
		return domain.Usuario{}, domain.NuevoErrorUsuario(
			"no se pudo identificar el usuario de Windows",
			"verifique que la sesión tenga un usuario asociado",
		)
	}
	rol, err := uc.lector.RolDe(ctx, completo)
	if err != nil {
		return domain.Usuario{}, err
	}
	return domain.Usuario{
		WindowsNombre:   nombre,
		WindowsCompleto: completo,
		Rol:             rol,
	}, nil
}
