# LLM Wiki

Gestiona tus notas y tu conocimiento desde Claude Desktop.

Busca información, guarda ideas y decisiones y relaciona notas. Tus archivos siguen siendo Markdown, compatibles con Obsidian.

**Estado:** versión `0.2.2` (referencia `v0.2.2`), con las siete correcciones de plantillas y el nombre visible **LLM Wiki**. El identificador técnico sigue siendo `llm-wiki`, incluidos CLI, skills y rutas. La validación del paquete y el estado de publicación e instalación se registran en el [informe 0.2.2](tests/acceptance/release-0.2.2.json).

Cambios de esta entrega: [changelog v0.2.2](CHANGELOG.md).

## Qué puedes hacer

- Buscar y leer notas para recuperar contexto entre sesiones.
- Guardar ideas, decisiones y tareas mediante los workflows del plugin.
- Revisar enlaces, organización y vigencia de la información.
- Consultar áreas y vincular una misma nota con varias, conservando su ubicación.
- Crear o actualizar OKRs con datos aportados y campos pendientes explícitos.
- Preparar la estructura de un vault y exportar notas a JSON, Markdown u OKF.

Un **vault** es tu carpeta de notas. No necesitas mantener Obsidian abierto ni configurar claves de IA adicionales.

## Instalación

Necesitas Claude Desktop en macOS con la pestaña Code. El paquete incluye el ejecutable `plugins/llm-wiki/bin/llm-wiki`: no necesitas Go, Python ni Node para usarlo.

### Desde GitHub

En una sesión de la pestaña Code, registra el marketplace y luego instala el plugin:

```text
/plugin marketplace add santiago-migoni/llm-wiki
/plugin install llm-wiki@llm-wiki
```

La referencia `v0.2.2` identifica la entrega; comprueba su publicación en el [informe de entrega](tests/acceptance/release-0.2.2.json). Fijar la versión con una referencia Git en este flujo aún no está verificado en Claude Desktop.

### Desde un clon local

Desde la raíz del clon, que contiene `.claude-plugin/marketplace.json`:

```text
/plugin marketplace add /ruta/absoluta/al/clon/llm-wiki
/plugin install llm-wiki@llm-wiki
```

Si el marketplace registrado apunta a un clon antiguo y falla al cargar, elimínalo desde la gestión de plugins y vuelve a añadirlo con la ruta actual. Que el plugin aparezca instalado no confirma que el origen siga disponible.

## Primera prueba

Abre un chat nuevo en Claude Desktop y envía este mensaje, usando una carpeta de notas de prueba, reemplazando la ruta y `<zona-horaria-IANA>` por la zona horaria que quieras usar:

> Usa llm-wiki con el vault `/ruta/a/mis-notas`. Selecciónalo para este proyecto, configura la zona horaria `<zona-horaria-IANA>` y carga su contexto. Lista las notas y lee una, sin modificar archivos.

Después puedes probar acciones concretas:

- «Usa llm-wiki para buscar las notas sobre mi proyecto».
- «Usa el workflow de captura de llm-wiki para guardar esta idea: …».
- «Usa el workflow de salud de llm-wiki para revisar mis notas y mostrar los hallazgos».

Cada escritura informa los archivos cambiados y los conflictos.

## Estado y límites

La búsqueda actual es léxica: busca coincidencias de palabras, sin acreditar equivalencia con búsqueda semántica. Las modalidades de los workflows siguen en validación.

Calendarios, Telegram, X, transcripción de medios, automatizaciones y vistas de Obsidian tienen pruebas o capacidades pendientes. No se activan hooks ni horarios al instalar. El binario incluye Intel, pero su ejecución en esa arquitectura aún no se verificó.

Consulta el [plan y su avance](docs/superpowers/plans/2026-10-08-codex-plugin.md) y los [resultados de aceptación](tests/acceptance/results.json) para distinguir implementación de comportamiento verificado.

Los informes de aceptación anteriores conservan las versiones de desarrollo que
probaron. La validación del paquete actual se registra en
[el informe 0.2.2](tests/acceptance/release-0.2.2.json).
La [evidencia de plantillas](docs/superpowers/evidence/2026-10-12-template-regressions.md) registra los ocho cruces de layout/preset, 60 plantillas Markdown, 24 Bases y dos colas, con pruebas MCP del paquete sin toolchain de desarrollo. No acredita vistas ejecutadas en Obsidian ni sesiones nuevas de Desktop.
La [verificación en un chat nuevo de Desktop 0.2.2](docs/superpowers/evidence/2026-10-12-desktop-0.2.2.md) acredita 30 herramientas MCP expuestas y 28 llamadas a 12 herramientas: bootstrap `obsidian`/`researcher`, escritura recuperable, lectura, búsqueda, inventario de áreas completo y salud sin errores ni advertencias. Manifiestos, build y checksum corroboran 0.2.2; las respuestas de las herramientas probadas no exponen la versión del proceso ni se observó su ruta. Cierra el pendiente de carga de esquemas en un chat nuevo y del flujo probado; no verifica refrescar sesiones existentes.
La [aceptación Desktop sobre la corrección local 0.2.0](docs/superpowers/evidence/2026-10-10-desktop-areas.md) conserva el binario y alcance realmente probados.
El [informe 0.1.0](tests/acceptance/release-0.1.0.json) conserva su alcance histórico.

