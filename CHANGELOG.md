# Changelog

Registro de cambios del plugin llm-wiki. La versión inicial contiene capacidades cuya validación sigue pendiente.

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
