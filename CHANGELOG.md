# Changelog

Registro de cambios del plugin LLM Wiki. La versión inicial contiene capacidades cuya validación sigue pendiente.

## v0.2.3 — 2026-10-18

### Correcciones

- El nombre de un área se toma del H1 que sigue a un comentario HTML cerrado en la
  misma línea, en lugar de caer al nombre del archivo.
- Si la carpeta de áreas no resuelve, `wiki_list_areas` informa solo `invalid_folder`
  y ya no marca como `misplaced_area` las áreas que sí están en su carpeta.
- Una cancelación durante el análisis de áreas genera un único diagnóstico.

### Notas

- Incluye la migración a Claude descrita en la sección siguiente. La evidencia de
  Desktop sigue correspondiendo a 0.2.2; no se repitió para 0.2.3.

## Migración a Claude (incluida en 0.2.3)

### Cambios

- El plugin se distribuye para Claude Desktop: `.claude-plugin/plugin.json` y
  `.claude-plugin/marketplace.json`. El servidor MCP usa
  `${CLAUDE_PLUGIN_ROOT}/bin/llm-wiki serve`.
- Se retiran los manifiestos, el marketplace y la configuración de Codex
  (`.codex-plugin/`, `.mcp.json`, `mcp.json`, `plugin.json` raíz, `.agents/` y `.codex/`).
- El empaquetador valida la nueva ruta de lanzamiento y la coherencia entre
  el catálogo y el manifiesto del plugin.
- Las skills y referencias nombran a Claude como host.

### Sin cambios de runtime

- Las plantillas y el texto generado en vaults nuevos conservan su contrato
  actual; la lectura de notas antiguas que mencionan Codex sigue funcionando.

## v0.2.2 — 2026-10-12

### Correcciones

- Placeholders YAML entrecomillados en las plantillas, incluido `last_interaction`
  de Person: bootstrap deja el inventario de áreas completo.
- Salud acepta tags en listas YAML inline y de bloque; conserva diagnósticos
  para elementos no textuales.
- Salud distingue plantillas de notas finales: mantiene errores estructurales y
  YAML, sin exigir fechas resueltas ni analizar enlaces placeholder como notas.
- Literature Note usa `{{source}}` para enlazar la obra, separado de `{{title}}`.
- El detector de contenido pendiente reconoce los placeholders actuales y la
  sintaxis anterior; respeta código y Mustache desconocido legítimo.
- Post incluye `post` en tags y conserva `content`.
- Las colas de lectura incluyen `type`, tags, `ai-first` y contexto futuro.
- Nombre visible **LLM Wiki** en plugin y marketplace; identidad técnica
  `llm-wiki` conservada.

### Validación y distribución

- Regresiones del fallo y de casos legítimos; generación y salud de los ocho
  cruces `wiki`/`obsidian` con `executive`, `builder`, `creator` y `researcher`:
  60 plantillas Markdown, 24 Bases y dos colas. Plantillas humanas existentes
  preservadas; sin migración automática de bóvedas.
- Suite con carreras, vet y aceptación MCP real del paquete sin toolchain de
  desarrollo. La procedencia, checksums, firma y arquitecturas del paquete
  reconstruido se registran en el informe de entrega.
- Verificación real en un chat nuevo de Desktop: 30 herramientas expuestas y
  28 llamadas a 12 herramientas; bootstrap `obsidian`/`researcher` aplicado e
  idempotente, escritura recuperable y lectura/búsqueda correctas, áreas completas,
  ocho plantillas con YAML válido y salud de 15 notas sin errores ni advertencias.
  Manifiestos/build/checksum corroboran 0.2.2; las herramientas probadas no devuelven
  la versión del proceso y su ruta no se observó.
- Se cierra la carga de esquemas en un chat nuevo y el flujo probado. La aceptación
  anterior de dos chats sigue asociada a 0.2.0; no se verifica refresco de sesiones
  existentes, ejecución Intel, renderizado Obsidian/Dataview ni calendarios externos.

Evidencia: [informe 0.2.2](tests/acceptance/release-0.2.2.json) y
[regresiones de plantillas](docs/superpowers/evidence/2026-10-12-template-regressions.md),
[verificación Desktop](docs/superpowers/evidence/2026-10-12-desktop-0.2.2.md) y
[registro de llamadas Desktop](tests/acceptance/desktop-0.2.2.json).

## v0.2.1 — 2026-10-10

### Correcciones

- Arranque MCP compatible con el cargador de Codex Desktop: ambos manifiestos
  usan `./bin/llm-wiki`, `serve` y `cwd: "./"`, sin placeholders en el comando.
- Validación de ambos descriptores y prueba de arranque desde una ruta con espacios.
- Arrays MCP con tipos explícitos: objetos en `changes` y `cases`, strings en
  `scope` y `relevant`, contenido binario base64 y destino nullable.
- La prueba de alternancia de paquetes verifica la versión real de cada proceso;
  no ejecuta inadvertidamente el paquete nuevo durante downgrade.

### Documentación y validación

- Evidencia real de 24 casos de doce workflows, 14 escenarios de guardas y
  recuperación en dos chats Desktop sobre la corrección local 0.2.0. Fuentes raw,
  metadata humana y columnas Kanban preservadas; OKRs desconocidos pendientes.