### Áreas y OKRs

Solicita «consulta mis áreas», «crea el área …» o «vincula esta nota con …»;
para OKRs, solicita crear, consultar o actualizar el objetivo. Áreas usan
`wiki/areas/` o `Areas/`; OKRs `wiki/okrs/` o `OKRs/`, salvo mapa explícito.
Los enlaces se guardan en el cuerpo. La clasificación es opcional: no duplica
notas, cambia metadata humana ni crea categorías obligatorias.

`wiki_list_areas` informa cobertura y diagnósticos; un inventario incompleto no
permite afirmar que un área no existe. La ejecución sintética de los doce workflows y la recuperación en dos chats
se verificaron sobre la corrección local 0.2.0. Los chats abiertos aún conservaban
la declaración antigua de los lotes después de reinstalar: el cargador nuevo
ya publicaba los esquemas correctos. La instalación de la entrega actual se registra
en el informe 0.2.2. La carga en un chat nuevo y el inventario completo se verificaron
para 0.2.2 en la evidencia Desktop citada arriba; no se repitió la matriz completa
de dos chats ni el refresco de sesiones existentes. Consulta la
[matriz de casos](tests/acceptance/areas-cases.json) y el
[plan de este incremento](docs/superpowers/plans/2026-10-10-areas-transversales.md).

## Ayuda y documentación

Reporta problemas en el chat de prueba con versión del plugin, macOS, acción y error, sin notas privadas.

- [Workflows disponibles](docs/superpowers/evidence/2026-10-08-workflows.md)
- [Validación de Desktop](tests/acceptance/desktop.md)
- [Arquitectura](docs/superpowers/ARCHITECTURE.md) y [estructura del código](docs/superpowers/STRUCTURE.md)
- [Configuración, persistencia y recuperación](docs/superpowers/INFRASTRUCTURE.md)

## Desarrollo

El trabajo de desarrollo se guía por [CLAUDE.md](CLAUDE.md), que define roles para
implementación, revisión de código, seguridad y distribución.

### Límites de recursos del runtime

Las lecturas de notas, metadata (incluido el catálogo de workflows) y adjuntos admiten hasta 8 MiB por archivo. La paginación conserva el hash del archivo completo; superar el límite devuelve un error, sin presentar un hash parcial. El escaneo admite hasta 64 MiB de Markdown y 10.000 entradas, incluidos directorios y adjuntos. Al superar un límite, devuelve cobertura incompleta y detiene las operaciones que necesitan ese escaneo. La exportación limita también a 64 MiB la suma de notas y adjuntos cargados.

El diagnóstico de duplicados comparte un presupuesto de 1.000 comparaciones y 8 millones de celdas de distancia de edición. Cada cuerpo admite hasta 4.096 caracteres normalizados y 64 KiB antes de normalizar. Cuando no puede completar el análisis, informa `DUPLICATE_COVERAGE` y similitud no disponible. Las escrituras de archivos sobre 8 MiB se rechazan antes de registrar una operación. Los archivos existentes sobre el límite tampoco se cargan para preparar modificaciones o snapshots. El catálogo comparte los límites de archivos, bytes totales y entradas, con lecturas cancelables.

Solo para modificar el código o reconstruir el paquete: Go 1.27 o posterior y las herramientas de macOS `lipo` y `codesign`.

Desde la raíz del repositorio:

```sh
go test -race ./...
go vet ./...
go run ./cmd/package --output /tmp/llm-wiki-marketplace
```

La carpeta de salida debe ser nueva. El empaquetador genera un marketplace con el binario universal, skills, referencias y checksums. Los cambios y sus pruebas deben respetar la [especificación del proyecto](docs/superpowers/specs/2026-10-08-codex-plugin-spec.md).

## Procedencia

Adaptado de obsidian-second-brain v0.17.0, commit `b0089f7666d0c1e3db41c9d587b6f3c997665235`, de Eugeniu Ghelbur. El material adaptado conserva su atribución y licencia MIT en [NOTICE](plugins/llm-wiki/NOTICE).
