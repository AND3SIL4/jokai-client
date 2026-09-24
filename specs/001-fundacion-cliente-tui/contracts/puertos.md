# Contratos Internos: Fundación del cliente TUI

**Rama**: `001-fundacion-cliente-tui` | **Fecha**: 2026-09-22

Los contratos de 001 son internos de Go (puertos) y del archivo de configuración. El
contrato wire (archivos del protocolo fileserver) pertenece al spec 002.

## 1. Puertos de la aplicación (`internal/application/puertos.go`)

```go
// DescubridorBots enumera los bots válidos de un ambiente (driven, filesystem).
DescubridorBots interface {
    Descubrir(ctx context.Context, ambiente domain.Ambiente) ([]domain.Bot, []string, error)
    // []string = advertencias (bots omitidos, permisos, etc.), sin abortar el resultado.
}

// LectorUsuarios resuelve el rol de un usuario de Windows (driven, BD; stub en 001).
LectorUsuarios interface {
    RolDe(ctx context.Context, windowsCompleto string) (domain.Rol, error)
}

// EmisorComandos publica solicitudes de ejecución en el fileserver (driven; stub en 001).
EmisorComandos interface {
    SolicitarEjecucion(ctx context.Context, solicitud domain.Solicitud) error
}

// Reloj abstrae el tiempo (driven) para pruebas determinísticas.
Reloj interface {
    Ahora() time.Time
}
```

Reglas de contrato:
- Ningún puerto recibe o devuelve tipos de infraestructura (`*sql.DB`, `tea.*`, rutas de
  módulos externos). Solo tipos de `domain`.
- Los errores devueltos son de dominio (`internal/domain/errores.go`); la UI los traduce
  a mensajes (causa+sugerencia) — jamás muestra `err.Error()` crudo de infraestructura.
- `ctx context.Context` en toda operación que toca FS/BD (cancelable al conmutar ambiente).

## 2. Casos de uso expuestos (driving ports efectivos)

```go
// IdentificarUsuario: usuario de Windows actual + rol del mapeo.
IdentificarUsuario(ctx) (domain.Usuario, error)

// DescubrirBots: catálogo del ambiente indicado.
DescubrirBots(ctx, ambiente) (ResultadoDescubrimiento, error)
// ResultadoDescubrimiento{ Bots []domain.Bot; Advertencias []string; Ambiente domain.Ambiente }

// AmbienteActivo/ConmutarAmbiente: gestionados por el modelo TUI con domain.Ambiente;
// sin caso de uso dedicado en 001 (estado de UI puro).
```

## 3. Contrato del archivo `config.toml` (ver `config.example.toml`)

| Tabla | Campo | Tipo | Obligatorio | Validación |
|---|---|---|---|---|
| `fileserver` | `raiz` | string (literal) | sí | ruta absoluta (UNC `\\srv\...` o `X:\...`) |
| `fileserver` | `db` | string | sí | relativa a `raiz`, sin `..` |
| `descubrimiento` | `archivo_entrada` | string | no (default `main.py`) | nombre de archivo sin rutas |
| `descubrimiento` | `profundidad_maxima` | int | no (default 5) | >= 0; 0 = sin límite |
| `ambientes.dev` | `ruta_bots` | string | sí | ruta absoluta |
| `ambientes.prod` | `ruta_bots` | string | sí | ruta absoluta |
| `ui` | `intervalo_refresco` | string duración | no (default `2s`) | formato duración Go |

Errores de contrato (HU4): archivo ausente → mensaje con ruta buscada y sugerencia de
copiar `config.example.toml`; TOML inválido → línea/columna; campo inválido → tabla+campo
+regla violada. Todos con `accion sugerida` en español.

## 4. Teclas estándar de la UI (contrato de usuario)

| Tecla | Acción | Pantallas |
|---|---|---|
| ↑/↓ | navegar lista | ListadoBots |
| Tab | conmutar ambiente (dev↔prod) | ListadoBots, DetalleBot |
| F5 o r | refrescar catálogo | ListadoBots |
| Enter | abrir detalle del bot | ListadoBots |
| Esc | regresar | DetalleBot, Error |
| q / Ctrl+C | salir | todas |

El pie de pantalla (`bubbles/v2 help`) siempre refleja estas teclas por pantalla.
