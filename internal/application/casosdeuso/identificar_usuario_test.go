package casosdeuso

import (
	"context"
	"errors"
	"testing"

	"jokai-client/internal/domain"
)

type fakeLectorUsuarios struct {
	rol domain.Rol
	err error
}

func (f *fakeLectorUsuarios) RolDe(ctx context.Context, windowsCompleto string) (domain.Rol, error) {
	return f.rol, f.err
}

func TestIdentificarUsuarioRolMapeado(t *testing.T) {
	uc := NewIdentificarUsuario(&fakeLectorUsuarios{rol: domain.RolSoporte})
	uc.usuarioSO = func() (string, string, error) { return "andrea", `EMPRESA\andrea`, nil }

	u, err := uc.Ejecutar(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if u.WindowsNombre != "andrea" {
		t.Errorf("WindowsNombre = %q, se esperaba 'andrea'", u.WindowsNombre)
	}
	if u.WindowsCompleto != `EMPRESA\andrea` {
		t.Errorf("WindowsCompleto = %q, se esperaba 'EMPRESA\\andrea'", u.WindowsCompleto)
	}
	if u.Rol != domain.RolSoporte {
		t.Errorf("Rol = %q, se esperaba soporte", u.Rol)
	}
}

func TestIdentificarUsuarioSinMapeo(t *testing.T) {
	uc := NewIdentificarUsuario(&fakeLectorUsuarios{rol: domain.RolSinMapear})
	uc.usuarioSO = func() (string, string, error) { return "andrea", `EMPRESA\andrea`, nil }

	u, err := uc.Ejecutar(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if u.Rol != domain.RolSinMapear {
		t.Errorf("Rol = %q, se esperaba sinmapear", u.Rol)
	}
	if !u.Rol.SoloLectura() {
		t.Error("el rol sin mapear debe ser de solo lectura")
	}
}

func TestIdentificarUsuarioErrorDelLector(t *testing.T) {
	uc := NewIdentificarUsuario(&fakeLectorUsuarios{err: errors.New("bd caída")})
	uc.usuarioSO = func() (string, string, error) { return "andrea", `EMPRESA\andrea`, nil }

	if _, err := uc.Ejecutar(context.Background()); err == nil {
		t.Error("el error del lector debe propagarse")
	}
}

func TestIdentificarUsuarioErrorDelSO(t *testing.T) {
	uc := NewIdentificarUsuario(&fakeLectorUsuarios{rol: domain.RolNegocio})
	uc.usuarioSO = func() (string, string, error) { return "", "", errors.New("sin sesión") }

	_, err := uc.Ejecutar(context.Background())
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if eu.Causa == "" || eu.Sugerencia == "" {
		t.Error("el error debe llevar causa y sugerencia para el usuario")
	}
}
