# Investigación Técnica: Fundación del cliente TUI

**Rama**: `001-fundacion-cliente-tui` | **Fecha**: 2026-09-22

## Pila verificada (Fase 0)

### Bubble Tea v2 — verificado contra el proxy de módulos Go

- `github.com/charmbracelet/bubbletea/v2` **v2.0.9** (estable; v2.0.0 liberada tras 9 RC).
- `github.com/charmbracelet/bubbles/v2` **v2.2.1** (spinner, viewport, help, keys).
- `github.com/charmbracelet/lipgloss/v2` **v2.0.6** (estilos).
- Cambios relevantes vs v1: ruta de importación `/v2`; el renderer de v2 redibuja solo lo
  dirty (menos parpadeo); eventos de mouse mejorados; API Model (Init/Update/View) igual
  en esencia — `tea.Cmd`, `tea.Msg`, `tea.Quit` sin cambios de concepto.
- Comandos async: una `tea.Cmd` lanza el caso de uso (que internamente usa goroutines) y
  devuelve un `tea.Msg` con el resultado; es el patrón canónico para escaneos en segundo
  plano. La UI nunca llama FS directamente.

### SQLite en Go sin CGO

- `modernc.org/sqlite`: traducción a Go puro del motor; sin CGO permite `go build` simple y
  cross-compile desde Windows sin gcc. Es la única opción viable aquí (requisito: exe
  autocontenido, constitución V).
- Para 001 no se abre la BD: el adaptador lector queda stub que devuelve "sin mapear".
  La apertura real (modo lectura, timeouts, `PRAGMA query_only`) se diseña en 002.
- Recordatorio clave (ya en constitución III): SQLite sobre SMB no garantiza locks; el
  cliente solo leerá; las escrituras las consolida un único proceso escritor.

### TOML

- `github.com/pelletier/go-toml/v2`: decodificación a structs con etiquetas `toml:...`,
  soporte de strings literales (rutas UNC con backslashes sin escapar), errores con línea
  y columna (útil para mensajes de HU4: "campo X en línea N").
- Alternativa `burntsushi/toml` descartada: go-toml/v2 tiene mejor reporte de posición de
  error y API de decodificación más moderna.

### Detección de usuario de Windows

- `os/user.Current()` en Windows devuelve el usuario de la sesión (usa la API de Windows,
  no variables de entorno, por lo que no es falseable con `set USERNAME` fácilmente).
- Fallback: variable `USERNAME` si `user.Current()` falla (servicios o contextos raros).
- El nombre corto (sin dominio) se muestra en el header; dominio+usuario completo se
  conserva para el mapeo futuro (spec 005: tabla usuarios).

### Escaneo de carpetas sobre SMB

- `os.ReadDir` + `filepath.WalkDir` funcionan sobre rutas UNC (`\\server\share\...`) y
  letras de red mapeadas; el costo por directory-open sobre SMB es el cuello de botella,
  no la API.
- Estrategia: profundidad limitada (`profundidad_maxima`, default 5) y `WalkDir` con corte
  por profundidad (`filepath.SkipDir` al superar el límite) para no caminar árboles
  completos de bots grandes (venv de Python pueden tener miles de archivos bajo `Lib/`).
- Determinismo: recorrer subcarpetas en orden alfabético; ante varios `main.py`, el
  primero encontrado en recorrido pre-orden (el más superficial) gana.
- Errores por subcarpeta (permisos, carpetas fantasma de SMB) no abortan el escaneo: se
  acumulan como advertencias por bot omitido y se muestran al final del listado.
- Datos de diseño medidos: abrir una carpeta SMB tarda ~5-50ms; 100 bots × 5 niveles
  puede tardar segundos → imprescindible async con spinner (CE-002).

### Convenciones Go del proyecto

- Module path: `jokai-client` (nombre corto; es una herramienta interna, sin hosting
  público previsto).
- Go 1.26 con `go.mod` estándar; tests con stdlib `testing` (tabla-driven), sin testify
  (YAGNI, constitución VII).
- Errores: `errors.New`/`fmt.Errorf` con `%w`; errores de dominio como valores
  sentinela/constructores (ej. `domain.ErrRutaInvalida`), mensajes de usuario en español.

## Riesgos y mitigaciones

| Riesgo | Mitigación |
|---|---|
| SMB lento congela el arranque | El escaneo solo ocurre tras mostrar la pantalla base, como tea.Cmd async |
| WalkDir descarga masiva en venv | Corte por profundidad + SkipDir; profundidad configurable |
| v2 de bubbles tiene menos ejemplos comunitarios | Componentes usados (spinner/help/keys) son estables en v2; fallback: implementar spinner a mano (~40 líneas) si algo no cuadra |
| Config con rutas relativas confusas | Validación: rutas de ambiente requeridas absolutas (UNC o letra); error claro si no (HU4) |

## Decisiones tomadas (resumen)

1. Bubble Tea **v2** (última estable, requisito del usuario de "última versión").
2. `modernc.org/sqlite` como único driver SQLite del cliente (sin CGO).
3. `pelletier/go-toml/v2` para config.
4. Usuario Windows vía `os/user` con fallback `USERNAME`.
5. Descubrimiento async por `tea.Cmd` con corte de profundidad y advertencias por bot.
6. Stubs de sqlite/filequeue incluidos desde 001 para fijar puertos (sin rewiring futuro).
