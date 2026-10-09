---
title: Equivalencias de integraciones y ciclo de vida en Codex
status: draft
version: v001
created_at: "2026-10-08"
updated_at: "2026-10-08"
approval_reference: null
---

# Equivalencias de integraciones y ciclo de vida en Codex

## Alcance y criterio

Este análisis contrasta los registros I-001–I-012 y T-001–T-044 de
[integraciones](../inventario-codex/INTEGRACIONES.md) y [ciclo de vida](../inventario-codex/CICLO-DE-VIDA.md)
con la [Constitución vigente](../../CONSTITUTION.md) y la
[Arquitectura vigente](../../ARCHITECTURE.md). La fuente funcional es la copia
local `.repos/obsidian-second-brain`, commit
`b0089f7666d0c1e3db41c9d587b6f3c997665235`.

La Constitución v003 mantiene como meta las capacidades funcionales completas,
pero D-007 dirige a resolverlas desde Codex sin integraciones propias con
proveedores externos ni claves de esos servicios. La Arquitectura v004 está en
borrador: propone Codex Desktop, marketplace Git, macOS primero, Go para el
núcleo/herramientas propias y MCP local por STDIO. AC-019 deja como pendientes la
equivalencia de calendario, Telegram, X y medios, y la reconciliación del
adaptador opcional de Obsidian y dependencias de medios. No hay plugin propio
implementado ni validado.

En las tablas, **sustitución propuesta** significa una ruta que respeta la
dirección actual, no una capacidad disponible comprobada. **Brecha** señala que
no se puede afirmar equivalencia por la inspección estática. La disponibilidad
concreta de las capacidades del host Codex debe comprobarse en la versión,
cuenta y superficie objetivo; no se atribuye a Codex acceso a datos o cuentas
ajenas por defecto.

## Integraciones I-001–I-012

