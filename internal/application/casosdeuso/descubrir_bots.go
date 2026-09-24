package casosdeuso

import (
	"context"
	"sort"

	"jokai-client/internal/application"
	"jokai-client/internal/domain"
)

type ResultadoDescubrimiento struct {
	Bots         []domain.Bot
	Advertencias []string
	Ambiente     domain.Ambiente
}

type DescubrirBots struct {
	descubridor application.DescubridorBots
}

func NewDescubrirBots(descubridor application.DescubridorBots) *DescubrirBots {
	return &DescubrirBots{descubridor: descubridor}
}

func (uc *DescubrirBots) Ejecutar(ctx context.Context, ambiente domain.Ambiente) (ResultadoDescubrimiento, error) {
	if !ambiente.EsValido() {
		return ResultadoDescubrimiento{}, domain.ErrAmbienteInvalido
	}
	bots, advertencias, err := uc.descubridor.Descubrir(ctx, ambiente)
	if err != nil {
		return ResultadoDescubrimiento{}, err
	}
	sort.Slice(bots, func(i, j int) bool { return bots[i].Nombre < bots[j].Nombre })
	return ResultadoDescubrimiento{
		Bots:         bots,
		Advertencias: advertencias,
		Ambiente:     ambiente,
	}, nil
}
