# Evidencia de implementación local — 2026-10-08

Estado parcial de ejecución del plan aprobado, ejecución directa v003. Este registro distingue núcleo implementado, assets y aceptación del producto. No certifica instalación en Desktop ni paridad completa.

## Entorno

macOS 27.0.1, arm64. Go 1.27.2 darwin/arm64, aislado para desarrollo; SDK oficial MCP Go v1.8.0. SHA-256 del archivo de toolchain: `76812b213b1b2302c978d28fa52fa92d541704b9e7d9d5db8002c50e4018c4c5`. Codex CLI 0.135.0. No instalación global de Go.

## T1 — Registro y transporte

Commits `33c992c`, `14a4981`. Configuración versionada, asociación explícita proyecto/vault, lock interproceso y recarga bajo lock, reemplazo atómico. Zona IANA opcional validada; vacío no infiere zona del host. MCP STDIO con selección y consulta; stdout reservado al protocolo.

Los tests de dos registros abiertos y dos procesos concurrentes reprodujeron pérdida de asociaciones antes de la corrección. Después pasaron `go test -race ./internal/config ./internal/mcp`, `go test ./cmd/llm-wiki -count=1` y `go vet ./internal/config ./internal/mcp ./cmd/llm-wiki`. Revisión independiente Luna aprobó los cambios y corroboró las suites con race.

Prueba de proceso: selección, reinicio y consulta con estado persistente; raíz nueva devolvió `no vault selected`. Handshake MCP 2025-06-18, tools/list, timezone opcional y round-trip con zona explícita. Entorno reducido a PATH de sistema; stderr vacío. US-001.AC-02 sigue pendiente por interacción de selección en Desktop.

## T2 — Lectura y cobertura

Commits `9d8f223`, `1cc43d5`. Lectura paginada por runas, límite predeterminado 20000 cuando limit=0; límites negativos rechazados. SHA-256 de bytes completos. Escaneo completo con exclusiones, errores y truncamiento por cancelación. Layout mixto devuelve ambigüedad.

Los regresores reprodujeron aliases hacia estado protegido, escaneo truncado silenciosamente y layout mixto asumido. Segunda revisión detectó variantes de mayúsculas en macOS; la corrección reserva `.git` y `.llm-wiki` sin distinguir mayúsculas. Incluye protección de lectura del manual heredado. `go test -race ./internal/vault` y `git diff --check` pasaron. La prueba de permisos denegados se ejecutó sin skip. Revisión independiente Luna aprobó ambos rounds. Falta integrar estas operaciones al transporte MCP y probarlas desde Desktop.

## T6 — Assets, sin aceptación funcional

Commit `5bea20a`: router y 47 workflows, referencias, atribución MIT y matrices F/I/T. Pasaron validaciones estáticas de 48 frontmatter, enlaces relativos y JSON. Dos rondas de revisión Luna corrigieron políticas de reconcile, agenda con datos aportados, destinos de save, umbrales de síntesis e inferencia de fechas/tareas. El catálogo declara `behavior_verified:false`: faltan backend, pruebas reales por modalidad y continuidad entre sesiones.

## Empaquetado

