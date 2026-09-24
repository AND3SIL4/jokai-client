# Contrato: Protocolo de orquestación vía fileserver

**Rama**: `002-protocolo-orquestacion` | **Versión de formato**: 1 | **Fecha**: 2026-09-22

> Este contrato es **compartido** entre `jokai-client` (TUI Go) y `jokai-server`
> (agentes Python + consolidador). Cualquier cambio **incompatible** incrementa
> `version_formato` (y `user_version` de la BD) y se documenta aquí antes de tocar
> código (Gobernanza de la constitución).

## 1. Componentes y fronteras

| Componente | Repo | Escribe | Lee |
|---|---|---|---|
| TUI (cliente) | jokai-client | `inbox/<maquina>/*.cmd` | `estado.json` (snapshot) |
| Agente por máquina | jokai-server | `estado/*.st`, `botasignaciones` (config) | su `inbox/` |
| Consolidador | jokai-server | BD SQLite, `estado.json`, purgas | `inbox/`, `estado/` |

La BD SQLite (fuente de verdad) **solo** la escribe el consolidador (constitución III).

## 2. Layout del fileserver

```
raiz_jokai/
├── inbox/
│   └── <maquina>/          # comandos dirigidos a esa máquina
│       ├── <uuid>.cmd            # comando de ejecución
│       ├── <uuid>.cancelar.cmd   # (reservado 006; misma atomicidad)
│       └── <uuid>.cmd.tomando    # tomo del agente (take en curso)
├── estado/
│   ├── <uuid>.st                 # reporte de estado (id = uuid de la ejecución)
│   └── <uuid>.st.tomando         # tomo del consolidador
├── procesados/YYYY/MM/DD/        # archivos drenados (retención 30d, purga consolidador)
└── estado.json                   # SNAPSHOT publicado por rename por el consolidador
```

## 3. Comando (`<uuid>.cmd`) — JSON

```json
{
  "version_formato": 1,
  "id": "3f2a...uuid...",
  "bot_nombre": "facturacion",
  "bot_ambiente": "prod",
  "bot_ruta_entrada": "\\\\FILESERVER\\bots-prod\\facturacion\\main.py",
  "maquina": "BOTPC01",
  "usuario": "JOKAI\\ana",
  "fecha": "2026-09-22T14:00:00Z"
}
```

### Publicación atómica (regla obligatoria)

1. Escribir `<uuid>.cmd.tmp` (misma carpeta).
2. `rename` → `<uuid>.cmd`.
3. Los lectores **solo** consideran el nombre final `.cmd`.

Los agentes no leen `.tmp`; jamás se abre/trunca el `.cmd` final.

## 4. Reporte de estado (`<uuid>.st`) — JSON

```json
{
  "version_formato": 1,
  "id": "<uuid de la ejecución>",
  "solicitud_id": "<uuid de la solicitud>",
  "estado": "EN_PROCESO | EXITOSA | FALLIDA",
  "inicio": "2026-09-22T14:00:01Z",
  "fin": null,
  "codigo": null,
  "mensaje": null
}
```

- El agente publica estados **intermedios** (`EN_PROCESO`) y **finales** (
  `EXITOSA` con `codigo`/`fin`, `FALLIDA` con `mensaje`).
- Misma atomicidad que §3 (`.tmp` → rename) y take con `.tomando`.

## 5. Snapshot (`estado.json`) — JSON (lectura del TUI)

Publicado por el **consolidador** con rename atómico (jamás a medias).

```json
{
  "version_formato": 1,
  "generado_en": "2026-09-22T14:00:00Z",
  "maquinas": [
    { "nombre": "BOTPC01", "ambiente": "prod",
      "ultimo_latido": "2026-09-22T13:59:00Z",
      "bots": ["facturacion", "nomina"] }
  ],
  "solicitudes": [
    { "id": "...", "bot_nombre": "facturacion", "bot_ambiente": "prod",
      "maquina": "BOTPC01", "usuario": "JOKAI\\ana", "fecha": "2026-09-22T14:00:00Z",
      "estado": "SOLICITADA" }
  ],
  "ejecuciones": [
    { "id": "...", "solicitud_id": "...", "estado": "EN_PROCESO",
      "inicio": "...", "fin": null, "codigo": null, "mensaje": null }
  ]
}
```

