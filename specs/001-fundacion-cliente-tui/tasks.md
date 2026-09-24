---
description: "Lista de tareas para implementar la fundación del cliente TUI"
---

# Tareas: Fundación del cliente TUI

**Entrada**: Documentos de diseño de `/specs/001-fundacion-cliente-tui/`

**Requisitos previos**: plan.md (requerido), spec.md (requerido), research.md, data-model.md, contracts/

**Pruebas**: Incluidas (la constitución VI exige verificación `go test ./...` en verde).

**Organización**: Tareas agrupadas por historia de usuario para permitir implementación y prueba independientes.

## Formato: `[ID] [P?] [HU] Descripción`

- **[P]**: Puede ejecutarse en paralelo (distintos archivos, sin dependencias)
- **[HU]**: Historia de usuario asociada (HU1-HU4)

## Convención de rutas

Proyecto único: `cmd/`, `internal/` en la raíz del repositorio (ver plan.md).

---

## Fase 1: Preparación (infraestructura compartida)

- [ ] T001 Crear `go.mod` (módulo `jokai-client`, Go 1.26) y traer dependencias:
      bubbletea/v2, bubbles/v2, lipgloss/v2, pelletier/go-toml/v2, modernc.org/sqlite
- [ ] T002 [P] Crear `.gitignore` (exe, `bin/`, `config.toml`) y `config.example.toml`
      según contracts/config.md
- [ ] T003 [P] Redactar `README.md` (visión, build, run, flujo SDD) y `AGENTS.md`
      (comandos y convenciones para agentes de código)

---

## Fase 2: Fundacional (prerrequisitos bloqueantes)

**⚠️ CRÍTICO**: Ninguna historia de usuario puede comenzar antes de completar esta fase.

- [ ] T004 `internal/domain/ambiente.go`: tipo `Ambiente` (dev/prod, parse, etiquetas)
      + `internal/domain/errores.go`: errores de dominio sentinelas
- [ ] T005 [P] `internal/domain/usuario.go`: `Usuario`, `Rol` (4 valores, `Etiqueta()`,
      `SoloLectura()`)
- [ ] T006 [P] `internal/domain/bot.go`: `Bot` + `NuevoBot` con validaciones
- [ ] T007 `internal/domain/config.go`: tipos `Configuracion` + `NuevaConfiguracion`
      con validación de dominio (ambientes obligatorios, rutas absolutas, db relativa)
- [ ] T008 `internal/application/puertos.go`: `DescubridorBots`, `LectorUsuarios`,
      `EmisorComandos`, `Reloj` según contracts/puertos.md
- [ ] T009 `internal/adapters/driven/config/toml.go`: carga TOML → struct cruda →
      `domain.NuevaConfiguracion`; errores con línea/columna (HU4)
- [ ] T010 `cmd/jokai/main.go`: composition root (parse `--config`, wire completo,
      arranque del TUI) — compila desde T004-T009

**Punto de control**: `go build ./...` verde con stubs; fundación lista para historias.

---

## Fase 3: Historia de Usuario 1 - Identidad y ambiente (P1) 🎯 MVP

**Objetivo**: Header con usuario de Windows y ambiente conmutable.
**Prueba independiente**: Arrancar y ver usuario+ambiente; Tab conmuta sin reiniciar.

### Pruebas HU1 (escribir primero, deben fallar)

- [ ] T011 [P] [HU1] `internal/application/casosdeuso/identificar_usuario_test.go`:
      con fake `LectorUsuarios` (rol mapeado y error) y stub (sin mapear)
- [ ] T012 [P] [HU1] `internal/domain/ambiente_test.go`: parse válido/inválido,
      etiquetas en español

### Implementación HU1

- [ ] T013 [HU1] `internal/application/casosdeuso/identificar_usuario.go`: detecta
      usuario Windows (os/user + fallback USERNAME) y consulta `LectorUsuarios`
- [ ] T014 [HU1] `internal/adapters/driven/sqlite/lector.go`: stub `LectorUsuarios`
      → `RolSinMapear` (marcado TODO 002/005)

**Punto de control**: HU1 funcional y testeable de forma independiente.

---

## Fase 4: Historia de Usuario 2 - Catálogo de bots (P1) 🎯 MVP

**Objetivo**: Listado async de bots por escaneo con spinner y refresco.
**Prueba independiente**: Carpeta de ambiente con subcarpetas válidas/inválidas → listado exacto.

