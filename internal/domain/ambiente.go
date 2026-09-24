package domain

type Ambiente struct {
	valor string
}

func AmbienteDev() Ambiente  { return Ambiente{valor: "dev"} }
func AmbienteProd() Ambiente { return Ambiente{valor: "prod"} }

func ParseAmbiente(s string) (Ambiente, error) {
	switch s {
	case "dev":
		return AmbienteDev(), nil
	case "prod":
		return AmbienteProd(), nil
	default:
		return Ambiente{}, ErrAmbienteInvalido
	}
}

func (a Ambiente) String() string { return a.valor }

func (a Ambiente) Etiqueta() string {
	switch a.valor {
	case "dev":
		return "desarrollo"
	case "prod":
		return "producción"
	default:
		return "desconocido"
	}
}

func (a Ambiente) EsValido() bool {
	return a.valor == "dev" || a.valor == "prod"
}

func (a Ambiente) Alternativo() Ambiente {
	if a.valor == "dev" {
		return AmbienteProd()
	}
	return AmbienteDev()
}
