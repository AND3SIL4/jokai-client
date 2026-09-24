# Especificación de Funcionalidad: Fundación del cliente TUI (jokai-client)

**Rama de la funcionalidad**: `001-fundacion-cliente-tui`

**Creada**: 2026-09-22

**Estado**: Borrador

**Entrada**: Descripción del usuario: "Cliente TUI de orquestación de bots: arranque, configuración de ambientes dev/prod, usuario de Windows, descubrimiento de bots por carpetas del fileserver y navegación base."

## Escenarios de Usuario y Pruebas *(obligatorio)*

### Historia de Usuario 1 - Identidad y ambiente visible al arrancar (Prioridad: P1)

Al abrir la aplicación, el usuario ve de inmediato quién es (usuario de Windows detectado
automáticamente, sin login) y en qué ambiente está trabajando (desarrollo o producción).
Puede conmutar entre ambientes en cualquier momento; el ambiente activo es siempre evidente
para evitar ejecutar acciones sobre el ambiente equivocado.

**Por qué esta prioridad**: Toda la interfaz y las acciones futuras dependen de saber
identidad y ambiente; sin esto nada más es seguro de usar.

**Prueba independiente**: Se puede probar por completo arrancando la aplicación con una
configuración válida y verificando que el encabezado muestra el usuario de Windows y el
ambiente activo, y que conmutar el ambiente actualiza la pantalla sin reiniciar.

**Escenarios de aceptación**:

1. **Dado** un PC con sesión de Windows iniciada, **cuando** el usuario abre la aplicación
   con una configuración válida, **entonces** el encabezado muestra su usuario de Windows y
   el ambiente activo por defecto.
2. **Dado** la aplicación abierta, **cuando** el usuario conmuta de ambiente, **entonces**
   el encabezado refleja el nuevo ambiente y la aplicación sigue funcionando sin reiniciar.
3. **Dado** un usuario de Windows sin mapeo de rol conocido, **cuando** abre la aplicación,
   **entonces** se le identifica por su usuario de Windows con rol "sin mapear (solo
   lectura)" indicado junto a su nombre.

---

### Historia de Usuario 2 - Catálogo de bots del ambiente (Prioridad: P1)

El usuario ve el listado de bots disponibles en el ambiente activo sin conocer rutas
técnicas: el sistema escanea la carpeta del ambiente en el fileserver, toma cada subcarpeta
directa como un bot (su nombre de carpeta es el nombre del bot) y ubica el punto de entrada
buscando el archivo ejecutable del bot de forma recursiva dentro de su carpeta. Las
subcarpetas que no contienen dicho archivo no se muestran. El escaneo ocurre en segundo
plano (la carpeta de red puede tardar) con indicador de progreso visible, y el usuario puede
refrescar el catálogo cuando lo necesite.

**Por qué esta prioridad**: Ver los bots disponibles es la razón de ser del catálogo; sin
listado no hay punto de partida para ninguna acción de orquestación.

**Prueba independiente**: Se puede probar creando una carpeta de ambiente con subcarpetas
(con y sin punto de entrada) y verificando que el listado muestra exactamente los bots
válidos con su nombre.

**Escenarios de aceptación**:

1. **Dado** una carpeta de ambiente con subcarpetas que contienen el archivo de entrada en
   distinta profundidad, **cuando** se completa el escaneo, **entonces** cada subcarpeta
   con entrada aparece como bot con el nombre de su carpeta y se muestra su ruta de
   entrada.
2. **Dado** una subcarpeta sin el archivo de entrada en ningún nivel permitido, **cuando**
   se completa el escaneo, **entonces** esa subcarpeta no aparece en el catálogo.
3. **Dado** un bot con varios archivos de entrada a distinta profundidad, **cuando** se
   descubre, **entonces** se usa el más superficial (regla determinista).
4. **Dado** un escaneo en curso, **cuando** el usuario observa la pantalla, **entonces**
   ve un indicador de actividad en lugar de una interfaz congelada.
5. **Dado** el catálogo visible, **cuando** el usuario solicita refrescar, **entonces** el
   escaneo se reejecuta en segundo plano y el listado se actualiza al terminar.

---

### Historia de Usuario 3 - Detalle del bot y ayuda de navegación (Prioridad: P2)

