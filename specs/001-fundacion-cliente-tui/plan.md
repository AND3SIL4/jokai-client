# Plan de Implementación: Fundación del cliente TUI

**Rama**: `001-fundacion-cliente-tui` | **Fecha**: 2026-09-22 | **Spec**: [spec.md](./spec.md)

**Entrada**: Especificación de funcionalidad desde `/specs/001-fundacion-cliente-tui/spec.md`

## Resumen

El cliente TUI de jokai arranca, carga y valida su configuración TOML (fileserver, BD y
ambientes dev/prod), detecta el usuario de Windows, muestra un catálogo de bots descubiertos
por escaneo del fileserver (subcarpeta directa = bot, archivo de entrada recursivo) y ofrece
navegación por teclado con manejo de errores comprensible. Es la base hexagonal sobre la que
se montarán el protocolo de orquestación (002) y las acciones de ejecución (003+).

## Contexto Técnico

**Lenguaje/Versión**: Go 1.26 (toolchain instalado: go1.26.4 windows/amd64)

**Dependencias primarias**:
- `github.com/charmbracelet/bubbletea/v2` v2.0.9 (TUI, Elvish-style update loop)
- `github.com/charmbracelet/bubbles/v2` v2.2.1 (spinner, viewport, help, key bindings)
- `github.com/charmbracelet/lipgloss/v2` v2.0.6 (estilos)
- `github.com/pelletier/go-toml/v2` (config TOML)
- `modernc.org/sqlite` (solo lectura futura; se incluye el stub del lector, sin uso activo en 001)

**Almacenamiento**: Fileserver SMB para descubrimiento de bots (solo lectura de carpetas);
SQLite del fileserver NO se consulta en 001 (puerto definido, adaptador stub).

**Testing**: `testing` stdlib + fakes de puertos (sin frameworks externos; tabla-driven).

**Plataforma destino**: Windows 10/11 (rutas UNC soportadas en toda la config).

**Tipo de proyecto**: CLI/TUI de escritorio (monolito hexagonal, un binario `jokai-client.exe`).

**Metas de rendimiento**: arranque < 3s con config válida; UI nunca bloqueada durante escaneo
(CE-001, CE-002); escaneo de 100 carpetas en segundo plano.

**Restricciones**: sin CGO (driver SQLite puro Go); único escritor de BD (constitución III);
todo texto visible en español (constitución IV).

**Escala/Alcance**: decenas de usuarios; 10-500 bots por ambiente; v1 = este spec.

## Verificación de Constitución

*COMPUERTA: debe pasar antes de la investigación de Fase 0. Re-verificar tras el diseño de Fase 1.*

| Principio | Estado | Nota |
|---|---|---|
| I. Spec antes que código | OK | Este plan deriva del spec 001 aprobado |
| II. Hexagonal estricta | OK | Estructura abajo; dominio sin deps; composition root única |
| III. Único escritor SQLite | OK | 001 no escribe BD; adaptador sqlite es lector de solo lectura (stub en 001) |
| IV. Español | OK | UI, mensajes, specs y código en español |
| V. Distribución sin instalación | OK | `go build` produce exe autocontenido sin CGO |
| VI. Testabilidad | OK | Puertos + fakes; descubridor probado contra carpetas temporales |
| VII. YAGNI | OK | Solo lo del spec 001; puertos nombrados para evolucionar (002/005) |

## Estructura del Proyecto

### Documentación (esta funcionalidad)

```text
specs/001-fundacion-cliente-tui/
├── plan.md              # Este archivo
├── research.md          # Investigación Fase 0
├── data-model.md        # Entidades y modelo de datos
├── quickstart.md        # Validación manual rápida
├── contracts/           # Contratos internos de 001
│   ├── puertos.md       # Interfaces Go (puertos) de la aplicación
│   └── config.md        # Contrato del archivo config.toml
├── checklists/
│   └── requirements.md  # Checklist del spec
└── tasks.md             # Salida de /speckit.tasks
```

### Código Fuente (raíz del repositorio)

