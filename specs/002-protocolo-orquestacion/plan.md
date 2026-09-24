# Plan de Implementación: Protocolo de orquestación vía fileserver

**Rama**: `002-protocolo-orquestacion` | **Fecha**: 2026-09-22 | **Spec**: [spec.md](./spec.md)

**Entrada**: Especificación de funcionalidad aclarada (`/speckit.clarify` 2026-09-22) de `/specs/002-protocolo-orquestacion/spec.md`.

## Summary

Definir y fijar el **protocolo de orquestación por fileserver** entre `jokai-client`
(TUI Go) y `jokai-server` (agentes Python, repo hermano futuro), e implementar en este
repo la parte de cliente:

- **HU1** publicar `Solicitud` → archivo de comando `.cmd` en `inbox/<máquina>/` con
  escritura atómica y UUID (adaptador real del puerto `EmisorComandos`).
- **HU2** consultar estado leyendo el **snapshot** `estado.json` publicado por el
  consolidador (adaptador lector real + pantallas de ejecuciones).
- **HU3** es **solo contrato** en este repo: se documenta el esquema de la BD, los
  formatos de comando/estado y la semántica del agente/consolidador para que
  `jokai-server` los implemente contra este spec versionado.
- **HU4** concurrencia y atomicidad: UUID por construcción, `O_CREATE|O_EXCL`,
  publicación por rename, take con `.tomando` y TTL de huérfanos.

Decisiones del `/clarify` integradas: lectura por snapshot (no BD directa), asignación
bot→máquina por config del agente (publicada en el snapshot), bootstrap de la BD por el
consolidador, timeouts 4h/60s/5min y retención 30d/90d configurables.

## Technical Context

- **Language/Version**: Go 1.26 (cliente TUI). El contrato compartido lo implementará
  Python (jokai-server), pero eso NO es trabajo de esta rama.
- **Primary Dependencies**: Bubble Tea v2 + bubbles v2 + lipgloss v2 (existentes);
  `github.com/pelletier/go-toml/v2` (config ampliada); `modernc.org/sqlite` SOLO como
  dependencia de tipo del lector (en esta rama el cliente ya no abre la BD: lee
  snapshot; se mantiene la dependencia para el contrato y futuros usos de solo lectura
  si el snapshot no basta). UUID con `crypto/rand` + stdlib (sin librerías nuevas:
  preserva la regla de no-dependencia de domain).
- **Storage**: fileserver SMB como puente: `inbox/<maquina>/*.cmd`, `estado/*.st`
  (publicados por agentes), `estado.json` (snapshot del consolidador),
  `procesados/YYYY/MM/DD/` (drenados). BD SQLite es fuente de verdad pero el cliente
  **no la toca** (constitución III).
- **Testing**: `go test ./...` con fakes de puertos y carpetas temporales locales que
  reproducen la topología del fileserver (inbox/estado/snapshot). Tests table-driven.
- **Target Platform**: Windows (PC del usuario); rutas UNC y letras de unidad.
- **Project Type**: Aplicación de escritorio TUI (cliente de un sistema de orquestación).
- **Performance Goals**: El TUI sigue un refresco configurable (`ui.intervalo_refresco`,
  defecto 2s); la lectura de snapshot nunca debe bloquear la interfaz (lecturas
  asíncronas con mensaje de resultado, patrón ya usado para el descubrimiento).
  CE-004: el estado de una ejecución se obtiene en segundos.
- **Constraints**:
  - El cliente escribe SOLO archivos de comando; jamás la BD (constitución III).
  - Sin sockets/APIs: toda comunicación es por el fileserver.
  - Publicación atómica: nunca se abre/trunca el archivo final; temporal + rename.
  - Nombres únicos por construcción (UUID v4 en minúsculas).
  - Todos los identificadores, mensajes de UI y errores en español.
  - `go build ./... && go vet ./... && go test ./...` en verde antes de cerrar tareas.
- **Scale/Scope**: 10+ TUI simultáneos, decenas de máquinas de bots; volumen de
  ejecuciones moderado (bots de negocio con retención 90d en BD).

## Constitution Check

