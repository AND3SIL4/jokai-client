package casosdeuso

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"jokai-client/internal/domain"
)

type emisorFake struct {
	solicitudes []domain.Solicitud
	err         error
}

func (e *emisorFake) SolicitarEjecucion(_ context.Context, s domain.Solicitud) error {
	if e.err != nil {
		return e.err
	}
	e.solicitudes = append(e.solicitudes, s)
	return nil
}

type snapshotFake struct {
	snapshot domain.Snapshot
	err      error
}

func (s *snapshotFake) Leer(context.Context) (domain.Snapshot, error) {
	return s.snapshot, s.err
}

type generadorFake struct {
	ids []string
	err error
}

func (g *generadorFake) NuevoID() (string, error) {
	if g.err != nil {
		return "", g.err
	}
	if len(g.ids) == 0 {
		return "550e8400-e29b-41d4-a716-446655440001", nil
	}
	id := g.ids[0]
	g.ids = g.ids[1:]
	return id, nil
}

type relojFake struct{ ahora time.Time }

func (r relojFake) Ahora() time.Time { return r.ahora }

func nuevoSolicitar() (*SolicitarEjecucion, *emisorFake, *snapshotFake, domain.Configuracion, domain.Usuario) {
	emisor := &emisorFake{}
	snapshot := &snapshotFake{
		snapshot: domain.Snapshot{
			VersionFormato: domain.VersionFormatoActual,
			Maquinas: []domain.MaquinaSnapshot{
				{
					Nombre:   "BOTPC01",
					Ambiente: domain.AmbienteProd(),
					Bots:     []string{"facturacion"},
				},
			},
		},
	}
	cfg := domain.Configuracion{}
	usuario := domain.Usuario{WindowsCompleto: `JOKAI\ana`, Rol: domain.RolNegocio}
	uc := NewSolicitarEjecucion(emisor, snapshot, &generadorFake{}, relojFake{time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC)})
	return uc, emisor, snapshot, cfg, usuario
}

func newBotSolicitud(t *testing.T, nombre string) domain.Bot {
	t.Helper()
	bot, err := domain.NuevoBot(nombre, domain.AmbienteProd(), `C:\bots\`+nombre, `C:\bots\`+nombre+`\main.py`)
	if err != nil {
		t.Fatal(err)
	}
	return bot
}

func TestSolicitarEjecucionExitoso(t *testing.T) {
	uc, emisor, _, _, usuario := nuevoSolicitar()
	bot := newBotSolicitud(t, "facturacion")

	sol, err := uc.Ejecutar(context.Background(), bot, usuario)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if len(emisor.solicitudes) != 1 {
		t.Fatalf("el emisor debe recibir 1 solicitud, recibió %d", len(emisor.solicitudes))
	}
	got := emisor.solicitudes[0]
	if got.Maquina != "BOTPC01" {
		t.Errorf("la máquina destino debe resolverse del snapshot, se obtuvo %q", got.Maquina)
	}
	if got.ID != "550e8400-e29b-41d4-a716-446655440001" {
		t.Errorf("ID = %q, se esperaba el generado", got.ID)
	}
	if got.VersionFormato != domain.VersionFormatoActual {
		t.Errorf("VersionFormato = %d", got.VersionFormato)
	}
	if want := time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC); !got.Fecha.Equal(want) {
		t.Errorf("Fecha = %v, se esperaba %v", got.Fecha, want)
	}
	if sol.ID != got.ID {
		t.Errorf("la solicitud devuelta debe ser la publicada, se obtuvo %q", sol.ID)
	}
}

func TestSolicitarEjecucionRolSoloLectura(t *testing.T) {
	uc, _, _, _, _ := nuevoSolicitar()
	bot := newBotSolicitud(t, "facturacion")
	usuarioSinPermiso := domain.Usuario{WindowsCompleto: `JOKAI\ana`, Rol: domain.RolSinMapear}

	_, err := uc.Ejecutar(context.Background(), bot, usuarioSinPermiso)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(strings.ToLower(eu.Causa), "solo lectura") {
		t.Errorf("la causa debe mencionar el rol de solo lectura: %q", eu.Causa)
	}
}

func TestSolicitarEjecucionBotNoAsignado(t *testing.T) {
	uc, emisor, _, _, usuario := nuevoSolicitar()
	bot := newBotSolicitud(t, "inventario")

	_, err := uc.Ejecutar(context.Background(), bot, usuario)
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if !strings.Contains(eu.Causa, "no está asignado a ninguna máquina") {
		t.Errorf("la causa debe indicar bot sin máquina: %q", eu.Causa)
	}
	if len(emisor.solicitudes) != 0 {
		t.Error("el emisor no debe recibir solicitudes si el bot no está asignado")
	}
}

func TestSolicitarEjecucionFallaAlPublicar(t *testing.T) {
	uc, _, _, _, _ := nuevoSolicitar()
	ito := &emisorFake{err: domain.NuevoErrorUsuario("fallo al publicar", "reintente")}
	uc.Emisor = ito
	bot := newBotSolicitud(t, "facturacion")

	_, err := uc.Ejecutar(context.Background(), bot, usuarioNegocio())
	var eu *domain.ErrorUsuario
	if !errors.As(err, &eu) {
		t.Fatalf("se esperaba ErrorUsuario, se obtuvo: %v", err)
	}
	if emisor := ito; len(emisor.solicitudes) != 0 {
		t.Errorf("el emisor falló, no debe registrar solicitudes; tiene %d", len(emisor.solicitudes))
	}
}

func usuarioNegocio() domain.Usuario {
	return domain.Usuario{WindowsCompleto: `JOKAI\ana`, Rol: domain.RolNegocio}
}