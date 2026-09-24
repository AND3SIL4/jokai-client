---
description: "Lista de tareas para implementar el protocolo de orquestación"
---

# Tareas: Protocolo de orquestación vía fileserver

**Entrada**: Documentos de diseño de `/specs/002-protocolo-orquestacion/`

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

- [ ] T001 Ampliar `config.example.toml` con la sección `[protocolo]` (snapshot,
      timeout_ejecucion, latido_agente, umbral_agente_inactivo, retencion_archivos,
      retencion_historial) con los defectos del data-model.md §Configuración, y los
      comentarios de uso
- [ ] T002 [P] Crear `internal/domain/protocolo.go`: constantes del protocolo
      (`VersionFormatoActual = 1`, nombres de carpetas `inbox/estado/procesados`,
      sufijos `.cmd`, `.st`, `.tomando`, `.tmp`) según contracts/protocolo-fileserver.md — dominio puro, sin deps externas

---

## Fase 2: Fundacional (prerrequisitos bloqueantes)

**⚠️ CRÍTICO**: Ninguna historia puede comenzar antes de completar esta fase.

- [ ] T003 `internal/domain/solicitud.go`: añadir campo `Maquina` (no vacía) y
      `VersionFormato` (>= 1); nueva función constructora que valide: id UUID v4,
      `maquina` no vacía, `usuario` no vacío, `version_formato >= 1`
      (data-model.md §Solicitud)
- [ ] T004 [P] Crear `internal/domain/ejecucion.go`: tipo `Ejecucion` con los campos
      del data-model.md §Ejecucion y tipo `EstadoEjecucion` con constantes
      `SOLICITADA | EN_PROCESO | EXITOSA | FALLIDA | CANCELADA` + etiquetas en
      español + función `EsFinal()`
- [ ] T005 [P] Ampliar `internal/domain/config.go`: struct `ProtocoloConfig`
      (snapshot, timeout_ejecucion, latido_agente, umbral_agente_inactivo,
      retencion_archivos, retencion_historial) con defectos `4h / 60s / 5min /
      720h / 2160h` y snapshot `estado.json` relativo sin `..`; añadir
      `Protocolo ProtocoloConfig` a `Configuracion` y su parámetro + validaciones
      (duraciones > 0) en `NuevaConfiguracion`
- [ ] T006 [P] Ampliar `internal/adapters/driven/config/toml.go`: parsear la sección
      `[protocolo]` aplicando defectos y validando cada campo con nombre de campo
      en el `ErrorUsuario` (patrón T009 de 001)
- [ ] T007 Ampliar `internal/application/puertos.go`: interfaces `LectorSnapshot`
      (`Leer(ctx) (domain.Snapshot, error)`) y `GeneradorIdentificadores`
      (`NuevoID() (string, error)`); añadir `Snapshot` como tipo de dominio con las
      proyecciones `Maquinas`, `Solicitudes`, `Ejecuciones` del data-model.md §Snapshot

**Punto de control**: `go build ./...` + `go vet ./...` + `go test ./...` verdes con
los stubs; fundación lista para las historias.

---

## Fase 3: Historia de Usuario 1 - Solicitar una ejecución (P1) 🎯 MVP

**Objetivo**: Un usuario con permisos solicita un bot; se publica el comando en
`inbox/<maquina>/` con UUID y el TUI muestra el identificador de seguimiento.
**Prueba independiente**: Entrar en un bot, solicitar su ejecución y verificar que
aparece `inbox/<maquina>/<uuid>.cmd` con todos los datos del contrato.

### Pruebas HU1 (escribir primero, deben fallar)

- [ ] T008 [P] [HU1] `internal/adapters/driven/filequeue/emisor_test.go`: con fake de
      carpeta temporal de la topología `inbox/<maquina>/` verifica que
      `SolicitarEjecucion` produce `<uuid>.cmd` válido, ignora `.tmp` a medio escribir,
      respeta `maquina` y devuelve `ErrorUsuario` claro si `inbox/<maquina>/` no existe
- [ ] T009 [P] [HU1] `internal/application/casosdeuso/solicitar_ejecucion_test.go`:
      con fakes `EmisorComandos`, `LectorSnapshot`, `GeneradorIdentificadores` y
      `Reloj`; casos: rol sin permiso → error claro (RF-002, spec AC2), bot no
      asignado a máquina → error claro, éxito → emisor llamado con solicitud completa
      (maquina resuelta de `Maquinas[].bots`, VersionFormato = 1)

### Implementación HU1

