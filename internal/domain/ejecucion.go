package domain

import "time"

type EstadoEjecucion string

const (
	EstadoSolicitada EstadoEjecucion = "SOLICITADA"
	EstadoEnProceso  EstadoEjecucion = "EN_PROCESO"
	EstadoExitosa    EstadoEjecucion = "EXITOSA"
	EstadoFallida    EstadoEjecucion = "FALLIDA"
	EstadoCancelada  EstadoEjecucion = "CANCELADA"
)

func (e EstadoEjecucion) EsValido() bool {
	switch e {
	case EstadoSolicitada, EstadoEnProceso, EstadoExitosa, EstadoFallida, EstadoCancelada:
		return true
	default:
		return false
	}
}

func (e EstadoEjecucion) EsFinal() bool {
	return e == EstadoExitosa || e == EstadoFallida || e == EstadoCancelada
}

func (e EstadoEjecucion) Etiqueta() string {
	switch e {
	case EstadoSolicitada:
		return "Solicitada"
	case EstadoEnProceso:
		return "En proceso"
	case EstadoExitosa:
		return "Exitosa"
	case EstadoFallida:
		return "Fallida"
	case EstadoCancelada:
		return "Cancelada"
	default:
		return "Desconocido"
	}
}

type Ejecucion struct {
	ID          string
	SolicitudID string
	Estado      EstadoEjecucion
	Inicio      *time.Time
	Fin         *time.Time
	Codigo      *int
	Mensaje     *string
}

type Snapshot struct {
	VersionFormato int
	GeneradoEn     time.Time
	Maquinas       []MaquinaSnapshot
	Solicitudes    []SolicitudSnapshot
	Ejecuciones    []Ejecucion
}

type MaquinaSnapshot struct {
	Nombre       string
	Ambiente     Ambiente
	UltimoLatido *time.Time
	Bots         []string
}

type SolicitudSnapshot struct {
	ID          string
	BotNombre   string
	BotAmbiente Ambiente
	Maquina     string
	Usuario     string
	Fecha       time.Time
	Estado      EstadoEjecucion
}