| ID | Resultado de referencia | Sustitución compatible con Codex/Go | Brecha y prueba pendiente |
|---|---|---|---|
| I-001 | 12 operaciones MCP sobre vault: consulta, paginación, captura, actualización, reemplazo, movimiento, validación, backlinks, salud e instrucciones. | Go/MCP STDIO puede exponer operaciones deterministas locales; Codex elige y explica el workflow. | **Propuesta; sin implementación.** Probar las 12 operaciones, confinamiento de rutas, paginación, archivos concurrentes, bookkeeping y fallos parciales desde el paquete instalado, sin leer `.repos`. |
| I-002 | Linter visual read-only dentro de Obsidian, con lista de hallazgos y navegación a notas. | Go puede ejecutar lint/reportes invocados desde Codex; las reglas de referencia pueden convertirse en fixtures compartidos. | Se pierde la UI y navegación dentro de Obsidian. TypeScript/API Obsidian no son necesarios para revisar notas desde Codex; conservar esa superficie requiere decisión separada conforme a AC-010/AC-019. Comparar hallazgos TS de referencia y Go/Codex con el mismo corpus si se mantiene. |
| I-003 | Captura automática por Telegram de texto, voz, fotos, PDF y enlaces, con polling, estado, cola y escritura al vault. | Codex puede procesar contenido que el usuario entregue explícitamente en una sesión y el núcleo Go guardarlo. | **Brecha verificable:** no hay receptor remoto, polling, identidad/allowlist ni cola automática. La captura manual no equivale al bot. No añadir bot, API de Telegram, token ni claves de IA bajo AC-019; medir por separado cada tipo de adjunto entregado a Codex. |
| I-004 | Investigación web/académica con fuentes, agregación, caché, warnings, modo libre y ruta pagada. | La capacidad web disponible en Codex puede recuperar fuentes; Codex sintetiza y Go guarda procedencia/notas. | No está probado qué cobertura, límites, citas, caché o continuidad tiene la superficie destino. Probar conjunto de consultas, fuentes, errores, límites y atribución en Codex Desktop. No recrear agregador/proveedor propio ni requerir claves. |
| I-005 | Investigación profunda: compara vault con fuentes web/X y produce delta para propagación. | Go busca corpus local con cobertura; Codex puede conducir búsqueda/síntesis cuando sus capacidades estén disponibles. | X no se cubre automáticamente; investigar con un conjunto conocido y confirmar baseline, citas, truncamiento y delta. Si Codex no expone el contenido requerido, registrar brecha; no añadir servicio/API. |
| I-006 | Tendencias de X (24–72 h), análisis de posts, hilos y respuestas, guardado opcional. | Codex podría analizar URLs/texto aportados por el usuario o resultados accesibles por capacidades nativas del host. | **Brecha probable a validar:** búsqueda temporal amplia, tendencias y respuestas completas dependen de acceso real. Probar URLs y consultas fechadas, cobertura y citas en la superficie prevista. No asumir acceso a X ni integrar xAI/X. |
| I-007 | YouTube: transcripción, metadatos/comentarios, resumen y extracción opcional de frames. | Codex puede sintetizar transcripción o capturas proporcionadas; Go registra fuente y adjuntos. | URL por sí sola no demuestra acceso a transcript/video. Probar video público/privado, idiomas, duración, transcript ausente y lectura visual. Sin `yt-dlp`/`ffmpeg`, la extracción automática de frames no está cubierta; no declarar equivalencia multimodal sin prueba. |
| I-008 | Podcast: RSS, transcripciones, audio/chunks, resumen y nota. | Codex procesa transcript o archivo de audio que el usuario aporte si la superficie habilitada lo admite; Go guarda el resultado. | No se encontró una herramienta nativa general de transcripción en la disponibilidad inspeccionada para esta tarea; no asumirla. RSS/descarga/transcripción automatizada y compatibilidad Spotify no demostradas. Probar con y sin transcript, audio largo, formato y límites del host. No instalar SDK/API de transcripción de proveedor ni pedir claves. |
| I-009 | Síntesis de corpus propio subido a Gemini File Search temporal y limpieza remota. | Go puede seleccionar y entregar al contexto de Codex los documentos locales necesarios; Codex responde con fuentes. | No equivale a indexación/grounding de corpus arbitrario ni garantiza límite/cobertura/privacidad. Probar corpus mayor al contexto, referencias exactas, omisiones y estado de carga. No subirlo a otro servicio ni crear store externo. |
| I-010 | Extractor Tavily, hasta 3 páginas y 8.000 caracteres, usado por investigación profunda. | Lectura web de Codex, si disponible, o texto/URL entregados por el usuario; preservar URL, fecha y truncamiento. | Cobertura de extracción, paywalls, contenido dinámico y límites no verificados. Fixture con página completa, parcial y fallida; reportar lo que Codex realmente leyó. No crear extractor/API propia para suplir faltantes. |
| I-011 | Proveedores de IA/búsqueda, modelos, fallback, coste/cache y guardado. | Codex es el motor de razonamiento; Go conserva estado local, procedencia y operaciones sobre vault. | No trasladar escaleras de modelos, ledger de tarifas de proveedores ni claves. Verificar cómo exponer capacidad no disponible, costes propios del host y datos efectivamente leídos; no afirmar offline. |
| I-012 | Pipeline multimodal de RSS/audio/video, formatos de transcript, frames y chunking. | Codex procesa formatos que soporte directamente; Go puede administrar archivos locales y metadatos. | La capacidad por formato y tamaño es desconocida hasta ensayarla. `ffmpeg` solo parece necesario si se conserva extracción/transcodificación local de audio/video; no es requisito del núcleo Go. Probar sin ese binario y documentar degradación, y con él solo si se autoriza como dependencia local opcional. |

### Lectura de brechas externas

Con la dirección aprobada, un resultado inaccesible desde las capacidades nativas
de Codex permanece como brecha funcional. Una alternativa manual (pegar texto,
adjuntar archivo o proporcionar URL) puede conservar parte del resultado, pero
cambia el flujo de captura, cobertura o automatización; debe evaluarse por tarea y
no contarse como paridad automática. No instalar conectores, bot, SDK de proveedor,
servicio cloud ni solicitar claves para cerrar esas brechas sin una decisión
posterior explícita.

