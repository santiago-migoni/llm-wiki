# Escritura, enlaces y propagación

Antes de crear algo, busca por título, aliases, persona y carpeta. Una coincidencia exacta suele actualizar la nota existente; un nombre parecido pero distinto requiere título que los distinga. No declares ausencia a partir de memoria o de una búsqueda parcial.

Toda escritura propaga solo relaciones que el workflow exige: decisión → proyecto y diaria; tarea → tablero, proyecto/diaria cuando existan; persona → interacción diaria; dev log → actividad de proyecto y diaria; síntesis → fuentes y log conforme al modo. No crees huérfanos. Lee la nota completa y actualiza su sección concreta; preserva párrafos humanos, historial y bloques fuera de sentinels. Si el workflow regenera un bloque, reemplaza solo `@generated:start`…`@generated:end`; nunca toques `@user` ni texto exterior.

Conserva formato de tablero: headings H2 son columnas, tareas checkboxes, fechas `@{YYYY-MM-DD}`, prioridad 🔴/🟡/🟢. Las terminadas se mueven a Done con tachado; no se eliminan. Tableros no llevan `## For future agent`. Para archivos Markdown nuevos, sigue [ai-first-rules.md](ai-first-rules.md); lee 1–2 notas vecinas para igualar nombres de secciones y estilo.

Raw existente nunca se modifica. Añade una fuente nueva con creación exclusiva. Actualizaciones usan contenido/hash leído como precondición; reemplazo exacto requiere una única coincidencia; move requiere destino inexistente. Para lote usa operaciones `wiki_*` disponibles y revisables. Haz dry-run solo cuando el modo lo indique; no conviertas el dry-run en escritura.

Respeta la política específica de cada skill: escribe los resultados que el workflow guarda por defecto, ofrece solo lo que el workflow ofrece, y confirma solo las acciones que el workflow manda confirmar. Ante cambios existentes derivados de una fuente externa, usa la regla de confirmación de [ai-first-rules.md](ai-first-rules.md). Reporta archivos creados/actualizados, omitidos, conflictos y errores parciales.

No ejecutes scripts ni shells para sustituir capacidades ausentes. Usa herramientas `wiki_*` si están expuestas y las herramientas nativas de Codex disponibles para lectura, razonamiento o edición segura; nunca afirmes que una herramienta externa, transcripción, calendario, X, hook o tarea programada corrió si la sesión no la ofreció.