- Documentación vigente actualizada y paquete universal 0.2.1 reconstruido.
- Conservados los límites: recarga del esquema en sesiones existentes, instalación
  Desktop de 0.2.1, renderizado Obsidian, calendario remoto, ejecución Intel y
  resultados de usuario EXI-002/003 pendientes. No se presume equivalencia completa.
- Entrega preparada localmente; el tag 0.2.1 aún no se publica en el remoto.

Evidencia: [informe 0.2.1](tests/acceptance/release-0.2.1.json) y
[aceptación Desktop 0.2.0](docs/superpowers/evidence/2026-10-10-desktop-areas.md).

## v0.2.0 — 2026-10-10

### Funcionalidades

- Áreas opcionales elegidas por el usuario; una nota puede vincularse con varias
  mediante enlaces en el cuerpo, conservando su ubicación y contenido.
- Lector `wiki_list_areas` compartido por CLI y MCP, con hashes, aliases,
  cobertura, diagnósticos e inventarios parciales explícitos.
- Workflows de áreas y OKRs, e integración opcional en doce workflows existentes.
- Resolución de carpetas de áreas, OKRs y entidades con precedencia de mapas
  explícitos y errores de destino sin fallback silencioso.

### Correcciones

- Las tablas de convenciones de nombres conservan los mapas de carpetas válidos.
- Nombres de área mediante H1 ATX/Setext; se descartan comentarios HTML y bloques
  de código, y no se confunden listas numeradas con encabezados Setext.
- Fuentes de mapa ilegibles producen `map_errors` e inventarios incompletos.
- Destinos bloqueados por archivos y enlaces incompatibles se diagnostican antes
  de crear relaciones; títulos como «C#» conservan su nombre visible.

### Distribución, documentación y validación

- Runtime, manifiestos y paquete macOS universal en `0.2.0`; referencia Git
  `v0.2.0`. Documentación vigente y evidencia vinculadas a esta entrega.
- Suite Go con detector de carreras, vet, pruebas de recuperación y continuidad
  entre procesos aprobadas. Firma, arquitecturas, checksums y procedencia del
  paquete verificados; prueba MCP con PATH del sistema aprobada.
- Las tres reproducciones del PR pasaron sobre el binario recompilado.
- Instalación y workflows en Codex Desktop, dos chats reales, ejecución Intel y
  resultados de usuario EXI-002/003 continúan pendientes. Las simulaciones de
  instrucciones no equivalen a aceptación del host.

Evidencia: [registro de áreas](docs/superpowers/evidence/2026-10-10-areas.md) e
[informe 0.2.0](tests/acceptance/release-0.2.0.json).

## 0.1.0 — 2026-10-09

Primera versión registrada en el historial público del proyecto.

### Funcionalidades

- Runtime local en Go, con CLI y transporte MCP STDIO para gestionar notas Markdown.
- Selección explícita de vault, lectura, búsqueda léxica, análisis de enlaces y
  diagnóstico de salud.
- Escrituras recuperables con precondiciones por hash, snapshots y journal.
- Workflows declarativos, preparación de la estructura del vault y exportación
  a JSON, Markdown y OKF.
- Marketplace y paquete de instalación con ejecutable universal para macOS,
  skills, referencias, metadatos de compilación y checksums.
- Documentación de instalación desde GitHub, clon local y terminal, junto con
  evidencia de aceptación y límites de capacidades.

### Correcciones y seguridad

- Corrección del fallo del parser ante contenido Markdown compuesto solo por
  espacios en blanco.
- Límites de lectura y escritura de 8 MiB por archivo y presupuestos compartidos
  para recorridos, catálogo y exportación, evitando cargas de memoria sin límite.
- Recorridos de directorio por lotes, lecturas cancelables y reporte de cobertura
  incompleta cuando se exceden los presupuestos.
- Presupuestos de comparaciones y distancia de edición en el diagnóstico de
  duplicados para limitar su costo computacional.
- Ejemplos portables, identidad genérica del proyecto y eliminación de rutas
  personales de la documentación versionada.
- Metadatos del binario sin atribución a un commit que no contiene sus fuentes;
  procedencia registrada mediante hash del árbol de fuentes.

### Desarrollo

- Instrucciones de desarrollo en `AGENTS.md` y configuración de Codex en `.codex/`.
- Roles de implementación, revisión de código, seguridad y distribución;
  los revisores trabajan en modo de solo lectura.
- Hasta tres agentes delegados simultáneos, con modelos heredados del entorno
  del colaborador.

### Validación y límites conocidos

- Versión 0.1.0: suite Go con detector de carreras, vet, esquemas de manifiestos,
  checksums, firma y prueba MCP aislada con persistencia entre procesos aprobados
  en macOS con Apple Silicon. El handshake MCP confirma la versión 0.1.0.
- La instalación por CLI tiene evidencia de versiones de desarrollo anteriores;
  no se repitió la instalación de 0.1.0.
- Validación completa de instalación y workflows en Codex Desktop pendiente.
- El paquete incluye arquitectura Intel; su ejecución en Intel no está verificada.
- La búsqueda es léxica; no se acredita equivalencia con búsqueda semántica.
- Integraciones externas, transcripción, automatizaciones y vistas de Obsidian
  conservan capacidades o pruebas pendientes. La instalación no activa hooks
  ni horarios.

Consulta `plugins/llm-wiki/capabilities.json` y `tests/acceptance/results.json`
para la evidencia detallada de capacidades y pruebas.