La disponibilidad visible en esta sesión incluye herramientas de web, CUA y
automatización de Codex, pero puede depender de plugins, superficie, configuración
o cuenta; no prueba que una instalación del marketplace las tenga. En particular,
se observó un plugin Apple Calendar, que es una integración de otra herramienta y
queda fuera de AC-019: no usarlo ni contarlo como equivalente de calendario. La
fuente lista funciones de calendario en F-008, F-010 y F-013 de
[COMANDOS.md](../inventario-codex/COMANDOS.md). En el alcance actual, Codex puede
razonar sobre un calendario exportado o datos aportados por el usuario, pero no
hay evidencia de lectura/escritura de agenda, creación/reprogramación de eventos,
asistentes ni reconciliación bidireccional. Esta capacidad queda como brecha hasta
probar un mecanismo permitido por la decisión vigente.

Web/CUA disponibles en una sesión tampoco autorizan a navegar servicios como
bypass de AC-019 ni prueban que Codex tenga acceso completo a X, YouTube, podcasts
o contenido de pago. Validar las superficies permitidas en Codex Desktop destino;
mantener separadas capacidades del host y conectores de terceros.

`ffmpeg` es una herramienta local y no un proveedor de IA, pero su presencia
impone distribución, instalación y soporte adicionales. AC-011 ya orienta a
configurar dependencias de medios al usar la función y emitir diagnóstico; no
prueba que `ffmpeg` esté autorizado o sea inevitable. La ruta Codex debe primero
probar archivos/transcripts/medios que Codex procese sin esa herramienta. Si el
resultado visual de YouTube o la transcodificación de episodios sigue siendo un
requisito, presentar dependencia opcional, origen, licencias, plataformas,
instalación y prueba de fallback para decisión; no incorporarla por inferencia.

El adaptador TypeScript para Obsidian atiende una UI dentro de otra aplicación.
Para uso dedicado desde Codex, el lint determinista puede residir en Go y el
razonamiento/explicación en Codex; por ello TS no es requisito del núcleo ni de
la experiencia Codex. La arquitectura registra AC-010 (preferencia de adaptador
mínimo para funciones que requieran Obsidian), y AC-019 pide reconciliarla; esta
nota no la revoca. Antes de implementarlo, delimitar qué capacidad exclusiva de
la UI se conservaría y demostrar que no duplica o contradice las reglas Go.

## Ciclo de vida T-001–T-044