*GATE: debe pasar antes de Phase 0 y revalidarse tras Phase 1.*

| Principio | Estado | Justificación |
|-----------|--------|---------------|
| I. Especificación antes que código | ✅ | Spec 002 aclarado; este plan deriva de las HUs y RF ratificados. |
| II. Hexagonal estricta | ✅ | Puertos en `internal/application` (`EmisorComandos`, nuevo `LectorSnapshot`, `GeneradorIdentificadores`); adaptadores `driven/filequeue` y `driven/snapshot`; dominio puro. Ningún tipo de infraestructura cruza el núcleo. |
| III. Único escritor de SQLite | ✅ | Refuerza: el cliente ni siquiera lee la BD, consume `estado.json`. La publicación de comandos es el único escribir del cliente y va a archivos planos. |
| IV. Español | ✅ | Contractos, mensajes y código en español. |
| V. Sin instalación | ✅ | Sigue siendo un único exe; la config se amplía con la sección `[protocolo]`. |
| VI. Testabilidad | ✅ | Protocolo probado contra carpetas temporales con la topología real del fileserver; casos de uso con fakes. |
| VII. YAGNI | ✅ | HU3 se limita a contrato (sin agente Python aquí); solo se implementa lo exigido por HU1/HU2/HU4 en su lado cliente. |

Gates: sin violaciones. La lectura por snapshot (decisión del /clarify) es coherente con
III; el UUID en stdlib preserva la regla de no-dependencia de `domain`.

## Project Structure

### Documentación (esta funcionalidad)

```text
specs/002-protocolo-orquestacion/
├── plan.md               # este archivo (/speckit.plan)
├── research.md           # investigación resuelta + decisiones /clarify
├── data-model.md         # entidades, validaciones, ciclo de estados
├── quickstart.md         # validación end-to-end del lado cliente
├── contracts/
│   ├── protocolo-fileserver.md   # formatos .cmd/.st, snapshot, layout, atomicidad, TTL
│   └── esquema-sqlite.md         # DDL + user_version + políticas de retención
└── tasks.md              # lo genera /speckit.tasks (NO se crea en /speckit.plan)
```

### Código fuente (raíz del repo)

```text
cmd/jokai/main.go                       # composition root: wire real de filequeue + snapshot
internal/domain/
├── solicitud.go                        # (existe) ampliar: máquina destino, version_formato
├── ejecucion.go                        # NUEVA: Ejecucion + EstadoEjecucion (SOLICITADA/EN_PROCESO/EXITOSA/FALLIDA/CANCELADA)
└── config.go                           # ampliar Configuracion: sección [protocolo]
internal/application/
├── puertos.go                          # ampliar: LectorSnapshot, GeneradorIdentificadores
└── casosdeuso/
    ├── solicitar_ejecucion.go          # NUEVA (HU1): valida rol+bot, resuelve máquina, publica
    ├── listar_ejecuciones.go           # NUEVA (HU2): lee snapshot, envuelve ResultadoEjecuciones
    └── ver_ejecucion.go                # NUEVA (HU2): detalle por id
internal/adapters/driven/
├── filequeue/emisor.go                 # REEMPLAZA stub (HU1): atómico, UUID, inbox/<maquina>/
├── snapshotlector/lector.go            # NUEVA (HU2): lee estado.json + detecta ausencia (RF-010)
└── config/toml.go                      # ampliar: valores [protocolo] con defaults
internal/adapters/driving/tui/
├── modelo.go                           # ampliar: pantalla ejecuciones + flujo solicitar
├── pantallas.go                        # ampliar: ListaEjecuciones, DetalleEjecucion, confirma solicitud
└── teclas.go                           # ampliar: bindings nuevos (solicitar, listar ejecuciones)
specs/002-protocolo-orquestacion/       # contratos versionados (Gobernanza: cambiar = user_version+1)
```

**Structure Decision**: Proyecto único existente (sin nuevos módulos de Go). El
protocolo se documenta en `contracts/` y NO se traduce a tipos compartidos de Go para
preservar la independencia de `domain`; los formatos se validan en los adaptadores.

## Complexity Tracking

> Sin violaciones constitucionales que justificar: tabla vacía.