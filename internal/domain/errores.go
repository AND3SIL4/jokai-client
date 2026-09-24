package domain

import "errors"

var (
	ErrAmbienteInvalido = errors.New("el ambiente debe ser 'dev' o 'prod'")
	ErrNoImplementado    = errors.New("esta funcionalidad aún no está implementada")
)

type ErrorUsuario struct {
	Causa      string
	Sugerencia string
}

func (e *ErrorUsuario) Error() string {
	if e.Sugerencia != "" {
		return e.Causa + ". " + e.Sugerencia
	}
	return e.Causa
}

func NuevoErrorUsuario(causa, sugerencia string) *ErrorUsuario {
	return &ErrorUsuario{Causa: causa, Sugerencia: sugerencia}
}