| ID | Equivalencia propuesta con Codex/Go | Estado o prueba que falta |
|---|---|---|
| T-001 | Skills Codex y núcleo Go resuelven manual, reglas y root explícito. | Propuesta; probar selección de vault/layout y que no se asuma contexto cargado. |
| T-002 | Reglas Markdown portables como instrucciones; validadores Go opcionales. | Propuesta; probar fixtures de frontmatter, fuente, frescura y enlaces. |
| T-003 | Búsqueda Go con cobertura; Codex razona y cita; fuentes son datos. | Propuesta; probar corpus grande, truncamiento, prompt injection en fuente y duplicados. |
| T-004 | Operaciones MCP Go coordinan las escrituras deterministas; Codex decide contenido/rutas. | Propuesta; probar propagación multiarchivo, fallo parcial y recuperación. |
| T-005 | Codex mantiene tiempo del evento y tiempo de incorporación en el esquema del vault. | Semántica aún no validada; casos temporales con fuente y contradicción. |
| T-006 | Skill/manual de vault carga o consulta SOUL/CRITICAL_FACTS según layout. | Mecanismo/ruta Codex pendientes; medir límite/contexto y fallback de lectura. |
| T-007 | Núcleo Go preserva `raw/`, mantiene índice/log según convenciones. | Probar inmutabilidad de origen, layouts alternativos y migración idempotente. |
| T-008 | Codex propone evolución; Go aplica cambios deterministas; separado lo interactivo de lo programado. | Probar contradicción clara/ambigua e historial; no hay automatización propia. |
| T-009 | Go detecta schema/layout existente; Codex respeta formatos Obsidian portables. | Probar wiki/Obsidian, Dataview y vault sin plugin. |
| T-010 | Comando/operación Go lint de frescura y reporte. | No implementado; probar ventanas, fechas, suppressions y configuración local. |
| T-011 | Skills pueden generar o actualizar manual preservando delimitadores y secciones locales. | Probar regeneración con contenido local y markers incompletos. |
| T-012 | Bootstrap en Go guiado por Codex, con preview de archivos a crear. | No implementado; pruebas destructivas/idempotentes en vault vacío y existente. |
| T-013 | Instalación se resuelve como plugin Codex/marketplace Git; no copiar instalador Claude. | No implementado; instalación, actualización y rollback en Codex Desktop macOS. |
| T-014 | Onboarding explícito Codex configura vault, MCP STDIO y preferencias en ubicación externa al paquete. | No implementado; permisos, rutas con espacios, rutas inválidas y limpieza. |
| T-015 | Skills/contexto de Codex y selección explícita de vault reemplazan SessionStart Claude. | Equivalencia automática por comprobar en Codex; probar dentro/fuera del vault y contexto grande. |
| T-016 | Reglas de skill más validación Go al final de operaciones de escritura. | Hook Codex no afirmado; probar validación tras cada herramienta y fallo del validador. |
| T-017 | No trasladar agente PostCompact que invoca Claude. Considerar rutina Codex solo si una superficie propia autorizada lo soporta. | La propagación autónoma no está equivalente; gate, alcance, concurrencia y auditoría pendientes. |
| T-018 | Hook Hermes queda fuera del runtime Codex. Su resultado de continuidad puede cubrirse por consulta explícita. | No equivalente a ejecución al cierre; conservar brecha, no adaptar Hermes. |
| T-019 | Codex Automations/heartbeat son candidatos para rutinas, solo con activación expresa y archivos accesibles. | No equivalencia demostrada; probar cuatro rutinas, permisos, disponibilidad de Desktop, reportes y política nightly. |
| T-020 | Skills en sesión Codex cubren interacción; tareas programadas serían un modo aparte. | Probar mismo resultado en ambas superficies; el wrapper `codex exec` del original no es requisito. |
| T-021 | Build del marketplace genera paquete autocontenido; scripts Go/skills incluidos. | No implementado; comprobar artefactos, exclusiones y que build no borre datos de usuario. |
| T-022 | Plugin Codex nativo sustituye build Agent Skills del upstream. | Referencia de empaquetado, no runtime nuevo; verificar que cada skill resuelve su herramienta MCP. |
| T-023 | Mantener soporte solo en superficies acordadas; evitar adaptador CLI legacy salvo requerimiento. | CLI no prometido por arquitectura; probar compatibilidad únicamente si se incluye en alcance. |
| T-024 | Adaptadores de otros agentes se omiten; no cambian formatos de datos interoperables. | Decisión consistente con Constitución; probar que paquete Codex no contiene runtimes ajenos. |
| T-025 | Skills nativas evitan wrappers que lanzan otra instancia de Codex. | Probar invocación desde sesión instalada; no reutilizar wrapper como evidencia. |
| T-026 | Go puede ofrecer actualización versionada con backup y rollback del paquete/estado. | No implementado; probar fast-forward equivalente, cambios locales, interrupción y recuperación. |
| T-027 | Herramienta Go de salud/integridad con reporte separado de reparaciones. | No implementado; validar fixtures de links, duplicados, cobertura y permisos read-only. |
| T-028 | Go repara coincidencias deterministas; Codex clasifica ambiguos sin aplicar automáticamente. | Probar dry-run, code fences, aliases, ambiguos e idempotencia. |
| T-029 | Merge coordinado Go más contenido propuesto por Codex y redirect con procedencia. | Probar frontmatter, enlaces entrantes, colisiones y rollback; no mantenimiento autónomo. |
| T-030 | Migración Go conserva fuente y ofrece dry-run antes de reemplazo. | Probar secciones, colisiones, idempotencia e interrupción. |
| T-031 | Go actualiza solo bloque generado del índice. | Probar sentinels presentes/rotos y preservación del texto circundante. |
| T-032 | Go implementa utilidades locales (grafo, commits, export) por demanda. | Catálogo de alcance y fixtures pendientes; no afirmar paridad genérica. |
| T-033 | Go genera lint y export OKF/Markdown cuando se invoque. | Probar esquema, Unicode, YAML inválido y salidas; dependencia PyYAML no debe heredarse. |
| T-034 | Evaluación local Go de retrieval y métricas reproducibles. | Probar recall/MRR/cobertura sin enviar notas a servicio externo; proteger datos reales de fixtures distribuidos. |
| T-035 | Corpus sintético evalúa retrieval; calidad semántica se evalúa dentro de Codex si hay método permitido. | La evaluación de proveedores/juez externo se excluye; especificar criterio Codex y repetir resultados. |
| T-036 | Herramientas de inventario/conformance para builds del marketplace. | No es función de usuario; validar `--check` contra artefacto generado y evitar escrituras no pedidas. |
| T-037 | Skills Codex conducen recap/review/learn/reconcile/health/reindex. | Probar resultado de cada comando y diferencia interactivo/automático. |
| T-038 | Codex interpreta conversación; núcleo Go persiste personas/proyectos/tareas/daily/log según schema. | Probar no duplicación, zona horaria, destino, y autorización contextual antes de escrituras múltiples. |
| T-039 | Codex hace síntesis/crítica; Go provee grafo, canvas y operaciones deterministas. | Probar que interpretaciones no se presentan como hechos y que export abre en aplicaciones compatibles. |
| T-040 | Skills conservan reglas de escritura, sentinel, estado, fechas y archivo. | Pruebas de regresión de texto humano, bloques generados y archivado. |
| T-041 | Descripciones de skills codifican triggers explícitos/proactivos en Codex. | Probar activación multilingüe, falsos positivos y exclusiones reales. |
| T-042 | Go/MCP sustituye las operaciones locales del MCP de referencia; captura remota no se traslada. | 12 herramientas y lint Codex por probar; Telegram permanece brecha explícita. No asumir TS/API Obsidian. |
| T-043 | Marketplace versionado y notas de cambios preservan personalizaciones fuera del paquete. | Probar actualización/reversión con cambios del usuario y catálogo Git; remoto/publicación no autorizados por esta nota. |
| T-044 | Sin claves de proveedores ni ledger de sus tarifas; documentar tratamiento de datos/costes de Codex y estado local. | No declarar offline o coste cero. Auditar configuración, logs, telemetría del host y cualquier dependencia local al implementar. |

