package domain

type Rol string

const (
	RolNegocio   Rol = "negocio"
	RolTecnico   Rol = "tecnico"
	RolSoporte   Rol = "soporte"
	RolSinMapear Rol = "sinmapear"
)

func (r Rol) Etiqueta() string {
	switch r {
	case RolNegocio:
		return "Negocio"
	case RolTecnico:
		return "Técnico"
	case RolSoporte:
		return "Soporte"
	default:
		return "Sin mapear (solo lectura)"
	}
}

func (r Rol) SoloLectura() bool {
	return r == RolSinMapear
}

type Usuario struct {
	WindowsNombre   string
	WindowsCompleto string
	Rol             Rol
}
