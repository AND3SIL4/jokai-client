package filequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"jokai-client/internal/domain"
)

type Emisor struct {
	raiz string
}

func NuevoEmisor(raiz string) *Emisor {
	return &Emisor{raiz: raiz}
}

type comandoJSON struct {
	VersionFormato int    `json:"version_formato"`
	ID             string `json:"id"`
	BotNombre      string `json:"bot_nombre"`
	BotAmbiente    string `json:"bot_ambiente"`
	BotRutaEntrada string `json:"bot_ruta_entrada"`
	Maquina        string `json:"maquina"`
	Usuario        string `json:"usuario"`
	Fecha          string `json:"fecha"`
}

func (e *Emisor) SolicitarEjecucion(ctx context.Context, solicitud domain.Solicitud) error {
	if _, err := os.Stat(e.raiz); err != nil {
		return domain.NuevoErrorUsuario(
			fmt.Sprintf("la raíz del fileserver '%s' no existe", e.raiz),
			"verifique el montaje del fileserver y la configuracion 'fileserver.raiz'",
		)
	}

	carpeta := filepath.Join(e.raiz, domain.CarpetaInbox, solicitud.Maquina)
	if err := os.MkdirAll(carpeta, 0o755); err != nil {
		return domain.NuevoErrorUsuario(
			fmt.Sprintf("no se pudo crear la carpeta de comandos '%s'", carpeta),
			"verifique los permisos de escritura en el fileserver",
		)
	}

	comando := comandoJSON{
		VersionFormato: solicitud.VersionFormato,
		ID:             solicitud.ID,
		BotNombre:      solicitud.Bot.Nombre,
		BotAmbiente:    solicitud.Bot.Ambiente.String(),
		BotRutaEntrada: solicitud.Bot.RutaEntrada,
		Maquina:        solicitud.Maquina,
		Usuario:        solicitud.Usuario.WindowsCompleto,
		Fecha:          solicitud.Fecha.UTC().Format("2006-01-02T15:04:05Z"),
	}
	datos, err := json.Marshal(comando)
	if err != nil {
		return domain.NuevoErrorUsuario(
			"no se pudo serializar el comando de ejecución",
			"reintente la solicitud",
		)
	}

	nombre := filepath.Join(carpeta, solicitud.ID+domain.SufijoComando)
	temp := nombre + domain.SufijoTemp

	// RF-001 (§8 del contrato): jamás se pisa un UUID existente ni se
	// abre/trunca el .cmd final; el .tmp se crea con O_CREATE|O_EXCL.
	if _, err := os.Stat(nombre); err == nil {
		return domain.NuevoErrorUsuario(
			fmt.Sprintf("ya existe un comando publicado con el identificador '%s'", solicitud.ID),
			"reintente la solicitud para generar un identificador nuevo",
		)
	}
	archivo, err := os.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return domain.NuevoErrorUsuario(
			fmt.Sprintf("no se pudo crear el comando temporal '%s'", temp),
			"verifique los permisos de escritura en el fileserver",
		)
	}
	if _, err := archivo.Write(datos); err != nil {
		archivo.Close()
		_ = os.Remove(temp)
		return domain.NuevoErrorUsuario(
			fmt.Sprintf("no se pudo escribir el comando temporal '%s'", temp),
			"verifique los permisos de escritura en el fileserver",
		)
	}
	if err := archivo.Close(); err != nil {
		_ = os.Remove(temp)
		return domain.NuevoErrorUsuario(
			fmt.Sprintf("no se pudo cerrar el comando temporal '%s'", temp),
			"verifique el estado del fileserver",
		)
	}
	if err := os.Rename(temp, nombre); err != nil {
		_ = os.Remove(temp)
		return domain.NuevoErrorUsuario(
			fmt.Sprintf("no se pudo publicar el comando '%s'", nombre),
			"verifique los permisos del fileserver",
		)
	}
	return nil
}