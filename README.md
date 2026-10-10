# llm-wiki

Gestiona tus notas y tu conocimiento desde Codex Desktop.

Busca información, guarda ideas y decisiones y relaciona notas. Tus archivos siguen siendo Markdown, compatibles con Obsidian.

**Estado:** versión `0.2.0` (tag `v0.2.0`), con áreas opcionales y workflow de OKRs. El núcleo y el paquete se probaron en macOS con Apple Silicon. La instalación por CLI tiene evidencia de versiones de desarrollo anteriores; la validación completa en Desktop sigue pendiente.

Cambios de esta entrega: [changelog v0.2.0](CHANGELOG.md).

## Qué puedes hacer

- Buscar y leer notas para recuperar contexto entre sesiones.
- Guardar ideas, decisiones y tareas mediante los workflows del plugin.
- Revisar enlaces, organización y vigencia de la información.
- Consultar áreas y vincular una misma nota con varias, conservando su ubicación.
- Crear o actualizar OKRs con datos aportados y campos pendientes explícitos.
- Preparar la estructura de un vault y exportar notas a JSON, Markdown u OKF.

Un **vault** es tu carpeta de notas. No necesitas mantener Obsidian abierto ni configurar claves de IA adicionales.

## Instalación

Necesitas Codex Desktop en macOS. El paquete incluye el ejecutable `plugins/llm-wiki/bin/llm-wiki`: no necesitas Go, Python ni Node para usarlo.

### Desde GitHub en Codex Desktop

El repositorio [santiago-migoni/llm-wiki](https://github.com/santiago-migoni/llm-wiki) contiene el catálogo y el paquete de esta entrega en la referencia Git `v0.2.0`.

1. Abre **Complementos → Agregar → Agregar marketplace**.
2. En **Origen**, introduce `santiago-migoni/llm-wiki` o `https://github.com/santiago-migoni/llm-wiki.git`.
3. En **Referencia de Git**, introduce `v0.2.0`. Deja **Rutas dispersas** vacío.
4. Pulsa **Agregar marketplace**, selecciona el marketplace **llm-wiki** e instala el plugin **llm-wiki**.
5. Abre un chat nuevo. Si el plugin no aparece, reinicia la aplicación.

### Desde un clon local en Codex Desktop

1. Abre **Complementos → Agregar → Agregar marketplace**.
2. En **Origen**, introduce la ruta absoluta a la raíz del clon, por ejemplo `/ruta/al/clon/llm-wiki`. Deja vacíos **Referencia de Git** y **Rutas dispersas**. Usa la raíz que contiene `.agents/plugins/marketplace.json`.
3. Selecciona el marketplace **llm-wiki** e instala el plugin **llm-wiki**.
4. Abre un chat nuevo. Si el plugin no aparece, reinicia la aplicación.

Si el marketplace ya registrado apunta a un clon o worktree antiguo y falla al cargar su manifiesto, elimina ese origen desde la administración de marketplaces y vuelve a agregarlo con la ruta actual. Que el plugin aparezca instalado no confirma que el origen siga disponible.

### Desde la terminal

Para instalar la entrega `v0.2.0` desde GitHub, con el CLI de Codex disponible:

```sh
codex plugin marketplace add https://github.com/santiago-migoni/llm-wiki.git --ref v0.2.0
codex plugin add llm-wiki@llm-wiki
codex plugin list --marketplace llm-wiki
```

Como alternativa, desde la raíz del clon local:

```sh
codex plugin marketplace add "$PWD"
codex plugin add llm-wiki@llm-wiki
codex plugin list --marketplace llm-wiki
```

El último comando permite comprobar la versión y si el plugin está habilitado. Después abre un chat nuevo en Desktop.

## Primera prueba

Escribe `$llm-wiki` en un chat nuevo y selecciona su skill. Usa una carpeta de notas de prueba y envía este mensaje, reemplazando la ruta y `<zona-horaria-IANA>` por la zona horaria que quieras usar:

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
[la evidencia de áreas 0.2.0](docs/superpowers/evidence/2026-10-10-areas.md).
El [informe 0.1.0](tests/acceptance/release-0.1.0.json) conserva su alcance histórico.

### Áreas y OKRs

Solicita «consulta mis áreas», «crea el área …» o «vincula esta nota con …»;
para OKRs, solicita crear, consultar o actualizar el objetivo. Áreas usan
`wiki/areas/` o `Areas/`; OKRs `wiki/okrs/` o `OKRs/`, salvo mapa explícito.
Los enlaces se guardan en el cuerpo. La clasificación es opcional: no duplica
notas, cambia metadata humana ni crea categorías obligatorias.

`wiki_list_areas` informa cobertura y diagnósticos; un inventario incompleto no
permite afirmar que un área no existe. La ejecución de los workflows en el host
y el ensayo con dos chats del paquete nuevo siguen pendientes. Consulta la
[matriz de casos](tests/acceptance/areas-cases.json) y el
[plan de este incremento](docs/superpowers/plans/2026-10-10-areas-transversales.md).

## Ayuda y documentación

Reporta problemas en el chat de prueba con versión del plugin, macOS, acción y error, sin notas privadas.

- [Workflows disponibles](docs/superpowers/evidence/2026-10-08-workflows.md)
- [Validación de Desktop](tests/acceptance/desktop.md)
- [Arquitectura](docs/superpowers/ARCHITECTURE.md) y [estructura del código](docs/superpowers/STRUCTURE.md)
- [Configuración, persistencia y recuperación](docs/superpowers/INFRASTRUCTURE.md)

## Desarrollo

El trabajo con Codex se guía por [AGENTS.md](AGENTS.md) y la
[configuración de desarrollo](.codex/README.md), que define roles para
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