- [ ] T010 [P] [HU1] Crear `internal/adapters/driven/identificadores/uuid.go`:
      `GeneradorIdentificadores` real con `crypto/rand` (UUID v4 minúsculas, sin
      librerías externas)
- [ ] T011 [P] [HU1] Reemplazar el stub de `internal/adapters/driven/filequeue/emisor.go`:
      publicación atómica RF-001 (escribir `<uuid>.cmd.tmp` en la misma carpeta →
      `rename` a `<uuid>.cmd`; nunca abrir/truncar el final), carpeta `inbox/<maquina>/`
      creada si falta, `ErrorUsuario` con Causa+Sugerencia si la raíz de `inbox` no existe
- [ ] T012 [HU1] `internal/application/casosdeuso/solicitar_ejecucion.go`: caso de uso
      que valida `Rol.SoloLectura()`, resuelve la máquina destino desde `LectorSnapshot`
      (maquinas del ambiente del bot cuyo `bots` contenga el bot; ninguna → error con
      sugerencia), genera `id` con `GeneradorIdentificadores`, fija `VersionFormato` y
      `Fecha` con `Reloj`, y llama a `EmisorComandos` — devuelve la solicitud para
      mostrar el id de seguimiento

**Punto de control**: HU1 funcional y testeable de forma independiente (los tests usan
fakes del lector de snapshot; el lector real llega en HU2).

---

## Fase 4: Historia de Usuario 2 - Seguir el estado de una ejecución (P1)

**Objetivo**: Listar ejecuciones (recientes y por bot) leyendo el snapshot y ver el
detalle de cada una con estado y mensajes.
**Prueba independiente**: Colocar un `estado.json` manual y verificar que la pantalla
refleja estados/mensajes para el id correcto; sin snapshot → aviso (RF-010).

### Pruebas HU2 (escribir primero, deben fallar)

- [ ] T013 [P] [HU2] `internal/adapters/driven/snapshotlector/lector_test.go`: con
      carpetas temporales; archivo ausente → `ErrorUsuario` con sugerencia (RF-010),
      TOML/JSON malformado → error claro con Causa+Sugerencia, `version_formato`
      mayor a la soportada → aviso CE-005 (error claro, no crash), válido → mapea
      `domain.Snapshot` correcto
- [ ] T014 [P] [HU2] `internal/application/casosdeuso/listar_ejecuciones_test.go` y
      `ver_ejecucion_test.go`: con fake `LectorSnapshot`; orden por `solicitud.fecha`
      descendente, filtro por bot, detalle por id de ejecución, id inexistente → error
      claro

### Implementación HU2

- [ ] T015 [HU2] Crear `internal/adapters/driven/snapshotlector/lector.go`:
      implementación real de `LectorSnapshot` que lee `estado.json` (ruta desde
      `config.Protocolo.Snapshot` relativa a la raíz), valida `version_formato` y
      mapea a `domain.Snapshot`; ausencia y malformado → `ErrorUsuario`
- [ ] T016 [P] [HU2] `internal/application/casosdeuso/listar_ejecuciones.go`:
      lee el snapshot y devuelve las ejecuciones ordenadas con su solicitud asociada
- [ ] T017 [P] [HU2] `internal/application/casosdeuso/ver_ejecucion.go`:
      devuelve el detalle de una ejecución por id (con solicitud y mensajes)
- [ ] T018 [HU2] Ampliar `internal/adapters/driving/tui/` (modelo.go, pantallas.go,
      teclas.go): pantalla Lista Ejecuciones + Detalle Ejecución (estado, fechas,
      código, mensaje), refresco con F5/intervalo, aviso len cuando no hay snapshot,
      y **wire** real de `LectorSnapshot` en `cmd/jokai/main.go`

**Punto de control**: HU2 funcional; HU1+HU2 visibles en el TUI real.

---

## Fase 5: Historia de Usuario 3 - La máquina ejecutora procesa comandos (P1)

**Contrato solo**: el agente Python y el consolidador los implementa `jokai-server`
(otro repo); en esta rama se ratifica el contrato y se prueba el ciclo a nivel de
archivos para dejar el contrato fijado (spec.md §HU3 AC1-AC4).

- [ ] T019 [P] [HU3] Ratificar `contracts/protocolo-fileserver.md` y
      `contracts/esquema-sqlite.md` contra spec.md (formatos .cmd/.st, snapshot,
      layout, atomicidad, TTL, timeouts y retención): corregir discrepancias y dejar
      el contrato cerrado (Gobernanza: incrementa `version_formato` si cambia algo
      incompatible)
