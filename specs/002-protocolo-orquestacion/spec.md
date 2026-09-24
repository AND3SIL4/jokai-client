# Especificación de Funcionalidad: Protocolo de orquestación vía fileserver

**Rama de la funcionalidad**: `002-protocolo-orquestacion`

**Creada**: 2026-09-22

**Estado**: Aclarado (pendiente de `/speckit.plan`)

**Entrada**: Descripción del usuario: "El fileserver es el puente: los TUI insertan solicitudes y consultan estado; un agente por máquina de bots ejecuta los main.py y reporta; una BD SQLite central es la fuente de verdad con un único proceso escritor."

## Aclaraciones

### Sesión 2026-09-22

- Q: ¿Cómo lee el TUI el estado de ejecuciones (BD directa vs snapshot)? → A: Snapshot del consolidador: el consolidador publica un `estado.json` (copia materializada de la BD) con publicación atómica (rename); el TUI solo lee ese archivo y nunca abre la BD.
- Q: ¿Cómo se decide qué máquina ejecuta cada bot? → A: Config del agente: cada agente declara por config qué bots (carpetas) le corresponden; esa asignación se publica en el snapshot y el TUI la consulta para resolver la máquina destino al publicar.
- Q: ¿Quién crea la BD la primera vez? → A: El consolidador (jokai-server) crea la BD con `PRAGMA user_version` al primer arranque; el TUI detecta la ausencia del snapshot y muestra un error claro con sugerencia (nunca escribe la BD).
- Q: ¿Valores por defecto para ejecuciones huérfanas? → A: Timeout de ejecución 4h, latido del agente cada 60s, umbral de agente inactivo 5min (ejecución → FALLIDA con mensaje). Todos configurables.
- Q: ¿Retención de archivos procesados e historial? → A: Archivos drenados a `procesados/YYYY/MM/DD/` con retención 30 días; historial de ejecuciones en BD con retención 90 días. Ambos configurables.

## Escenarios de Usuario y Pruebas *(obligatorio)*

### Historia de Usuario 1 - Solicitar una ejecución (Prioridad: P1)

Un usuario con permisos solicita desde el TUI la ejecución de un bot del ambiente activo.
La solicitud queda registrada en el fileserver de inmediato, visible para la máquina
ejecutora correspondiente, y el usuario recibe confirmación con un identificador único de
seguimiento.

**Por qué esta prioridad**: Es el acto central de la orquestación; sin solicitud no hay
ejecución ni seguimiento.

**Prueba independiente**: Se puede probar solicitando un bot de prueba y verificando que
aparece el comando en el inbox del fileserver con todos sus datos.

**Escenarios de aceptación**:

1. **Dado** un usuario con permisos y un bot válido, **cuando** solicita su ejecución,
   **entonces** se publica un comando con identificador único, bot, ambiente, usuario y
   fecha en el inbox del fileserver, y el TUI muestra el identificador de seguimiento.
2. **Dado** un usuario sin permisos de ejecución, **cuando** intenta solicitar, **entonces**
   se le informa claramente que su rol no lo permite.
3. **Dado** una solicitud publicada, **cuando** el fileserver está lento o el comando no
   puede escribirse, **entonces** el usuario ve un error claro con reintento y la
   solicitud NO queda a medias (o se publica completa o no se publica).

---

### Historia de Usuario 2 - Seguir el estado de una ejecución (Prioridad: P1)

El usuario consulta el estado de las solicitudes: pendiente, en proceso, finalizada
(éxito/fallo) o cancelada, con fecha y mensajes relevantes. El TUI muestra esta
información sin que el usuario ingrese a la máquina de bots.

**Por qué esta prioridad**: Ver qué pasó con la ejecución es la mitad del valor de
orquestar; sin estado, la solicitud es un acto de fe.

**Prueba independiente**: Se puede probar simulando estados en el fileserver y
verificando que el TUI los refleja para el identificador correcto.

**Escenarios de aceptación**:

1. **Dado** una solicitud en curso, **cuando** el usuario consulta, **entonces** ve su
   estado actual y la última actualización con marca de tiempo.
2. **Dado** una ejecución finalizada con fallo, **cuando** el usuario consulta,
   **entonces** ve el estado de fallo junto con el mensaje de error reportado.
