package domain

import (
	"path/filepath"
	"strings"
)

type Bot struct {
	Nombre       string
	Ambiente     Ambiente
	RutaRaiz     string
	RutaEntrada  string
	Advertencias []string
}

func NuevoBot(nombre string, ambiente Ambiente, rutaRaiz, rutaEntrada string) (Bot, error) {
	if strings.TrimSpace(nombre) == "" {
		return Bot{}, NuevoErrorUsuario("el nombre del bot no puede estar vacío", "revise la carpeta del bot en el fileserver")
	}
	if !ambiente.EsValido() {
		return Bot{}, ErrAmbienteInvalido
	}
	if !filepath.IsAbs(rutaRaiz) {
		return Bot{}, NuevoErrorUsuario("la carpeta del bot debe ser una ruta absoluta", "revise la ruta del ambiente en config.toml")
	}
	if !filepath.IsAbs(rutaEntrada) {
		return Bot{}, NuevoErrorUsuario("el archivo de entrada del bot debe ser una ruta absoluta", "revise la ruta del ambiente en config.toml")
	}
	rel, err := filepath.Rel(rutaRaiz, rutaEntrada)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return Bot{}, NuevoErrorUsuario("el archivo de entrada debe estar dentro de la carpeta del bot", "revise la estructura de carpetas del bot")
	}
	return Bot{
		Nombre:      nombre,
		Ambiente:    ambiente,
		RutaRaiz:    rutaRaiz,
		RutaEntrada: rutaEntrada,
	}, nil
}

func (b Bot) EsValido() bool {
	return strings.TrimSpace(b.Nombre) != "" &&
		b.Ambiente.EsValido() &&
		filepath.IsAbs(b.RutaRaiz) &&
		filepath.IsAbs(b.RutaEntrada)
}
