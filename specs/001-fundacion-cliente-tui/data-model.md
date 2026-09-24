# Modelo de Datos: Fundación del cliente TUI

**Rama**: `001-fundacion-cliente-tui` | **Fecha**: 2026-09-22

## Entidades de dominio (`internal/domain`)

### Bot

| Campo | Tipo | Descripción |
|---|---|---|
| Nombre | string | Nombre de la subcarpeta (= nombre del bot) |
| Ambiente | Ambiente | Ambiente al que pertenece (`dev`/`prod`) |
| RutaRaiz | string | Ruta absoluta de la carpeta del bot |
| RutaEntrada | string | Ruta absoluta del archivo de entrada (main.py más superficial) |
| Advertencias | []string | Vacío en 001; reservado para notas del escaneo (p. ej. entradas omitidas por profundidad) |

Invariantes:
- `Nombre` no vacío; `RutaRaiz` absoluta; `RutaEntrada` está dentro de `RutaRaiz`.
- Constructor `NuevoBot(...)` valida y devuelve error de dominio (`ErrBotInvalido`).

### Ambiente

| Campo | Tipo | Descripción |
|---|---|---|
| Valor | string | `dev` o `prod` |

- Tipo de dominio inmutable con constructores `AmbienteDev()` / `AmbienteProd()`.
- `ParseAmbiente(string) (Ambiente, error)` para input de config; rechaza otros valores
  con `ErrAmbienteInvalido`.
- Exactamente un ambiente está activo en la aplicación (estado del modelo TUI, no del
  dominio).

### Usuario

| Campo | Tipo | Descripción |
|---|---|---|
| WindowsNombre | string | Usuario de Windows de la sesión (sin dominio) |
| WindowsCompleto | string | Dominio\usuario cuando aplica (para mapeo futuro en BD) |
| Rol | Rol | Rol resuelto; en 001 siempre "sin mapear" |

### Rol

Valores: `negocio`, `tecnico`, `soporte`, `sinmapear`. En 001 solo `sinmapear` se produce
(el lector real de la BD llega con 002/005). `sinmapear` se muestra como
"sin mapear (solo lectura)". Cada rol expone `Etiqueta()` en español para la UI y
`SoloLectura()` (true para `sinmapear`; base de permisos de 005).

### Configuracion (tipos puros; la carga es un adaptador)

| Sección TOML | Campos |
|---|---|
| `[fileserver]` | `raiz` (UNC/letra, absoluta), `db` (relativa a `raiz`) |
| `[descubrimiento]` | `archivo_entrada` (default `main.py`), `profundidad_maxima` (default 5, 0 = ilimitado) |
| `[ambientes.dev]` | `ruta_bots` (absoluta) |
| `[ambientes.prod]` | `ruta_bots` (absoluta) |
| `[ui]` | `intervalo_refresco` (duración, default `2s`) |

- La validación es de dominio: `domain.NuevaConfiguracion(...)` exige ambos ambientes con
  ruta absoluta, `raiz` absoluta, `db` relativa; errores con campo+causa (HU4).
- El adaptador TOML solo mapea bytes → struct cruda; el dominio valida.

## Mensajes internos del update loop (adaptador driving)

| Mensaje | Contenido | Origen |
|---|---|---|
| `msjBotsDescubiertos` | `[]domain.Bot`, advertencias, ambiente del resultado | tea.Cmd del escaneo |
| `msjErrorEscaneo` | error de dominio + ambiente | tea.Cmd del escaneo |
| `msjUsuarioIdentificado` | `domain.Usuario` | tea.Cmd de arranque |
| `msjTick` | — | temporizador de spinner |

Regla de consistencia: todo mensaje async lleva el ambiente al que corresponde; si el
ambiente activo cambió mientras tanto, el resultado se descarta (caso límite del spec).

## Almacenamiento en 001

- **Sin BD**: ningún caso de uso de 001 toca SQLite; el puerto `LectorUsuarios` existe y su
  adaptador stub responde `sinmapear` (contrato respetado, sin datos reales).
- **Sin escrituras en fileserver**: 001 es de solo lectura total sobre el fileserver
  (escaneo). El puerto `EmisorComandos` existe con stub que devuelve
  `domain.ErrNoImplementado` (se implementa en 002/003).

## Evolución prevista (fuera de 001)

- Esquema SQLite compartido y su `PRAGMA user_version` → spec 002 (protocolo).
- Tabla `usuarios` (WindowsCompleto → Rol) → spec 005.
- `Solicitud`/`Ejecucion` y estados → specs 002/003/004.