3. **Dado** varias solicitudes de distintos bots, **cuando** el usuario consulta,
   **entonces** puede distinguir cada una por bot, ambiente, usuario y fecha.

---

### Historia de Usuario 3 - La máquina ejecutora procesa comandos (Prioridad: P1)

*(Historia técnica — la implementa `jokai-server`, pero el contrato se define aquí.)*
Cada máquina de bots ejecuta un agente que vigila los comandos dirigidos a ella, ejecuta
el `main.py` del bot indicado y reporta el progreso y resultado como archivos de estado.
Un único proceso consolidador drena los comandos y estados y es el único que escribe la
BD.

**Por qué esta prioridad**: Sin agente no hay ejecución real; su contrato debe quedar
fijado antes de construirlo.

**Prueba independiente**: Se puede probar publicando un comando falso en el inbox y
verificando que un agente de prueba lo toma, reporta estado y el consolidador lo
materializa en la BD.

**Escenarios de aceptación**:

1. **Dado** un comando en el inbox, **cuando** el agente de la máquina destino hace su
   ciclo, **entonces** lo toma exactamente una vez (ningún agente lo duplica) y reporta
   su estado en progreso.
2. **Dado** un bot cuya ejecución termina, **cuando** el agente reporta, **entonces**
   publica el resultado con código y mensajes como archivo de estado.
3. **Dado** archivos de comando y estado, **cuando** el consolidador drena, **entonces**
   la BD refleja la solicitud y su ejecución de forma consultable por los TUI.
4. **Dado** el consolidador caído, **cuando** se recupera, **entonces** procesa lo
   acumulado sin perder comandos ni estados.

---

### Historia de Usuario 4 - Conflicto y concurrencia seguros (Prioridad: P2)

Varios TUI y varios agentes trabajan a la vez sin perder solicitudes ni estados: los
archivos se escriben atómicamente, los nombres son únicos por construcción y los
procesos que compiten por los mismos recursos lo resuelven sin intervención humana.

**Por qué esta prioridad**: Sin esto el sistema es frágil en uso real simultáneo, pero
el flujo base ya funciona con un solo usuario (Por eso P2 y no P1).

**Prueba independiente**: Se puede probar lanzando varias solicitudes simultáneas y
verificando que todas quedan registradas con identificadores distintos.

**Escenarios de aceptación**:

1. **Dado** dos TUI solicitando al mismo tiempo, **cuando** ambos publican, **entonces**
   ambas solicitudes existen con identificadores distintos y ninguno pisa al otro.
2. **Dado** un comando que se está escribiendo, **cuando** otro proceso lo lee, **entonces**
   nunca ve un archivo a medias (publicación atómica).
3. **Dado** archivos de estado llegando durante el drenaje, **cuando** el consolidador
   procesa, **entonces** cada archivo se procesa exactamente una vez.

---

### Casos Límite

- ¿Qué pasa si un comando se publica y la máquina destino está apagada o sin agente?
- ¿Cómo se detecta y registra un agente caído a mitad de ejecución (estado huérfano)?
- ¿Qué ocurre si el consolidador encuentra un archivo de comando/estado malformado?
- ¿Cómo se limpian los archivos ya drenados (crecimiento del fileserver)?
- ¿Qué pasa si dos máquinas ejecutan agentes y ambas creen ser el consolidador?
- ¿Cómo se versiona el esquema de la BD para que cliente y agentes de distinta versión
  coexitstan?
- ¿Qué pasa con un bot eliminado del fileserver con solicitudes pendientes?
- ¿Quién crea la BD la primera vez y con qué estructura inicial?
  → El consolidador la crea al primer arranque con `PRAGMA user_version`; el TUI solo
  verifica que el snapshot exista y, si no, muestra un error con sugerencia (RF-010).

## Requisitos *(obligatorio)*

### Requisitos Funcionales

- **RF-001**: Las solicitudes de ejecución DEBEN publicarse como archivos de comando en
  un inbox del fileserver, con nombre único por construcción y escritura atómica.
- **RF-002**: Un comando DEBE incluir: identificador único, bot (nombre+ambiente+ruta de
  entrada fotografiada), máquina destino, usuario solicitante, fecha y versión del
  formato.