```text
cmd/
└── jokai/
    └── main.go               # composition root: config → adaptadores → casos de uso → TUI
internal/
├── domain/
│   ├── bot.go                # Bot, Ambiente, errores de dominio
│   ├── usuario.go            # Usuario, Rol, rol sin mapear
│   └── config.go             # Configuracion, AmbienteConfig, DescubrimientoConfig (tipos puros)
├── application/
│   ├── puertos.go            # DescubridorBots, LectorUsuarios, EmisorComandos, Reloj
│   └── casosdeuso/
│       ├── identificar_usuario.go   # HU1: usuario de Windows + rol
│       ├── descubrir_bots.go        # HU2: catálogo del ambiente activo
│       └── cambiar_ambiente.go      # HU1: conmutar ambiente con validación
├── adapters/
│   ├── driving/
│   │   └── tui/
│   │       ├── modelo.go     # Model raíz, Init/Update/View, estados de pantalla
│   │       ├── teclas.go     # key bindings por pantalla
│   │       ├── estilos.go    # lipgloss
│   │       ├── pantallas.go  # vista de listado, detalle, error
│   │       └── mensajes.go  # mensajes internos del update loop (resultados async)
│   └── driven/
│       ├── config/
│       │   └── toml.go       # carga + validación de config.toml (implementa cargador)
│       ├── fsdescubridor/
│       │   └── descubridor.go # escáner: subcarpetas directas + main.py recursivo
│       ├── sqlite/
│       │   └── lector.go      # stub lector (usuario sin mapear) — implementación real: 002/005
│       └── filequeue/
│           └── emisor.go     # stub emisor de comandos (error "no implementado") — real: 002/003
└── ...
config.example.toml
```

**Decisión de estructura**: monolito hexagonal de paquetes internos (`cmd`, `internal/domain`,
`internal/application`, `internal/adapters`). Un solo módulo Go, binario único. Los stubs
sqlite/filequeue existen para fijar los puertos ya en 001 (evita rewiring en 002/003) y
fallan explícitamente con errores de dominio, nunca con panics silenciosos.

## Diseño de la Solución

### Flujo de arranque (composition root)

1. `main.go` parsea `--config` (default `config.toml` junto al exe).
2. El cargador TOML lee y valida; ante fallo, la app entra en modo error (pantalla de error
   con causa+acción, HU4) — no `os.Exit` abrupto: el error se muestra por el TUI.
3. Se instancian adaptadores driven (fsdescubridor con config de descubrimiento; stubs
   sqlite/filequeue), casos de uso y el adaptador driving TUI con todo inyectado.
4. `identificar_usuario` detecta usuario Windows; el caso de uso consulta el puerto
   `LectorUsuarios` (stub → rol sin mapear).

### Descubrimiento (HU2)

- Caso de uso `descubrir_bots(ambiente)`: pide al puerto `DescubridorBots` el escaneo.
- Adaptador fsdescubridor: `os.ReadDir` de la ruta del ambiente → subcarpetas directas →
  por cada una, `filepath.WalkDir` hasta `profundidad_maxima` buscando `archivo_entrada`;
  el más superficial gana; sin entrada → omitido. Errores por carpeta NO abortan el escaneo
  (se recopilan y se muestran como advertencias en el resultado).
- Async: el TUI lanza el caso de uso en una goroutine (`tea.Cmd`) y recibe el resultado
  por mensaje; spinner visible mientras tanto (bubbles v2). Conmutar ambiente o refrescar
  cancela/reemplaza el escaneo en curso (el resultado viejo se descarta).

### Pantallas (driving)

- `ListadoBots`: header (usuario+rol, ambiente activo, estado escaneo), lista navegable,
  pie de atajos. `DetalleBot`: datos del bot. `Error`: mensaje+causa+sugerencia+reintentos.
- Ambiente se conmuta con Tab; refresco con F5/r; detalle con Enter; regreso con Esc;
  salida con q/Ctrl+C (estándar Bubble Tea).

## Seguimiento de Complejidad

> Sin violaciones de constitución que justificar. (Módulo único, sin repos extra, sin
> patterns fuera de puertos/adaptadores.)

## Fases

1. **Fase 0**: investigación técnica (research.md) — completa.
2. **Fase 1**: diseño (data-model.md, contracts/, quickstart.md) — completa.
3. **Fase 2**: tareas (tasks.md) — generadas tras este plan.
