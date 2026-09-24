# jokai-client

Cliente TUI de orquestación de bots para **jokai**. Permite a usuarios de negocio,
técnicos y de soporte descubrir, monitorear y (en fases futuras) ejecutar bots de
Python migrados desde Automation Anywhere **sin ingresar a las máquinas**.

## Arquitectura

```
┌─────────────────┐          ┌──────────── Fileserver (carpeta SMB) ─────────────┐
│  jokai-client    │  lee ◀──│  BD SQLite (fuente de verdad, único escritor)     │
│  (TUI Go/BT v2)  │         │  inbox/ comandos · estado/ reportes de agentes   │
│  PC del usuario  │  ─────▶ │  bots dev/prod (subcarpeta = bot, main.py)       │
└─────────────────┘ escribe  └────────────▲───────────────────▲─────────────────┘
            solo comandos (specs 002+)   │                   │
                          ┌──────────────┴──────┐           │
                          │ jokai-server (Python)│───────────┘
                          │ agente por máquina   │  ejecuta main.py
                          │ + consolidador único │  (repo hermano)
                          └─────────────────────┘
```

- **Descubrimiento de bots**: el catálogo vive en el filesystem. Cada subcarpeta
  directa de la ruta del ambiente es un bot (nombre = carpeta); se busca
  `main.py` (configurable) de forma recursiva con límite de profundidad.
- **SQLite**: el cliente **nunca escribe** la BD (regla de único escritor, ver
  constitución). En esta fase solo lee carpetas; el lector de BD llega con el
  spec 002.
- **Ambientes**: `dev` y `prod` conmutables en caliente (tecla Tab).

## Requisitos

- Go 1.26+ (para compilar). El binario resultante es autocontenido: el usuario
  final no instala nada.

## Construcción y ejecución

```powershell
go build -o bin\jokai-client.exe .\cmd\jokai

# con configuración (copie config.example.toml → config.toml y ajuste rutas)
.\bin\jokai-client.exe
.\bin\jokai-client.exe --config C:\ruta\a\config.toml
```

Teclas: `↑/↓` navegar · `enter` detalle · `tab` ambiente dev/prod · `F5/r`
refrescar · `q` salir.

## Verificación

```powershell
go build ./... ; go vet ./... ; go test ./...
```

## Desarrollo con SDD (Spec Kit)

Este proyecto sigue **Spec-Driven Development** con [GitHub Spec Kit](https://github.com/github/spec-kit).
El flujo por funcionalidad:

```
/speckit.specify   → specs/NNN-*/spec.md (requisitos, historias de usuario)
/speckit.clarify   → (opcional) resuelve ambigüedades
/speckit.plan      → plan.md + research + data-model + contracts + quickstart
/speckit.tasks     → tasks.md (tareas por historia de usuario)
/speckit.implement → implementación verificada (build + vet + test)
/speckit.converge  → cierre del ciclo
```

Principios obligatorios: constitución en `.specify/memory/constitution.md`
(hexagonal estricta, único escritor de SQLite, español, sin CGO, YAGNI).

## Hoja de ruta de specs

| # | Funcionalidad | Estado |
|---|---------------|--------|
| 001 | Fundación del cliente TUI (arranque, config, descubrimiento, navegación) | Implementada |
| 002 | Protocolo de orquestación vía fileserver | Implementada (cliente) · contrato cerrado |
| 003 | Disparo manual de ejecución | Pendiente |
| 004 | Seguimiento en vivo (estados + logs) | Pendiente |
| 005 | Roles y permisos por usuario de Windows | Pendiente |
| 006 | Ejecuciones programadas | Pendiente |
| 007 | Distribución (exe) y bootstrap de la BD | Pendiente |

## Estructura

```
cmd/jokai/            composition root (única: inyección de adaptadores)
internal/domain/      entidades puras: Bot, Ambiente, Usuario, Rol, Config, errores,
                      protocolo (carpetas/sufijos), Solicitud, Ejecucion, Snapshot
internal/application/ puertos (interfaces) + casos de uso en español
internal/adapters/
  driving/tui/        Bubble Tea v2 (modelo, pantallas, teclas, mensajes, estilos)
  driven/config/      carga + validación de config.toml (se incluye [protocolo])
  driven/fsdescubridor/ escáner de bots (subcarpetas + main.py recursivo)
  driven/identificadores/ UUID v4 (crypto/rand) para solicitudes y comandos
  driven/snapshotlector/ lector del snapshot `estado.json` publicado por el consolidador
  driven/filequeue/   emisor atómico de comandos (`.cmd.tmp` → rename → `.cmd`)
                      + lector `Snapshot/Lector` para ciclos de integración
  driven/sqlite/       lector de usuarios (stub hasta 005)
specs/                 especificaciones SDD (002 protocolo de orquestación)
.specify/ memory/       infraestructura Spec Kit + constitución
```

## Protocolo de orquestación (spec 002)

El cliente nunca escribe en la BD SQLite ni actúa como agente: la comunicación con
los bots usa un **fileserver** como buzón de archivos (carpeta SMB), y el estado se
lee de un **snapshot** que publica el consolidador.

Flujo TUI → fileserver → jokai-server:

```
usuario (TUI) ──solicitar──▶ cmd/jokai ──publica──▶ inbox/<maquina>/<uuid>.cmd
                                                          │ rename atómico
                                                          ▼
                                              agente ejecutor (jokai-server)
                                                          │ escribe
                                                          ▼
                                        estado/<uuid>.st · snapshot estado.json
                                                          │ lee
                                                        TUI (F5 / refresco)
```

- **Comandos**: el TUI publica `<uuid>.cmd` en `inbox/<maquina>/` mediante una
  escritura a `<uuid>.cmd.tmp` seguida de `rename` (publicación atómica, nunca se
  abre/trunca el `.cmd` final). El identificador es un UUID v4; el mismo UUID jamás
  se publica dos veces.
- **Snapshots**: el consolidador (único escritor) publica `estado.json` por la
  misma técnica atómica. El cliente solo lo **lee**; si no existe muestra un aviso
  en pantalla (sin snapshot → no hay ejecuciones que listar).
- **Contrato**: el formato de comandos `.cmd`, del snapshot y del esquema SQLite se
  definen en `specs/002-protocolo-orquestacion/contracts/`; el agente ejecutor y el
  consolidador los implementa `jokai-server` (repo hermano), según ese mismo
  contrato (spec 002, HU3).
- **Configuración**: la sección `[protocolo]` de `config.toml` fija el tiempo de
  espera de ejecución, el latido de los agentes y la retención de archivos (ver
  `config.example.toml` y `specs/002-protocolo-orquestacion/quickstart.md`).

## Relación con jokai-server

El agente ejecutor (Python, una instancia por máquina de bots) y el
consolidador (único escritor de la BD) viven en el repo hermano
`jokai-server`. Su contrato (formato de comandos, archivos de estado y esquema
SQLite) se define en `specs/002-protocolo-orquestacion/`.