- **RF-003**: Los agentes DEBEN reportar estado (en proceso/éxito/fallo/cancelado) y
  mensajes como archivos de estado por ejecución, también atómicos y únicos.
- **RF-004**: Exactamente un proceso consolidador DEBE ser el único escritor de la BD;
  ningún TUI ni agente de ejecución DEBE abrirla en modo escritura.
- **RF-005**: El consolidador DEBE publicar un snapshot (`estado.json`, copia
  materializada de la BD) con publicación atómica (rename), como mecanismo de lectura
  para los TUI; ningún TUI abre la BD directamente.
- **RF-006**: El TUI DEBE poder listar las ejecuciones (recientes y por bot) y mostrar
  su estado y mensajes leyendo el snapshot.
- **RF-007**: Los archivos drenados DEBEN procesarse exactamente una vez y los
  malformados DEBEN aislarse sin bloquear el flujo.
- **RF-008**: El esquema y los formatos de archivo DEBEN versionarse para coexistencia
  de versiones de cliente y agente.
- **RF-009**: La asignación bot→máquina la DEBE declarar cada agente por config
  (qué carpetas de bots le corresponden) y el snapshot DEBE publicar esa asignación
  para que el TUI resuelva la máquina destino al publicar un comando.
- **RF-010**: El consolidador DEBE crear la BD la primera vez (`PRAGMA user_version`
  como versionado); ante la ausencia del snapshot, el TUI DEBE mostrar un error claro
  con sugerencia.
- **RF-011**: Timeouts y retención DEBEN ser configurables con estos defectos:
  timeout de ejecución 4h, latido del agente 60s, umbral de agente inactivo 5min,
  retención de archivos procesados 30 días, retención de historial en BD 90 días.

### Entidades Clave

- **Solicitud**: pedido de ejecución (identificador, bot, ambiente, usuario, fecha,
  máquina destino).
- **Ejecución**: realización de una solicitud (estado, marcas de tiempo, mensajes,
  resultado).
- **Máquina/Agente**: máquina de bots con su agente (nombre, ambiente, latido).
- **Consolidador**: único escritor de la BD (designado; puede ser un agente primario).

## Criterios de Éxito *(obligatorio)*

### Resultados Medibles

- **CE-001**: Una solicitud publicada desde el TUI es visible para el agente destino en
  menos de su siguiente ciclo de vigilancia, sin intervención.
- **CE-002**: Cero solicitudes perdidas con 10 TUI solicitando en paralelo.
- **CE-003**: Cero corrupciones de la BD en operación continua; una caída del
  consolidador no pierde comandos ya publicados.
- **CE-004**: El usuario consulta el estado de una ejecución sin ingresar a la máquina
  de bots y lo obtiene en segundos.
- **CE-005**: Dos versiones adyacentes de cliente/agente coexisten sin romper el
  protocolo.

## Contexto técnico *(obligatorio)*

### Arquitectura

```
TUI (jokai-client, Go)  ── lee ──▶  snapshot (estado.json) publicado por el consolidador
    │                                   ▲
    │ escribe comandos (inbox/<máquina>)│ único escritor de la BD
    ▼                                   │
inbox/ y estado/ (fileserver SMB) ──▶   consolidador (jokai-server)
    ▲                                   ▼
    │ ciclos de vigilancia          BD SQLite (fuente de verdad)
agente por máquina de bots (Python)
```

- El TUI **nunca** abre la BD: solo lee el snapshot y escribe comandos.
- El consolidador: único escritor de la BD, publica el snapshot y drena comandos/estados.
- Los agentes: vigilan su `inbox/<máquina>`, ejecutan `main.py` y publican estados.

## Supuestos

- El fileserver es una carpeta compartida de red (no aloja procesos); el consolidador
  corre en una máquina Windows designada.
- Los agentes existen en cada máquina de bots (su implementación es `jokai-server`,
  repo hermano; este spec fija su contrato, no su código).
- Este cliente (jokai-client) implementa: publicación de comandos (adaptador real de
  `EmisorComandos`), lectura del snapshot (adaptador lector real) y pantallas de
  solicitudes/estado.
- La asignación bot→máquina la declaran los agentes por config (opción A de research.md);
  la resuelve el TUI consultando el snapshot. El bootstrap de la BD lo hace el
  consolidador.
