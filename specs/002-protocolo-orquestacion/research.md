# Investigación Preliminar: Protocolo de orquestación

**Rama**: `002-protocolo-orquestacion` | **Fecha**: 2026-09-22 | **Estado**: resuelto (véase spec.md §Aclaraciones)

> Nota: investigación de Fase 0 real para `/speckit.plan`. Los puntos abiertos se
> resuelven en `/speckit.clarify` antes de planificar.

## 1. Atomicidad de archivos sobre SMB

- **Rename sobre SMB**: `os.Rename` dentro de la misma carpeta compartida es la operación
  más cercana a atómica disponible (POSIX rename semántico en SMB2+). Patrón: escribir
  `<id>.tmp` → renombrar a `<id>.json`/`<id>.cmd`. Los lectores solo consideran el nombre
  final.
- **Create-exclusivo**: `os.OpenFile(O_CREATE|O_EXCL)` funciona sobre SMB y sirve de
  candado leve (nombrado por UUID evita colisiones por construcción).
- Evitar: abrir y truncar el archivo final directamente (lectores pueden verlo a medias).

## 2. Patrón inbox/outbox (colas sobre carpeta compartida)

- Inbox por máquina destino (`inbox/<maquina>/<uuid>.cmd`) simplifica el "quién ejecuta
  qué" y evita disputas entre agentes: cada agente solo mira SU carpeta.
- Drenaje con "marcar al tomar": el consumidor renombra `.cmd` → `.tomando` antes de
  procesar (atomicidad del take). El consolidador hace lo mismo con estados (`.st` →
  `.tomando`).
- Idempotencia: si un proceso muere tras tomar, el archivo `.tomando` huérfano se detecta
  por antigüedad (TTL) y se reencola o aísla.

## 3. Único escritor de SQLite

- Constitución III: solo el consolidador abre la BD en escritura, en una máquina
  designada (el fileserver NO aloja procesos).
- journal_mode DELETE (WAL no opera bien sobre SMB) + `synchronous=FULL` + transacciones
  cortas.
- Lectores del TUI por SMB: conexiones cortas, read-only (`mode=ro`), con reintento ante
  `SQLITE_BUSY`. **Resuelto en /clarify**: snapshot exportado por el consolidador
  (`estado.json` publicado con rename) — elimina el riesgo de leer durante journal
  caliente; el TUI nunca abre la BD.

## 4. Esquema SQLite (propuesta inicial para /clarify)

- `PRAGMA user_version = N` como versionado del contrato (constitución, gobernanza).
- Tablas: `solicitudes` (id, bot_nombre, bot_ambiente, bot_ruta_entrada, maquina,
  usuario, fecha, estado), `ejecuciones` (id, solicitud_id, estado, inicio, fin,
  codigo, mensaje), `usuarios` (windows_completo, rol), `maquinas` (nombre, ambiente,
  ultimo_latido), `horarios` (reservado 006).
- La BD se crea por bootstrap del consolidador (jokai-server) con `PRAGMA user_version`;
  el TUI solo detecta la ausencia del snapshot (resuelto en /clarify).

## 5. Asignación bot→máquina (opciones para /clarify)

| Opción | Descripción | Implicación |
|---|---|---|
| A. Por config del agente | Cada agente declara qué bots (o carpetas) le tocan | Simple, estático; cambio = editar config de máquina |
| B. Por tabla en BD | El consolidador asigna y los agentes piden trabajo | Más flexible; requiere lógica de asignación |
| C. Reclamo (claim) | Cualquier agente del ambiente puede reclamar el comando | Bueno para máquinas intercambiables; riesgo de duplicados si la resolución de reclamo falla |

- Los bots viven en el fileserver, pero un bot de escritorio puede requerir una máquina
  concreta (sesión/credenciales) → **resuelto en /clarify**: opción A (config del agente
  declara sus bots; el snapshot publica la asignación y el TUI resuelve la máquina).

## 6. Ciclo de vida de estados

```
SOLICITADA → (agente toma) → EN_PROCESO → EXITOSA | FALLIDA
                                   ↘ CANCELADA (cancelación en curso)
```
- Cancelación: comando de control en el inbox de la máquina (`cancelar.cmd` con id) — el
  agente mata el proceso hijo si aún corre.
- Estados huérfanos: agente caído → heartbeat (`maquinas.ultimo_latido`) + timeout de
  ejecución configurable → el consolidador marca FALLIDA con mensaje de máquina
  inalcanzable (propuesta para /clarify).

## 7. Limpieza y retención (para /clarify)

- Archivos drenados → carpeta `procesados/YYYY/MM/DD/` (retención configurable).
- Estadísticas/historial en BD: retención por tiempo o cantidad; purga del consolidador.

## 8. Versionamiento del protocolo

- Cada comando/estado lleva `version_formato` entero. Reglas: aditivo = mismo número;
  incompatible = +1 y los lectores toleran `version <= la suya`.
- La BD se versiona aparte (`user_version`); los componentes validan al arrancar y avisan
  si son más viejos que la BD (CE-005).

## Puntos abiertos que alimentaron `/speckit.clarify` (resueltos 2026-09-22)

1. Lectura del TUI: **snapshot del consolidador** (`estado.json` por rename).
2. Asignación bot→máquina: **opción A** (config del agente; snapshot publica la asignación).
3. Bootstrap de la BD: **el consolidador** (jokai-server) la crea.
4. Timeout de ejecución y estados huérfanos: **4h / 60s latido / 5min umbral**, configurables.
5. Retención de archivos procesados y de historial en BD: **30 días / 90 días**, configurables.
