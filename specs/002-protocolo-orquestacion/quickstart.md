# Quickstart: Protocolo de orquestación (lado cliente)

**Rama**: `002-protocolo-orquestacion` | **Fecha**: 2026-09-22

Validación manual del lado **jokai-client** sin jokai-server real: se simula el
consolidador escribiendo el snapshot y un agente-leído del inbox a mano.

## Prerrequisitos

- Binario `bin\jokai-client.exe` (o `go run ./cmd/jokai`).
- Config local (véase `config.example.toml` + nueva sección `[protocolo]`).
- Carpetas del fileserver simuladas localmente.

## Preparar fileserver de prueba (simulado)

```powershell
$raiz = "$env:TEMP\jokai-protocolo"
Remove-Item -Recurse -Force $raiz -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force "$raiz\inbox\BOTPC01" | Out-Null
New-Item -ItemType Directory -Force "$raiz\estado" | Out-Null
New-Item -ItemType Directory -Force "$raiz\procesados" | Out-Null
New-Item -ItemType Directory -Force "$raiz\bots-prod\facturacion" | Out-Null
Set-Content "$raiz\bots-prod\facturacion\main.py" 'print("hola")'
```

## Config de prueba

```toml
[fileserver]
raiz = 'C:\...\jokai-protocolo'
db   = 'jokai.db'

[descubrimiento]
archivo_entrada    = 'main.py'
profundidad_maxima = 5

[ambientes.dev]
ruta_bots = 'C:\...\jokai-protocolo\bots-dev'

[ambientes.prod]
ruta_bots = 'C:\...\jokai-protocolo\bots-prod'

[protocolo]
snapshot               = 'estado.json'
timeout_ejecucion      = '4h'
latido_agente          = '60s'
umbral_agente_inactivo = '5min'
retencion_archivos     = '720h'
retencion_historial    = '2160h'
```

## Ejecutar

```powershell
go run ./cmd/jokai --config config.toml
```

## Verificar (criterios del spec 002)

| # | Verificación | Resultado esperado |
|---|---|---|
| 1 | Arranque sin snapshot | Lista de ejecuciones muestra aviso en español: consolidador sin primer snapshot (RF-010); app viva, Esc/q funcionan |
| 2 | Snapshot presente | El TUI muestra maquinas con sus bots (asignación) y ejecuciones listadas |
| 3 | Publicar solicitud (Enter sobre bot → solicitar) | Aparece `inbox/BOTPC01/<uuid>.cmd` con `version_formato:1`, `maquina` = BOTPC01, usuario y fecha; TUI muestra el identificador de seguimiento |
| 4 | Estado simulado | El usuario escribe `estado.json` con `estado:"EN_PROCESO"` y refresca (F5) → la pantalla refleja EN_PROCESO |
| 5 | Sin máquina asignada | Bot no declarado en `maquinas[].bots` → solicitar da error claro: bot no asignado a máquina |
| 6 | Publicación atómica | Se simula un `.cmd.tmp` incompleto → el TUI/agente no lo considera (solo el `.cmd` final) |

## Simular consolidador / agente (a mano, sin jokai-server)

```powershell
# 1. snapshot inicial
$snap = @{
  version_formato = 1
  generado_en     = (Get-Date).ToUniversalTime().ToString('o')
  maquinas        = @(@{ nombre='BOTPC01'; ambiente='prod'; ultimo_latido=$null; bots=@('facturacion') })
  solicitudes     = @()
  ejecuciones     = @()
} | ConvertTo-Json -Depth 5
Set-Content "$raiz\estado.json" $snap

# 2. simular que un agente toma el comando y reporta EN_PROCESO
$solicitud = Get-Content "$raiz\inbox\BOTPC01\*.cmd" -Raw | ConvertFrom-Json
$ej = @{ id=(New-Guid); solicitud_id=$solicitud.id; estado='EN_PROCESO'; inicio=(Get-Date).ToUniversalTime().ToString('o'); fin=$null; codigo=$null; mensaje=$null } | ConvertTo-Json -Depth 3
# actualizar el snapshot: agregar solicitud est a y la ejecucion; volver a Set-Content
```

## Verificación automatizada

```powershell
go build ./... ; go vet ./... ; go test ./...
```

Los tres deben terminar sin errores (constitución VI). Los tests del protocolo corren
contra carpetas temporales con esta misma topología.