- [ ] T020 [HU3] Test de integración del ciclo en `internal/adapters/driven/
      filequeue/integracion_test.go`: con archivos temporales, un agente simulado del
      test toma el `.cmd` (maybe-busy: rename a `.tomando` de prueba) y publica un
      `.st`, un consolidador de prueba actualiza el snapshot, y `LectorSnapshot` lo
      refleja — el contrato HU3 queda fijado sin jokai-server

**Punto de control**: el contrato permitirá construir jokai-server sin ambigüedades.

---

## Fase 6: Historia de Usuario 4 - Conflicto y concurrencia seguros (P2)

**Prueba independiente**: Lanzar varias solicitudes simultáneas y verificar que todas
quedan registradas con identificadores distintos (spec.md §HU4 AC1-AC3, lado cliente).

- [ ] T021 [HU4] Test de concurrencia en `internal/adapters/driven/filequeue/
      emisor_concurrencia_test.go`: N goroutines solicitando contra la misma raíz →
      N archivos `.cmd` con UUIDs distintos, ninguno pisado, todos legibles y válidos
- [ ] T022 [HU4] Test de atomicidad en el mismo archivo: un `.cmd.tmp` incompleto NO
      se considera publicación (solo el `.cmd` final); verificar el patrón
      O_CREATE|O_EXCL + rename del emisor (RF-001, spec.md §HU4 AC2)
- [ ] T023 [HU4] Revisión final de `filequeue/emisor.go`: confirmar que no existe
      ningún camino que abra/trunque el `.cmd` final o pise un UUID existente; si algo
      falla corregirlo antes de continuar (spec.md §HU4)

**Punto de control**: concurrencia validada para el lado cliente del protocolo.

---

## Fase 7: Pulido y verificación cruzada

- [ ] T024 Ampliar `README.md` (raíz) con la sección del protocolo de orquestación
      (flujo TUI→fileserver→jokai-server, enlaces a contracts/) y actualizar el árbol
      de estructura con los adaptadores nuevos
- [ ] T025 [P] Ejecutar `quickstart.md` de 002 completo (simulando consolidador a
      mano) y registrar el resultado
- [ ] T026 `go build ./...` + `go vet ./...` + `go test ./...` en verde
- [ ] T027 [P] Revisión de consistencia: `go list -deps ./internal/domain` y
      `./internal/application` contienen solo stdlib; mensajes/identificadores en
      español; contractos 002 sin cambios pendientes contra spec.md

---

## Dependencias y Orden de Ejecución

### Dependencias de fase

- **Fase 1**: sin dependencias — iniciar de inmediato
- **Fase 2**: tras Fase 1 — **BLOQUEA todas las historias**
- **HU1 (Fase 3)**: tras Fase 2 — necesita T007 (puerto `LectorSnapshot`), no el lector real
- **HU2 (Fase 4)**: tras Fase 2 — necesita T015 (lector real) y la pantalla del TUI
- **HU3 (Fase 5)**: tras Fase 2 — necesita el emisor real (T011) para el test de ciclo
- **HU4 (Fase 6)**: tras Fase 3 (el emisor real ya es el punto de concurrencia)
- **Fase 7**: tras todas las historias

### Dentro de cada historia

- Pruebas primero (fallan) → implementación → verificación local (build+vet+test)
- Dominio/puertos antes que adaptadores; adaptadores antes que TUI

### Oportunidades paralelas

- T001+T002 (Fase 1); T004+T005+T006+T007 (Fase 2)
- T008+T009 (HU1); T010+T011 (HU1, tras las pruebas)
- T013+T014 (HU2); T016+T017 (HU2)
- T024+T025+T027 (Fase 7)

---

## Estrategia de Implementación

1. Fase 1+2 → protocolo fundacional compilando
2. HU1 + fakes → **MVP**: solicitar una ejecución y ver su comando en el inbox
3. HU2 → seguimiento de estado desde el snapshot (pantallas reales)
4. HU3 → contrato cerrado + test de ciclo (habilita jokai-server en paralelo)
5. HU4 → concurrencia/atomicidad del lado cliente
6. Fase 7 → quickstart y verificación final

## Notas

- [P] = distintos archivos, sin dependencias entre tareas
- Commit por tarea o grupo lógico (cuando el usuario lo solicite)
- Detener en cualquier punto de control para validar la historia de forma independiente
- El agente Python y el consolidador viven en `jokai-server` (repo hermano): cualquier
  cambio al contrato exige incrementar `version_formato`/`user_version` (Gobernanza)