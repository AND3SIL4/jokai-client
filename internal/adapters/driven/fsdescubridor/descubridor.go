package fsdescubridor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"jokai-client/internal/domain"
)

type Descubridor struct {
	rutas             map[domain.Ambiente]string
	archivoEntrada    string
	profundidadMaxima int
}

func Nuevo(rutaDev, rutaProd, archivoEntrada string, profundidadMaxima int) *Descubridor {
	return &Descubridor{
		rutas: map[domain.Ambiente]string{
			domain.AmbienteDev():  rutaDev,
			domain.AmbienteProd(): rutaProd,
		},
		archivoEntrada:    archivoEntrada,
		profundidadMaxima: profundidadMaxima,
	}
}

func (d *Descubridor) Descubrir(ctx context.Context, ambiente domain.Ambiente) ([]domain.Bot, []string, error) {
	ruta, ok := d.rutas[ambiente]
	if !ok || ruta == "" {
		return nil, nil, domain.ErrAmbienteInvalido
	}
	entradas, err := os.ReadDir(ruta)
	if err != nil {
		return nil, nil, domain.NuevoErrorUsuario(
			fmt.Sprintf("no se pudo leer la carpeta del ambiente %s ('%s')", ambiente.Etiqueta(), ruta),
			"verifique que la ruta exista en config.toml y que tenga permisos de lectura",
		)
	}

	var bots []domain.Bot
	var advertencias []string
	for _, e := range entradas {
		if !e.IsDir() {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		nombre := e.Name()
		raizBot := filepath.Join(ruta, nombre)
		rutaEntrada, err := d.buscarEntrada(ctx, raizBot)
		if err != nil {
			advertencias = append(advertencias, fmt.Sprintf("no se pudo examinar el bot '%s': %v", nombre, err))
			continue
		}
		if rutaEntrada == "" {
			continue
		}
		bot, err := domain.NuevoBot(nombre, ambiente, raizBot, rutaEntrada)
		if err != nil {
			advertencias = append(advertencias, fmt.Sprintf("el bot '%s' se ignoró: %v", nombre, err))
			continue
		}
		bots = append(bots, bot)
	}
	sort.Slice(bots, func(i, j int) bool { return bots[i].Nombre < bots[j].Nombre })
	return bots, advertencias, nil
}

func (d *Descubridor) buscarEntrada(ctx context.Context, raizBot string) (string, error) {
	nivel := []string{raizBot}
	profundidad := 0
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		var siguientes []string
		for _, dir := range nivel {
			entradas, err := os.ReadDir(dir)
			if err != nil {
				return "", err
			}
			for _, e := range entradas {
				if !e.IsDir() && e.Name() == d.archivoEntrada {
					return filepath.Join(dir, e.Name()), nil
				}
			}
			for _, e := range entradas {
				if e.IsDir() {
					siguientes = append(siguientes, filepath.Join(dir, e.Name()))
				}
			}
		}
		if len(siguientes) == 0 {
			return "", nil
		}
		profundidad++
		if d.profundidadMaxima > 0 && profundidad > d.profundidadMaxima {
			return "", nil
		}
		nivel = siguientes
	}
}
