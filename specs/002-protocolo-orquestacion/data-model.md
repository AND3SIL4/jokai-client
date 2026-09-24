# Modelo de Datos: Protocolo de orquestación

**Rama**: `002-protocolo-orquestacion` | **Fecha**: 2026-09-22

Entidades del protocolo orquestación. El cliente **solo** crea `Solicitud` (archivo
`.cmd`) y lee `Ejecucion`/`Maquina` (vía snapshot `estado.json`). El resto lo
materializa el consolidador en la BD (jokai-server).

## Solicitud

Pedido de ejecución publicado por el TUI como comando.

| Campo | Tipo | Reglas |
|---|---|---|
| `id` | string | UUID v4 en minúsculas, único por construcción |
| `version_formato` | int | `>= 1`; reglas de coexistencia en contracts/protocolo-fileserver.md §7 |
| `bot_nombre` | string | no vacío; nombre de carpeta del bot |
| `bot_ambiente` | string | `dev` \| `prod` |
| `bot_ruta_entrada` | string | ruta absoluta (UNC o letra) del `main.py`, fotografiada en el momento de solicitar |
| `maquina` | string | nombre de máquina destino (resuelto desde el snapshot) |
| `usuario` | string | `DOMINIO\usuario` del solicitante (identificador Windows) |
| `fecha` | string | ISO 8601 UTC de publicación |

**Validación**: `NuevoBot` existente cubre nombre/ruta; `Solicitud` añade `id` bien
formado (UUID), `maquina` no vacía, `usuario` no vacío, `version_formato >= 1`.

## Ejecucion

Realización de una solicitud; proviene del snapshot (el cliente no la escribe).

| Campo | Tipo | Reglas |
|---|---|---|
| `id` | string | UUID |
| `solicitud_id` | string | referencia a la Solicitud |
| `estado` | `EstadoEjecucion` | SOLICITADA \| EN_PROCESO \| EXITOSA \| FALLIDA \| CANCELADA |
| `inicio` | string \| null | ISO 8601 UTC |
| `fin` | string \| null | ISO 8601 UTC |
| `codigo` | int \| null | código de salida del proceso (si terminó) |
| `mensaje` | string \| null | mensaje/error reportado |

## Maquina

Agente de una máquina de bots; proviene del snapshot.

| Campo | Tipo | Reglas |
|---|---|---|
| `nombre` | string | identificador de la máquina |
| `ambiente` | string | `dev` \| `prod` |
| `ultimo_latido` | string \| null | ISO 8601 UTC; `null` si nunca latió |
| `bots` | []string | carpetas de bots que la máquina declara (config del agente) |

La asignación bot→máquina (decisión /clarify: opción A) se resuelve buscando el bot
`dev`/`prod` en `bots` de las máquinas del mismo ambiente. Sin máquina que lo reclame,
la solicitud se rechaza con error claro (ver Casos Límite del spec).

## Snapshot (interfaz de lectura del TUI)

`estado.json` (copia materializada que publica el consolidador por rename). Formato
completo en contracts/protocolo-fileserver.md §5.

```jsonc
{
  "version_formato": 1,
  "generado_en": "2026-09-22T14:00:00Z",
  "maquinas": [
    { "nombre": "BOTPC01", "ambiente": "prod", "ultimo_latido": "...", "bots": ["facturacion"] }
  ],
  "solicitudes": [
    { "id": "...", "bot_nombre": "facturacion", "bot_ambiente": "prod",
      "maquina": "BOTPC01", "usuario": "JOKAI\\ana", "fecha": "...", "estado": "EN_PROCESO" }
  ],
  "ejecuciones": [
    { "id": "...", "solicitud_id": "...", "estado": "FALLIDA",
      "inicio": "...", "fin": "...", "codigo": 1, "mensaje": "file not found" }
  ]
}
```

## Ciclo de vida de estados

```
SOLICITADA ─▶ EN_PROCESO ─▶ EXITOSA
    │              │
    ▼              ▼
CANCELADA      FALLIDA
```

- `SOLICITADA → EN_PROCESO`: el agente de la máquina toma el comando (take con
  `.tomando`).
- `EN_PROCESO → EXITOSA | FALLIDA`: el agente reporta el resultado como archivo `.st`.
- `SOLICITADA/EN_PROCESO → CANCELADA`: comando de cancelación (futuro; ver
  research.md §6, no es HU de esta rama salvo el estado en el snapshot).
- **Huérfana**: agente inactivo > umbral (5 min defecto) o ejecución > timeout
  (4h defecto) → el consolidador la marca `FALLIDA` con mensaje de agente inalcanzable.

## Snapshot: reglas de interpretación para el cliente

- El `estado` de una Ejecucion se toma como está (el consolidador es quien lo
  normaliza); el cliente **no infiere** estados.
- El cliente ordena ejecuciones por `solicitud.fecha` descendente para la lista.
- Ausencia de `estado.json` → `ErrorUsuario` con sugerencia de que el consolidador no
  ha publicado el primer snapshot (RF-010), no cierre abrupto.
- Malformado o `version_formato` mayor que la soportada → aviso de versión y error
  claro (CE-005).

## Configuración `[protocolo]` (nueva sección en config.toml)

| Clave | Defecto | Uso |
|---|---|---|
| `timeout_ejecucion` | `4h` | duración máxima de una ejecución (huérfana) |
| `latido_agente` | `60s` | intervalo esperado de heartbeat del agente |
| `umbral_agente_inactivo` | `5min` | sin latido en este lapso → máquina/agente caído |
| `retencion_archivos` | `720h` (30d) | purga de `procesados/` por el consolidador |
| `retencion_historial` | `2160h` (90d) | purga de historial en BD por el consolidador |
| `snapshot` | `estado.json` (relativo a raíz) | nombre del archivo de snapshot |

Estos valores los usa el **consolidador** (jokai-server); el cliente los declara solo
para conocerse y publicarlos coherentemente (config única por puesto). Las rutas son
siempre literales TOML (UNC).