La [documentación oficial de paquetes](https://developers.openai.com/plugins/build/plugins), consultada el 2026-10-08, y el help local del CLI corroboran marketplace local/Git, manifiestos portables y compatibilidad Codex. Aún no se creó ni instaló un paquete completo. El objetivo autorizado es marketplace Git; no se realizó publicación pública.

## Decisiones de ejecución

- Trabajo en rama aislada y worktree administrado; escrituras autorizadas con elevación porque el sandbox no incluye su ruta. Revertir el trabajo local sigue siendo posible.
- Encabezados normalizados únicamente en briefs temporales para el extractor; los requisitos aprobados permanecen intactos.
- Assets y lectura se adelantaron en áreas separadas mientras terminaba T1; integración y cierre permanecen sujetos a revisión. Si cambia un contrato, se adaptan antes de instalar.
- Mutaciones usarán `OpenWithState(root,stateDir)` explícito; evita inventar una segunda identidad para journals distinta del vault-id del registro. Puede ajustarse durante integración sin migración de datos reales.

Las brechas O-001–009 siguen abiertas según la especificación. Pruebas negativas y diagnósticos claros no sustituyen evidencia positiva de cada modalidad.

## T3 — Lotes y recuperación

Commits `47d3134`, `8d7b5d3`. `OpenWithState` exige estado externo explícito. Lotes create/update/move con precondiciones, journal de preparación, snapshots, bloqueo entre procesos y resultado parcial durable. Creaciones y destinos de movimientos se publican sin reemplazar un destino concurrente. Actualizaciones revalidan hash antes del rename, con el límite documentado frente a editores externos.

Dos rondas de revisión Luna cerraron creación concurrente, preparación interrumpida, identidad de registros al purgar, bloqueo por recuperación pendiente e inspección histórica. `go test -race ./internal/journal ./internal/vault`, `go test ./...`, `go vet ./internal/journal ./internal/vault` y `git diff --check` pasaron. Se probaron fallos deterministas después del primer archivo y durante snapshots, reapertura, reanudación, rollback que preserva ediciones humanas, ID con payload distinto, creación concurrente incluso con bytes iguales y purga que conserva pendientes.

Una caída entre publicación create-only y guardado de progreso no permite atribuir propiedad solo por igualdad de bytes: inspect/resume/rollback la presentan como conflicto. El archivo se conserva para reconciliación explícita; no se borra ni sobrescribe automáticamente. Operaciones terminales conservan su resultado histórico aunque después haya otros cambios. La retención de 30 días es un mínimo de purga explícita implementado, sin limpieza automática; O-006 sigue sujeto a la decisión de producto.

Faltan exposición completa MCP/CLI y aceptación instalada. T4 y T5 se ejecutan a continuación, en paralelo según la secuencia aprobada.

## T4/T5 — Verificación directa

Ejecución directa solicitada por el usuario. Suite `go test -race ./...` aprobada al retomar. Grafo: aliases, headings/blocks, embeds, links Markdown relativos, duplicados ambiguos, fences, Unicode, Canvas con scope y movimiento con rebasing/rollback. Se añadió healing unívoco y prueba RED→GREEN para heading propio y slug ambiguo/fence. Edición: planes capture/save/generated sections/move con bookkeeping y TZ explícita; source del movimiento conserva snapshot y el contenido del destino puede reescribir enlaces.

Búsqueda léxica reconstruible, cobertura, snippets/hash y métricas sobre dos casos sintéticos (recall@2=1, MRR=1; sin equivalencia semántica). Salud: metadata/contexto AI-first, enlaces, huérfanos, tareas vencidas, tags, plantillas, payload de fuentes, duplicados revisables, carpetas vacías mediante inventario explícito y FRESH-1–5 configurables. Corrección RED→GREEN de candidatos por título truncado. Semántica de contexto se señala para revisión de Codex, no se declara validada por forma textual. `go test -race ./internal/graph ./internal/vault ./internal/search ./internal/health` pasó tras las correcciones.

Falta aceptación de interacción instalada y unión con MCP. Se conservan las diferencias del host (render Canvas/Obsidian y recuperación semántica).

## T6 — núcleo de contexto y catálogo, ejecución directa

Por instrucción del usuario se continúa sin subagentes. `internal/catalog` prueba traversal, escape por symlink y personalizaciones conservadas ante cambio del paquete. `internal/service` recupera una decisión tras reabrir el registro; valida notas completas, incluso enlaces posteriores a los 20000 caracteres. El manual y los workflows tienen un acceso explícito confinado; el estado técnico continúa protegido. `internal/mcp` prueba guardar/leer/rechazar un hash obsoleto mediante el protocolo STDIO real.

Verificación: `go test -race ./...` y `go vet ./...`, Go 1.27.2 darwin/arm64, resultado satisfactorio. No valida aún instalación en Desktop ni todas las modalidades del inventario.

## T7 — operaciones directas integradas, aceptación parcial

Carpetas: dry-run no crea; create-only registra mkdir antes de archivos; rollback no elimina una carpeta con archivos humanos. Layout explícito wiki/obsidian se guarda en todas las asociaciones de la identidad. Bootstrap prueba 8 combinaciones de estilo/preset, assistant mode, Home/AGENTS humanos y repetición idempotente. Las plantillas actuales son una base Go: NO acreditan la fidelidad completa de todas las seed notes/Bases/vistas del original; ese alcance sigue pendiente. Init para actualizar bloques existentes también está pendiente.

Migración: combina secciones de una misma fecha, conserva el log íntegro como archivo con hash, y rechaza colisiones antes de sustituirlo. Merge: conserva ambos cuerpos, une aliases/listas, retiene conflictos de frontmatter en merged_from, actualiza backlinks inequívocos y crea redirect; preview/dry-run preceden aplicación explícita.

Exportación: JSON, Markdown y OKF; adjuntos con bytes preservados; destinos create-only; log acumulativo. OKF transforma enlaces fuera de fences, conserva campos desconocidos y conserva logs incompatibles como conceptos de historia. Validación estructural basada en §11 de https://github.com/GoogleCloudPlatform/open-knowledge-format/blob/main/SPEC.md (consulta 2026-10-08, versión 0.2). No prueba attestation ni equivalencia semántica. Cobertura incompleta se declara.

Verificación: `go test -race ./...` y `go vet ./...` pasan con Go 1.27.2 darwin/arm64. Las pruebas de integración guardan workflows fuera del paquete y exportan un adjunto binario; una exportación que colisiona no sobrescribe el bundle. Revisión directa agregó verificación de hash inmediatamente durante rollback y planificación de padres ausentes sin escrituras.

## T5/T7/T9 — actualización verificada de la ejecución directa

Este registro actualiza los pendientes históricos de las secciones anteriores. Bootstrap incorpora las plantillas, seed notes y Bases de las ocho combinaciones del original fijado, convertidas a placeholders Codex. Bases/Dataview se conservan como datos de compatibilidad; su render no está acreditado. Init actualiza bloques delimitados de manual/index y agrega auditoría conservando texto humano, `_CLAUDE.md` y AGENTS.md. Pruebas de idempotencia, preservación y presets pasaron. El catálogo usa YAML, categorías y plataformas, sin ejecutar workflows automáticamente.

Se expusieron `wiki_reindex` y `wiki_retrieval_eval`: índice léxico reconstruible en memoria, cambios por hash y evaluación de gold cases suministrados, sin generación ni escritura implícita. La búsqueda sin límite explícito devuelve hasta 20 resultados. `go test -race ./...` y `go vet ./...` pasaron tras estos cambios.

T9: generado paquete 0.1.0-dev.1 desde `deaa2b2`, universal x86_64/arm64, CGO=0, firmado ad-hoc y checksums verificados. Prueba de proceso MCP desde copia aislada con PATH de sistema: selección, guardar decisión/workflow, cerrar/reiniciar y recuperar ambos; pasó. Los esquemas portable plugin/mcp se validaron contra agent-plugins.org, usando jsonschema 4.25.1 solo para desarrollo.

Instalación soportada mediante `codex plugin marketplace add <worktree>` y `codex plugin add llm-wiki@llm-wiki`: instalado y habilitado 0.1.0-dev.1. La API CLI `plugin/read` reconoce 48 skills namespaced y MCP configurado. No acredita llamada en Desktop ni expansión de PLUGIN_ROOT en una sesión Desktop. No se publicó remoto.

La herramienta de control visual rechazó el acceso a com.openai.codex; no se usó una vía alternativa para evadir la restricción. Verificación visual solicitada al usuario, pendiente. Ejecución Intel y modalidades dependientes del host continúan pendientes.

## T9/T10 — segundo artefacto y recuperación entre procesos

0.1.0-dev.2 generado desde fuente `43ee802`. Ambos esquemas portables válidos; lipo identifica x86_64/arm64; codesign confirma firma ad-hoc válida. Prueba `TestPackagedBinaryWithoutDevelopmentToolchain` pasa con el paquete aislado y con la copia del catálogo. PATH únicamente de sistema: handshake/selección/guardado/lectura/workflow/reinicio, sin toolchain de usuario. Con `LLM_WIKI_PREVIOUS_PACKAGE` alterna dev.1 → dev.2 → dev.1 → dev.2 y compara todos los hashes del vault y estado, idénticos. Esta prueba es cambio de proceso/binario; no acredita el flujo de downgrade en Desktop.

`codex plugin add llm-wiki@llm-wiki` instaló dev.2; API `plugin/read` confirma versión, instalado/habilitado, 48 skills y MCP configurado.

`TestRecoveryAfterProcessExit` termina el primer proceso con exit 42 tras persistir un fallo parcial. Un segundo proceso inspecciona y reanuda, o rechaza rollback después de editar el primer archivo a "human". Pasa con race junto con idempotencia y purge que conserva pendientes. Son procesos de test, sin agentes ni chats nuevos. El límite de retención O-006 permanece abierto.

Se registran criterios locales T7 como passed con sus tests propietarios; las modalidades de skills y Desktop siguen pending. Ningún test local se presenta como métrica de éxito de usuarios.

## T8 — ensayo directo con capacidades presentes, parcial

Web nativa leyó la sección stdio de https://modelcontextprotocol.io/specification/2025-06-18/basic/transports. El agente redactó un dossier general con hechos, fecha, actores, cronología, fuentes, preguntas y límites de una sola fuente. Quedó guardado y re-leído mediante el binario empaquetado. No se evaluó modalidad académica ni ausencia de web.

Vault sintético: búsqueda de "piloto" escaneó 13 notas, seleccionó 12; el agente leyó sus textos y sintetizó conclusiones con citas, lagunas y una tarea propuesta. Un lote real creó síntesis/tarea/diaria/log. Todas las notas fuente conservan SHA-256. Artefactos en `tests/fixtures/host-capabilities`; casos y límites en `tests/acceptance/capabilities.json`. No acredita el modo File Search original ni una interacción Desktop.

La sesión expone herramientas opcionales Apple Calendar de otro plugin; no se llamaron ni se añadieron como dependencia. No se verificaron cola Telegram, extracción de audio/video, eventos/hooks, horarios ni costos por workflow. Los casos negativos de capacidad ausente siguen pending, no se registran respuestas simuladas como ejecución.