### Pruebas HU2 (escribir primero, deben fallar)

- [ ] T015 [P] [HU2] `internal/adapters/driven/fsdescubridor/descubridor_test.go`:
      tabla: bot nivel 1, nivel N, sin entrada omitido, varios main.py gana superficial,
      error de carpeta → advertencia no error, profundidad máxima respetada
- [ ] T016 [P] [HU2] `internal/domain/bot_test.go`: `NuevoBot` valida invariantes

### Implementación HU2

- [ ] T017 [HU2] `internal/adapters/driven/fsdescubridor/descubridor.go`: ReadDir +
      WalkDir con corte de profundidad, más superficial gana, advertencias acumuladas
- [ ] T018 [HU2] `internal/application/casosdeuso/descubrir_bots.go`: valida ambiente,
      llama puerto, envuelve `ResultadoDescubrimiento`

**Punto de control**: HU2 funcional y testeable de forma independiente.

---

## Fase 5: Historia de Usuario 3 - Detalle y navegación (P2)

**Prueba independiente**: Enter abre detalle con datos del bot; Esc regresa conservando posición.

- [ ] T019 [HU3] `internal/adapters/driving/tui/teclas.go`: key bindings por pantalla
      (contracts/puertos.md §4)
- [ ] T020 [HU3] `internal/adapters/driving/tui/modelo.go`: modelo raíz, estados
      (listado/detalle/error), teclas Tab/F5/r/Enter/Esc/q, descarte de resultados async
      de ambiente cambiado (msj con ambiente)
- [ ] T021 [HU3] `internal/adapters/driving/tui/pantallas.go`: vistas ListadoBots,
      DetalleBot, Error con ayuda (bubbles help)
- [ ] T022 [P] [HU3] `internal/adapters/driving/tui/estilos.go`: lipgloss (header,
      lista, footer, errores)

**Punto de control**: HU3 navegable; HU1+HU2 visibles en TUI real.

---

## Fase 6: Historia de Usuario 4 - Errores comprensibles (P2)

**Prueba independiente**: Config inválida y ruta inexistente → mensajes español + app viva.

- [ ] T023 [HU4] Pantalla Error con causa+sugerencia+reintento (parte de pantallas.go);
      mapear errores de config (archivo ausente, TOML inválido, campo inválido)
- [ ] T024 [HU4] `internal/adapters/driven/filequeue/emisor.go`: stub `EmisorComandos`
      → `ErrNoImplementado` (fija el puerto para 002/003)

**Punto de control**: HU4 funcional; ningún cierre abrupto.

---

## Fase 7: Pulido y verificación cruzada

- [ ] T025 [P] Ejecutar quickstart.md completo y registrar resultado
- [ ] T026 `go build ./...` + `go vet ./...` + `go test ./...` en verde
- [ ] T027 [P] Revisión de consistencia: mensajes todos en español; sin deps de
      infraestructura en domain/application (import-lint manual)

---

## Dependencias y Orden de Ejecución

### Dependencias de fase

- **Fase 1**: sin dependencias — iniciar de inmediato
- **Fase 2**: tras Fase 1 — **BLOQUEA todas las historias**
- **HU1/HU2 (Fases 3-4)**: tras Fase 2; paralelizables entre sí
- **HU3 (Fase 5)**: tras Fase 2; integra HU1+HU2 en el TUI (necesita sus mensajes)
- **HU4 (Fase 6)**: tras Fase 5 (necesita pantallas para mostrar errores)
- **Fase 7**: tras todas las historias

### Dentro de cada historia

- Pruebas primero (fallan) → implementación → verificación local
- Dominio antes que adaptadores; adaptadores antes que TUI

### Oportunidades paralelas

- T002+T003 (Fase 1); T005+T006 (Fase 2); T011+T012 (HU1); T015+T016 (HU2); T022 (HU3);
  T027 (Fase 7)

---

## Estrategia de Implementación

1. Fase 1+2 → fundación compilando
2. HU1 + HU2 con sus tests → **MVP demostrable** (header + catálogo async)
3. HU3 → navegación completa
4. HU4 → robustez de errores
5. Fase 7 → verificación final y quickstart

## Notas

- [P] = distintos archivos, sin dependencias entre tareas
- Commit por tarea o grupo lógico (cuando el usuario lo solicite)
- Detener en cualquier punto de control para validar la historia de forma independiente
