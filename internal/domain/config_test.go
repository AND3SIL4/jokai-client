package domain

import (
	"testing"
	"time"
)

func TestProtocoloConfigEsValida(t *testing.T) {
	valida := ProtocoloConfig{
		Snapshot:             "estado.json",
		TimeoutEjecucion:     4 * time.Hour,
		LatidoAgente:         60 * time.Second,
		UmbralAgenteInactivo: 5 * time.Minute,
		RetencionArchivos:    30 * 24 * time.Hour,
		RetencionHistorial:   90 * 24 * time.Hour,
	}
	if err := valida.EsValida(); err != nil {
		t.Fatalf("la config válida no debía fallar: %v", err)
	}

	casos := []struct {
		nombre   string
		modifica func(p *ProtocoloConfig)
		sub     string
	}{
		{"ruta en snapshot", func(p *ProtocoloConfig) { p.Snapshot = `sub\estado.json` }, "protocolo.snapshot"},
		{"snapshot vacío", func(p *ProtocoloConfig) { p.Snapshot = "" }, "protocolo.snapshot"},
		{"timeout cero", func(p *ProtocoloConfig) { p.TimeoutEjecucion = 0 }, "duraciones positivas"},
		{"latido negativo", func(p *ProtocoloConfig) { p.LatidoAgente = -time.Second }, "duraciones positivas"},
		{"retención cero", func(p *ProtocoloConfig) { p.RetencionHistorial = 0 }, "duraciones positivas"},
	}
	for _, c := range casos {
		p := valida
		c.modifica(&p)
		err := p.EsValida()
		if err == nil {
			t.Errorf("%s: se esperaba error", c.nombre)
			continue
		}
		if _, ok := err.(*ErrorUsuario); !ok {
			t.Errorf("%s: se esperaba ErrorUsuario, se obtuvo %T", c.nombre, err)
		}
	}
}