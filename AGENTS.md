# AGENTS.md — guía para agentes de código en jokai-client

## Comandos

```powershell
go build ./...            # compila todo
go vet ./...              # análisis estático
go test ./...             # tests (constitución VI: verificar SIEMPRE tras cambios)
```

Al terminar cualquier tarea: `go build ./...` + `go vet ./...` + `go test ./...`
en verde. No commitees sin una orden explícita del usuario.

## Arquitectura (hexagonal estricta — constitución, .specify/memory/constitution.md)

- `internal/domain/`: entidades puras, sin importar nada fuera de la stdlib.
- `internal/application/`: puertos (interfaces) y casos de uso; jerga de negocio
  en español; NUNCA importa de `adapters/`.
- `internal/adapters/driven/`: infraestructura que el núcleo llama (config,
  filesystem, sqlite, filequeue).
- `internal/adapters/driving/tui/`: interfaz gráfica (Bubble Tea v2) hacia el usuario.
- `cmd/jokai/main.go`: únicamente wire del composition root.
- Regla crítica: SQLite es de ÚNICO escritor (el consolidador de jokai-server).
  El cliente solo LEE la BD. Prohibido escribir en adapters de cliente.

### Reglas de no-dependencia (T027, vérificalo manualmente en cada cambio)

`go list -deps ./internal/domain` y `go list -deps ./internal/application` deben
contener solo stdlib. Si aparece algo de infraestructura, es una violación.

## Convenciones

- IDs, mensajes de UI, errores y nombres en **español**.
- Dependencias Bubble Tea: SIEMPRE `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`,
  `charm.land/lipgloss/v2` (v2 movió los módulos fuera de github.com). No mezclar con v1.
- Config en TOML con strings literales (comillas simples) para rutas UNC.
- Errores hacia el usuario: `*domain.ErrorUsuario` con `Causa` + `Sugerencia`.
- Tests table-driven, código/identificadores en español, fakes inyectados por el constructor.

## Flujo SDD del repo

Usar los comandos de `.opencode/commands/speckit.*` (Spec Kit) para
`/speckit.specify`, `/speckit.clarify`, `/speckit.plan`, `/speckit.tasks`,
`/speckit.implement`, `/speckit.converge`. Cada funcionalidad vive en
`specs/NNN-nombre/` (spec, plan, research, contracts, quickstart, tasks) y en
`.specify/feature.json`. Las reglas obligatorias del proyecto están en la
constitución; cualquier cambio de regla se hace por enmienda (ver sección §6).

## Relación con jokai-server

Este repo es el cliente TUI. El agente ejecutor Python (una instancia por
máquina) y el consolidador (único escritor de la BD) viven en `jokai-server`
(otro repo, no creado todavía). El contrato entre ambos se define en
`specs/002-protocolo-orquestacion/` y NO puede cambiar sin versionarlo.