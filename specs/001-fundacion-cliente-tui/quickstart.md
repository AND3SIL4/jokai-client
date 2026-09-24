# Quickstart: Fundación del cliente TUI

**Rama**: `001-fundacion-cliente-tui` | **Fecha**: 2026-09-22

Validación manual de la funcionalidad en menos de 5 minutos, sin fileserver real
(carpetas locales).

## Preparar ambiente de prueba

```powershell
# 1. Estructura de bots falsa (dev)
$raiz = "$env:TEMP\jokai-quickstart"
New-Item -ItemType Directory -Force "$raiz\dev\facturacion"  | Out-Null
New-Item -ItemType Directory -Force "$raiz\dev\inventario"   | Out-Null
New-Item -ItemType Directory -Force "$raiz\dev\sin-main"    | Out-Null   # debe ignorarse
New-Item -ItemType Directory -Force "$raiz\dev\inventario\src" | Out-Null
Set-Content "$raiz\dev\facturacion\main.py"  'print("factura")'   # nivel 1: gana
Set-Content "$raiz\dev\facturacion\main.py.bak" 'x'                # no cuenta
Set-Content "$raiz\dev\inventario\src\main.py" 'print("stock")'   # nivel 2: válido
New-Item -ItemType Directory -Force "$raiz\prod\nomina" | Out-Null
Set-Content "$raiz\prod\nomina\main.py" 'print("nomina")'
```

## Crear config de prueba

`config.toml` (junto al exe o pasada con `--config`):

```toml
[fileserver]
raiz = 'C:\...\jokai-quickstart'   # usa $raiz real
db   = 'jokai.db'

[ambientes.dev]
ruta_bots = 'C:\...\jokai-quickstart\dev'

[ambientes.prod]
ruta_bots = 'C:\...\jokai-quickstart\prod'
```

## Ejecutar

```powershell
go run ./cmd/jokai --config config.toml
```

## Verificar (criterios del spec)

| # | Verificación | Resultado esperado |
|---|---|---|
| 1 | Arrancar con config válida | Header muestra usuario de Windows + "sin mapear (solo lectura)" + ambiente dev |
| 2 | Catálogo dev | Spinner y luego: `facturacion`, `inventario`; **sin** `sin-main` |
| 3 | Detalle | Enter sobre `inventario` → nombre, ambiente dev, carpeta y `...\src\main.py` |
| 4 | Conmutar ambiente | Tab → ambiente prod → catálogo muestra `nomina` (escaneo en segundo plano) |
| 5 | Refresco | F5 reescanea con spinner |
| 6 | Config ausente | Renombrar config y arrancar → pantalla de error en español con ruta buscada y sugerencia; app responde (Esc/q funcionan) |
| 7 | Ambiente inexistente | `ruta_bots` apuntando a carpeta borrada → error visible en el listado, header sigue operativo |
| 8 | Salir | q → salida limpia, código 0 |

## Verificación automatizada

```powershell
go build ./... ; go vet ./... ; go test ./...
```

Los tres deben terminar sin errores (constitución VI).