## Matriz de pruebas de equivalencia prioritaria

Antes de declarar cobertura, validar en Codex Desktop macOS, con paquete instalado
desde un marketplace de prueba o una instalación aislada equivalente, y un vault
sintético. Mantener separado el resultado de empaquetado del resultado de usuario.

1. **Núcleo (I-001, T-001–T-011, T-027–T-031, T-037–T-040):** probar búsqueda completa y parcial, paginación, escritura propagada, conflictos concurrentes, fallo a mitad de una operación, rollback y preservación de contenido local en ambos layouts.
2. **Codex como fuente (I-004–I-010, I-011, T-035, T-044):** usar un set fijo de páginas, URLs X, videos y audios con transcript conocido; comparar cobertura, citas, fecha, truncamiento, modalidad realmente leída y comportamiento ante acceso denegado. Documentar capacidades del host, sin sumar proveedor o clave.
3. **Brechas de automatización (I-003, T-015–T-020):** verificar captura manual de adjuntos; probar carga de contexto, hooks y scheduling solo donde Codex los ofrezca en la superficie elegida. No contar polling Telegram o tareas que requieren otro runtime como equivalentes.
4. **Medios/dependencias (I-007, I-008, I-012):** ejecutar el caso base sin `ffmpeg`, `yt-dlp` ni herramientas auxiliares. Registrar resultado degradado; si frames o transcodificación son requisito, presentar la dependencia local para decisión y probar instalación, versión, salida, limpieza temporal y error.
5. **Distribución/calidad (I-002, T-012–T-014, T-021–T-026, T-032–T-036, T-041–T-043):** instalación macOS, actualización/rollback, paquete autocontenido, rutas con espacios, configuración externa al bundle, métricas sintéticas y consistencia entre reglas e instrucciones.

La verificación de la fuente sólo demuestra el build de `agent-skills` del original
(46 skills más `obsidian-core`) y un error de import de `vault_ops` en su salida;
no prueba funcionamiento propio en Codex. Ver [VERIFICACION.json](../inventario-codex/VERIFICACION.json).
