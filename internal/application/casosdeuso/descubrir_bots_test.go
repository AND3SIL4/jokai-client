package casosdeuso

import (
	"context"
	"errors"
	"testing"

	"jokai-client/internal/domain"
)

type fakeDescubridor struct {
	bots         []domain.Bot
	advertencias []string
	err          error
}

func (f *fakeDescubridor) Descubrir(ctx context.Context, ambiente domain.Ambiente) ([]domain.Bot, []string, error) {
	return f.bots, f.advertencias, f.err
}

func TestDescubrirBotsAmbienteInvalido(t *testing.T) {
	uc := NewDescubrirBots(&fakeDescubridor{})
	if _, err := uc.Ejecutar(context.Background(), domain.Ambiente{}); !errors.Is(err, domain.ErrAmbienteInvalido) {
		t.Errorf("se esperaba ErrAmbienteInvalido, se obtuvo: %v", err)
	}
}

func TestDescubrirBotsOrdenaPorNombre(t *testing.T) {
	rama := domain.AmbienteDev()
	b1, _ := domain.NuevoBot("zeta", rama, `C:\b\z`, `C:\b\z\main.py`)
	b2, _ := domain.NuevoBot("alfa", rama, `C:\b\a`, `C:\b\a\main.py`)
	uc := NewDescubrirBots(&fakeDescubridor{bots: []domain.Bot{b1, b2}})

	res, err := uc.Ejecutar(context.Background(), rama)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if res.Bots[0].Nombre != "alfa" || res.Bots[1].Nombre != "zeta" {
		t.Error("los bots deben venir ordenados por nombre")
	}
	if res.Ambiente != rama {
		t.Error("el resultado debe llevar el ambiente solicitado")
	}
}

func TestDescubrirBotsPropagaAdvertencias(t *testing.T) {
	uc := NewDescubrirBots(&fakeDescubridor{advertencias: []string{"algo pasó"}})
	res, err := uc.Ejecutar(context.Background(), domain.AmbienteProd())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(res.Advertencias) != 1 || res.Advertencias[0] != "algo pasó" {
		t.Error("las advertencias deben propagarse al resultado")
	}
}

func TestDescubrirBotsPropagaError(t *testing.T) {
	uc := NewDescubrirBots(&fakeDescubridor{err: errors.New("ruta caída")})
	if _, err := uc.Ejecutar(context.Background(), domain.AmbienteDev()); err == nil {
		t.Error("el error del descubridor debe propagarse")
	}
}