El usuario navega el catálogo con el teclado y puede abrir el detalle de un bot para ver su
información completa (nombre, ambiente, carpeta, archivo de entrada) antes de cualquier
acción futura. La pantalla siempre muestra los atajos de teclado disponibles.

**Por qué esta prioridad**: Confirma la información del bot y enseña la navegación que
usarán las siguientes funcionalidades, pero el catálogo ya es útil sin ella.

**Prueba independiente**: Se puede probar seleccionando un bot del listado y verificando
que el detalle muestra sus datos y que se regresa al listado.

**Escenarios de aceptación**:

1. **Dado** el catálogo con al menos un bot, **cuando** el usuario selecciona un bot y
   abre su detalle, **entonces** ve nombre, ambiente, carpeta raíz y archivo de entrada.
2. **Dado** el detalle abierto, **cuando** el usuario regresa, **entonces** vuelve al
   listado conservando su posición.
3. **Dado** cualquier pantalla visible, **cuando** el usuario mira el pie de pantalla,
   **entonces** ve los atajos de teclado aplicables a esa pantalla.

---

### Historia de Usuario 4 - Errores comprensibles, nunca cierres abruptos (Prioridad: P2)

Ante problemas frecuentes (archivo de configuración ausente o inválido, carpeta de
ambiente inexistente o inaccesible, fileserver lento o caído), la aplicación muestra un
mensaje claro en español indicando el problema y cómo resolverlo. La aplicación nunca se
cierra abruptamente por un error de entorno.

**Por qué esta prioridad**: Los usuarios son de negocio sin perfil técnico; un cierre
abrupto o un error en inglés les bloquea por completo, pero con mensajes claros pueden
autogestionarse o reportar correctamente a soporte.

**Prueba independiente**: Se puede probar arrancando con configuración inválida y con ruta
de ambiente inexistente, verificando que la aplicación muestra el error y sigue
responder al teclado.

**Escenarios de aceptación**:

1. **Dado** el archivo de configuración ausente, **cuando** el usuario abre la aplicación,
   **entonces** ve un mensaje que indica cómo crearlo a partir del ejemplo incluido.
2. **Dado** un archivo de configuración con sintaxis inválida o valores faltantes,
   **cuando** el usuario abre la aplicación, **entonces** se indica qué sección o campo
   falla, en español.
3. **Dado** la carpeta del ambiente inexistente o sin permisos, **cuando** se intenta
   escanear, **entonces** el catálogo muestra el error en lugar de un listado vacío
   silencioso, y el resto de la aplicación sigue operativo.
4. **Dado** cualquier pantalla con error visible, **cuando** el usuario navega o reintentaa,
   **entonces** la aplicación responde normalmente (no queda congelada ni se cierra).

---

### Casos Límite

- ¿Qué pasa cuando la carpeta del ambiente existe pero está vacía (sin subcarpetas)?
- ¿Cómo se maneja un nombre de subcarpeta con caracteres especiales, espacios o muy largo?
- ¿Qué ocurre si el usuario conmuta de ambiente mientras un escaneo está en curso?
- ¿Qué pasa si la carpeta de un bot contiene archivos de entrada en niveles más profundos
  que el límite configurado?
- ¿Cómo se comporta la aplicación si el fileserver responde extremadamente lento?
- ¿Qué pasa si el archivo de configuración define un ambiente sin ruta o con ruta relativa?

## Requisitos *(obligatorio)*

### Requisitos Funcionales

- **RF-001**: La aplicación DEBE leer su configuración desde un archivo TOML cuya ruta se
  puede pasar como argumento; sin argumento usa `config.toml` junto al ejecutable.
- **RF-002**: La configuración DEBE incluir: raíz del fileserver, ruta del archivo de base
  de datos (relativa a la raíz), y las rutas de bots de los ambientes `dev` y `prod`
  (ambos obligatorios).
- **RF-003**: La configuración DEBE permitir ajustar el nombre del archivo de entrada a
  buscar en cada bot (por defecto `main.py`) y la profundidad máxima de esa búsqueda
  (por defecto 5; 0 = sin límite).
- **RF-004**: La aplicación DEBE detectar el usuario de Windows de la sesión y mostrarlo
  en el encabezado; al no existir mapeo de rol disponible, DEBE mostrarse como "sin mapear
  (solo lectura)".