- `solicitudes[].estado` y `ejecuciones[].estado` ∈ `{SOLICITADA, EN_PROCESO,
  EXITOSA, FALLIDA, CANCELADA}`.
- Orden de `ejecuciones`: cualquier (el TUI ordena por `fecha` de la solicitud).
- Ausencia de `estado.json` → el TUI informa que el consolidador no ha publicado aún
  (RF-010) con sugerencia.

## 6. Asignación bot→máquina (decisión /clarify: opción A)

- Cada agente declara por config qué carpetas de bots maneja (su `bots`).
- El consolidador agrega eso al snapshot (`maquinas[].bots`).
- El TUI, al solicitar, resuelve la máquina destino buscando el bot en `maquinas` del
  mismo `ambiente`; si ninguna lo reclama → error claro con sugerencia (bot no
  asignado a máquina).

## 7. Versionado del protocolo

- `version_formato` entero por archivo. Reglas:
  - Cambio **aditivo** (nuevo campo opcional) → mismo número.
  - Cambio **incompatible** → `+1`; los lectores aceptan `version <= la suya`.
  - `version_formato > la suya` en un archivo → aviso explícito, no crash.
- La BD usa `PRAGMA user_version` aparte; el esquema vive en
  [esquema-sqlite.md](./esquema-sqlite.md) y se mantiene en sincronía por el
  consolidador.

## 8. Concurrencia y atomicidad (HU4)

- Nombre único por construcción: UUID v4 (minúsculas) del lado del escritor.
- Publicación: `O_CREATE|O_EXCL` sobre `.tmp` + `rename` al nombre final.
- Take (una vez): el consumidor renombra `.cmd` → `.cmd.tomando` (o `.st` →
  `.st.tomando`) antes de procesar; si el nombre final vuelve a aparecer, otro proceso
  puede tomarlo, pero el `.tomando` evita duplicados mientras esté presente.
- Huérfano: `.tomando` con antigüedad > `umbral_agente_inactivo` (5 min defecto) y sin
  progreso → el consolidador lo reclama/aísla (ver §9).
- Estado final publicado → el consolidador mueve `.cmd` + `.st` a `procesados/`.

## 9. Timeouts y huérfanos (defectos configurables)

| Parámetro | Defecto | Uso |
|---|---|---|
| `timeout_ejecucion` | `4h` | sin fin en este lapso → FALLIDA (mensaje de timeout) |
| `latido_agente` | `60s` | frecuencia esperada de heartbeat en `maquinas.ultimo_latido` |
| `umbral_agente_inactivo` | `5min` | sin latido → agente caído; ejecuciones → FALLIDA (agente inalcanzable) |

## 10. Retención (defectos configurables, las purga el consolidador)

| Parámetro | Defecto |
|---|---|
| `retencion_archivos` | `720h` (30 días) sobre `procesados/YYYY/MM/DD/` |
| `retencion_historial` | `2160h` (90 días) sobre historial en BD |

Los archivos drenados se mueven completos (nunca se borran sobre la marcha): el
movimiento a `procesados/` es la señal de que el consolidador ya los materializó.

## 11. Ratificación (2026-09-23, T019)

Sin cambios de formato (`version_formato` sigue en 1). Aclaraciones de documento:
take unificado `.cmd.tomando`/`.st.tomando` (§2 = §8), el `.st` se nombra con el `id`
de la ejecución (§2/§4) y `retencion_archivos` corregida a `720h` = 30 días (§10).
El ciclo completo queda fijado por el test de integración
`internal/adapters/driven/filequeue/integracion_test.go` (T020).