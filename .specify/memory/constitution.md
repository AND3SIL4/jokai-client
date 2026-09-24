# Constitución de jokai-client

## Principios Core

### I. Especificación antes que código (NO NEGOCIABLE)

Toda funcionalidad nace de una especificación bajo `specs/` siguiendo el flujo SDD de Spec Kit:
`/speckit.specify` → `/speckit.clarify` (opcional) → `/speckit.plan` → `/speckit.tasks` →
`/speckit.implement` → `/speckit.converge`. No se implementa código de funcionalidad sin su spec
aprobada; los cambios de comportamiento exigen actualizar el spec antes que el código.

### II. Arquitectura hexagonal estricta

- El dominio (`internal/domain`) es Go puro: sin dependencias de Bubble Tea, SQLite, sistema de
  archivos ni red.
- Los puertos (interfaces) se definen en `internal/application` orientados al negocio, jamás
  filtrando tipos de infraestructura.
- Los adaptadores van en `internal/adapters`: `driving/tui` (Bubble Tea) y `driven/*`
  (sqlite, filequeue, fsdescubridor, config).
- La única composition root es `cmd/jokai/main.go`: es el único lugar donde se instancian
  adaptadores y se inyectan en los casos de uso.
- Regla de dependencias: `adapters → application → domain`. Jamás al revés, jamás lateral.

### III. Único escritor de SQLite (NO NEGOCIABLE)

La BD SQLite del fileserver es la fuente de verdad compartida del sistema, pero este cliente
**nunca la escribe**: solo abre conexiones de lectura de corta duración. Toda comunicación que
requiere escritura (solicitudes de ejecución, estados, logs) viaja como archivos planos (JSON)
en el fileserver; un único proceso consolidador es el único escritor de la BD. Este principio
existe porque los locks de SQLite no son confiables sobre carpetas compartidas (SMB) y una
corrupción de la BD perdería el historial completo de ejecuciones.

### IV. Español como idioma del proyecto

La UI, los mensajes al usuario, los specs, la documentación y el código (identificadores
incluidos) se escriben en español. Los errores del dominio son tipos con mensaje de usuario,
no cadenas sueltas.

### V. Distribución sin instalación

Cada componente se entrega como un único ejecutable autocontenido (Go: binario estático sin
CGO; Python: empaquetado con sus recursos). El usuario de negocio no instala nada: copia el
exe y ejecuta. La configuración vive en un archivo TOML editable y todo valor de entorno
(rutas) es explícito en él.

### VI. Testabilidad por diseño

Los casos de uso se prueban con fakes de los puertos, sin red ni fileserver real. El
descubrimiento de bots y el protocolo de archivos se prueban contra carpetas locales
temporales. Antes de declarar terminada cualquier tarea: `go build ./...`,
`go vet ./...` y `go test ./...` en verde.

### VII. Escalabilidad progresiva (YAGNI)

Se construye solo lo que el spec actual exige, pero los puertos se nombran y orientan para
que evolucionar (permisos por rol, ejecuciones programadas, nuevos comandos del protocolo)
sea agregar un adaptador o un caso de uso, jamás rewiring global.

## Restricciones Técnicas del Proyecto

- **Stack del cliente**: Go (última estable) + Bubble Tea v2 + bubbles v2 + lipgloss v2.
- **SQLite**: driver `modernc.org/sqlite` (Go puro, sin CGO) para compilación estática y
  cruzada sin toolchain C. Esquema versionado vía `PRAGMA user_version`.
- **Configuración**: TOML (`config.toml`, ejemplo en `config.example.toml`) con rutas del
  fileserver, la BD y los dos ambientes (`dev`, `prod`).
- **Roles**: negocio, técnico y soporte, mapeados desde el usuario de Windows vía tabla de
  la BD; usuario sin mapeo = solo lectura.
- **Comunicación**: exclusivamente vía el fileserver (archivos planos + BD leída). No hay
  sockets, APIs ni servicios entre cliente y agente.
- **Plataforma destino**: Windows (clientes y máquinas de bots); rutas UNC de red permitidas
  en toda la config.

## Flujo de Desarrollo

1. La funcionalidad nueva se describe con `/speckit.specify <descripción>` y se ratifica con
   el checklist de calidad del spec.
2. `/speckit.plan` produce plan.md (+ research, data-model, contracts, quickstart); los
   contratos del protocolo fileserver se documentan en `contracts/` del spec correspondiente.
3. `/speckit.tasks` genera las tareas por historia de usuario; cada historia es entregable y
   verificable de forma independiente.
4. `/speckit.implement` ejecuta las tareas respetando las reglas de dependencia hexagonal y
   la verificación completa (build + vet + test) antes de marcar tarea como hecha.
5. `/speckit.converge` cierra el ciclo; el trabajo restante se registra como tareas, no como
   deuda implícita.

## Gobernanza

- Esta constitución prevalece sobre cualquier atajo de implementación; si un requisito nuevo
  la contradice, se enmienda la constitución primero (documentando el porqué y el impacto).
- Toda enmienda registra versión, fecha y motivación en este archivo.
- El protocolo de comunicación fileserver es un contrato compartido con `jokai-server`
  (agente Python): cualquier cambio al esquema de la BD o al formato de archivos de comando
  exige incrementar `PRAGMA user_version` y actualizar el spec del protocolo (002) antes de
  tocar código.

**Versión**: 1.0.0 | **Ratificada**: 2026-09-22 | **Última enmienda**: 2026-09-22