- **RF-005**: La aplicación DEBE mantener un ambiente activo visible (dev/prod) y permitir
  conmutarlo con el teclado en cualquier momento.
- **RF-006**: El catálogo de bots DEBE construirse escaneando las subcarpetas directas de
  la ruta del ambiente activo; el nombre del bot es el nombre de la subcarpeta.
- **RF-007**: Para cada subcarpeta la aplicación DEBE buscar el archivo de entrada de forma
  recursiva hasta la profundidad configurada; la subcarpeta sin entrada DEBE omitirse y la
  entrada más superficial DEBE ganar en caso de empates.
- **RF-008**: El escaneo DEBE ejecutarse en segundo plano mostrando un indicador de
  actividad; la interfaz DEBE permanecer responsiva durante el escaneo.
- **RF-009**: El usuario DEBE poder refrescar el catálogo con una tecla y al conmutar de
  ambiente el catálogo del nuevo ambiente DEBE cargarse automáticamente.
- **RF-010**: El usuario DEBE poder navegar el catálogo con el teclado y abrir el detalle
  de un bot con nombre, ambiente, carpeta y archivo de entrada.
- **RF-011**: Toda pantalla DEBE mostrar los atajos de teclado aplicables en un pie de
  pantalla.
- **RF-012**: Ante config ausente/inválida o errores de escaneo, la aplicación DEBE
  mostrar mensajes de error claros en español, con la causa y la acción sugerida, y DEBE
  permanecer operativa.
- **RF-013**: La aplicación DEBE permitir salir limpiamente con el teclado en cualquier
  pantalla.

### Entidades Clave

- **Bot**: un bot de automatización; nombre (su subcarpeta), ambiente, carpeta raíz y
  archivo de entrada (ruta relativa dentro de la carpeta del bot).
- **Ambiente**: contexto de trabajo (`dev` o `prod`) con su ruta de bots; exactamente uno
  está activo en la aplicación.
- **Usuario**: persona que opera el cliente; identificada por su usuario de Windows y un
  rol de negocio (el mapeo de roles se detalla en la funcionalidad 005).
- **Configuración**: parámetros del archivo TOML (fileserver, BD, ambientes, descubrimiento,
  interfaz); objeto cargado y validado al arrancar.

## Criterios de Éxito *(obligatorio)*

### Resultados Medibles

- **CE-001**: Un usuario de negocio abre la aplicación y, en menos de 3 segundos tras
  arrancar con config válida, ve su usuario de Windows y el ambiente activo en pantalla.
- **CE-002**: Con una carpeta de ambiente de 100 subcarpetas, el usuario puede navegar y
  usar la interfaz durante el escaneo sin congelamientos (el listado aparece al terminar
  el escaneo en segundo plano).
- **CE-003**: El 100% de los errores de configuración y de escaneo se muestran como
  mensajes en español con causa y sugerencia; cero cierres abruptos de la aplicación por
  errores de entorno.
- **CE-004**: El catálogo refleja fielmente las subcarpetas con archivo de entrada:
  ninguna subcarpeta sin entrada aparece, ninguna subcarpeta con entrada falta.
- **CE-005**: Conmutar de ambiente y refrescar el catálogo no requiere reiniciar la
  aplicación y se completa solo con el teclado.

## Supuestos

- Los usuarios ejecutan la aplicación en PCs Windows con sesión de dominio o local; el
  usuario de Windows es la identidad suficiente (sin login propio de la aplicación).
- El fileserver está accesible como carpeta compartida (ruta UNC o letra de red) y el
  usuario tiene permisos de lectura sobre las carpetas de bots.
- Los bots son carpetas cuyo punto de entrada se llama `main.py` salvo configuración
  distinta; esa estructura es la convención del equipo de migración.
- Los permisos por rol no se aplican en esta funcionalidad (solo se muestra el rol sin
  mapear); su aplicación es la funcionalidad 005.
- Las solicitudes de ejecución y el protocolo fileserver no son parte de esta
  funcionalidad (specs 002/003); aquí solo se definen los puertos que luego los usarán.
- La base de datos del fileserver puede no existir aún; esta funcionalidad no la consulta
  para nada esencial del catálogo.
