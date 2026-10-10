# Changelog

Registro de cambios del plugin llm-wiki. La versión inicial contiene capacidades cuya validación sigue pendiente.

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
