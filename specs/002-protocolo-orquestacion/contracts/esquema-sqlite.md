# Contrato: Esquema SQLite

**Rama**: `002-protocolo-orquestacion` | **user_version**: 1 | **Fecha**: 2026-09-22

> La BD la crea el **consolidador** (jokai-server) al primer arranque (RF-010); este
> documento fija el esquema como contrato. El cliente **nunca** escribe: ni archivos
> sobre esta BD. Cambios incompatibles → `user_version + 1` y actualizar este archivo.

## Initialización

```sql
PRAGMA journal_mode = DELETE;      -- WAL no opera bien sobre SMB
PRAGMA synchronous = FULL;         -- durabilidad ante corte
PRAGMA foreign_keys = ON;
PRAGMA user_version = 1;           -- versionado del contrato
```

## Tablas

### solicitudes

| Columna | Tipo | Reglas |
|---|---|---|
| `id` | TEXT PRIMARY KEY | UUID v4 |
| `main_version_formato` | INTEGER NOT NULL | versión del comando origen |
| `bot_nombre` | TEXT NOT NULL | |
| `bot_ambiente` | TEXT NOT NULL CHECK (bot_ambiente IN ('dev','prod')) | |
| `bot_ruta_entrada` | TEXT NOT NULL | ruta fotografiada del main.py |
| `maquina` | TEXT NOT NULL | máquina destino |
| `usuario` | TEXT NOT NULL | `DOMINIO\usuario` |
| `fecha` | TEXT NOT NULL | ISO 8601 UTC |
| `estado` | TEXT NOT NULL CHECK (estado IN ('SOLICITADA','EN_PROCESO','EXITOSA','FALLIDA','CANCELADA')) | |

### ejecuciones

| Columna | Tipo | Reglas |
|---|---|---|
| `id` | TEXT PRIMARY KEY | UUID v4 |
| `solicitud_id` | TEXT NOT NULL REFERENCES solicitudes(id) | |
| `estado` | TEXT NOT NULL CHECK (... igual que solicitudes) | |
| `inicio` | TEXT | ISO 8601 UTC |
| `fin` | TEXT | ISO 8601 UTC |
| `codigo` | INTEGER | código de salida |
| `mensaje` | TEXT | mensaje/error |

`CREATE INDEX idx_ejecuciones_solicitud ON ejecuciones(solicitud_id);`
`CREATE INDEX idx_solicitudes_fecha ON solicitudes(fecha DESC);`

### usuarios

| Columna | Tipo | Reglas |
|---|---|---|
| `windows_completo` | TEXT PRIMARY KEY | `DOMINIO\usuario` |
| `rol` | TEXT NOT NULL CHECK (rol IN ('negocio','tecnico','soporte')) | |

### maquinas

| Columna | Tipo | Reglas |
|---|---|---|
| `nombre` | TEXT PRIMARY KEY | |
| `ambiente` | TEXT NOT NULL CHECK (ambiente IN ('dev','prod')) | |
| `ultimo_latido` | TEXT | ISO 8601 UTC |

### botasignaciones

| Columna | Tipo | Reglas |
|---|---|---|
| `maquina` | TEXT NOT NULL REFERENCES maquinas(nombre) | |
| `bot` | TEXT NOT NULL | carpeta del bot |
| PRIMARY KEY (maquina, bot) | | |

### horarios (reservado 006)

Tabla reservada para ejecuciones programadas; no se crea en user_version 1.

## Snapshot de la BD (para el TUI)

El snapshot `estado.json` (contracts/protocolo-fileserver.md §5) es una proyección de
`solicitudes` + `ejecuciones` + `maquinas` + `botasignaciones`. El consolidador lo
serializa y lo publica por rename.

## Observabilidad

`maquinas.ultimo_latido` alimenta el umbral de agente inactivo (5 min defecto) y las
ejecuciones huérfanas (timeout 4h) que el consolidador marca `FALLIDA` (ver §9 del
protocolo).