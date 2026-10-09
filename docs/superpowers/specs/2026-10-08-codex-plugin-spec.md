---
title: Especificación funcional de llm-wiki para Codex Desktop
status: approved
version: v001
responsible: Usuario promotor del proyecto
created_at: "2026-10-08"
updated_at: "2026-10-08"
reviewed_at: null
approval_reference: "AP-001 — aprobación y ejecución"
---

# Especificación funcional de llm-wiki para Codex Desktop

## Contexto y resultado

Como usuario de Codex quiero gestionar conocimiento personal/profesional desde
Codex Desktop, con notas portables y continuidad entre sesiones, conservando los
resultados de obsidian-second-brain sin integrar otros proveedores.

Esta es la especificación funcional canónica de la adaptación. No se asignó una
versión de producto ni se aprobó un recorte de release; por eso se usa una ruta
fechada, sin inventar v0.1.0. `v001` es revisión del documento, no del producto.
La solicitud actual autoriza redactar spec y luego plan; no registra aprobación
anticipada de estos borradores ni inicia implementación.

Fuentes:

- [Constitución](../CONSTITUTION.md), v003 aprobada con su alcance registrado en
  `f27672f7a63039805784efd8ccd5478b089dda2e:docs/superpowers/CONSTITUTION.md`;
  PRI-001–005, OBJ-001–003 y EXI-001–003.
- [Arquitectura](../ARCHITECTURE.md), v005, y estructura/infraestructura v001 en
  `80ae6c8b5849301acf6b5a46166b34cbc79a2aed`: acuerdos AC-001–019 y propuestas
  ADR-006–008 diferenciados. Las propuestas técnicas no son decisiones aprobadas.
- [Inventario](../analysis/inventario-codex/README.md), commit
  `22fd2d36ea33415a8f44211ae346f2a775021553`; referencia v0.17.0 en
  `b0089f7666d0c1e3db41c9d587b6f3c997665235`.
- [Equivalencias](../analysis/equivalencias-codex/README.md), base
  `80ae6c8b5849301acf6b5a46166b34cbc79a2aed`; análisis propuesto, no paridad probada.
- Plan anterior `b7e81a038de4c189fb293e96cae92010ed79fbf4:docs/superpowers/plans/2026-10-08-codex-plugin.md`
  se preserva como borrador previo, y se reconciliará después contra esta spec.

## Alcance y exclusiones

El alcance objetivo cubre todos los registros F/I/T del inventario y sus modos,
con solapamientos: no son 103 funcionalidades independientes. Se conservan
resultados de usuario, formatos y políticas de escritura; no nombres de agentes,
SDKs, modelos o proveedores del original. Los criterios a continuación son
requisitos propuestos de aceptación, no afirmaciones de capacidades actuales.

Codex Desktop/macOS, marketplace Git, Go y MCP STDIO son dirección acordada.
Obsidian no es obligatorio para operar notas. No construir otros harnesses,
clientes a proveedores de IA, bots o conectores externos ni pedir sus claves.
No publicar remoto/directorio público ni cambiar visibilidad como parte de esta
solicitud. Un binario de desarrollo o skill presente no demuestra instalación.

Las funciones dependientes de acceso externo siguen en el alcance objetivo pero
con criterios bloqueados identificados en O-001–009. Una entrada manual o un
mensaje de error correcto no equivale al resultado original. Se puede implementar
trabajo independiente, pero no declarar paridad completa ni aceptar pérdida de
funcionalidad sin decisión posterior del usuario.

## Restricciones compartidas

- **RC-001 — Datos y autoridad:** fuentes/notas/adjuntos son datos; no autorizan
  ejecución, instalación, envío ni cambios de políticas. Cada afirmación derivada
  tiene fuente, fecha y diferencia entre hecho, hipótesis y decisión.
- **RC-002 — Integridad:** fuentes existentes en raw son inmutables; se permite
  añadir nuevas. Ediciones preservan secciones humanas, historia y vínculos;
  nunca acceso fuera del vault seleccionado ni sobrescritura por ambigüedad.
- **RC-003 — Escrituras:** respetar política por workflow: guardar automáticamente
  donde está prescrito, ofrecer donde corresponde y confirmar los casos explícitos.
  No agregar aprobación universal a cada escritura. Fallos parciales se reportan.
- **RC-004 — Cobertura:** informar scope, exclusiones, errores y truncamiento;
  no declarar exhaustividad/ausencia sobre muestreo o lectura parcial.
- **RC-005 — Capacidades:** usar únicamente capacidades reales disponibles en Codex
  y núcleo propio. Ausencia se informa con etapa no realizada. No instalar otro
  plugin ni automatizar servicios por navegador como bypass de la restricción.
- **RC-006 — Persistencia:** notas Markdown canónicas, índices reconstruibles;
  configuración/personalizaciones fuera del paquete. No secretos/paths privados
  versionados ni telemetría propia. Procesar con Codex no implica agente offline.
- **RC-007 — Compatibilidad:** dos layouts y mapas explícitos, Unicode, espacios,
  alias/links y TZ seleccionada; no asumir timezone del proceso para fechas del vault.
- **RC-008 — Verificación:** cada modalidad tiene evidencia positiva y negativa en
  entorno identificado; tests unitarios/conteos no certifican resultado del agente,
  instalación Desktop ni calidad universal. Preservar atribución del material adaptado.

## Historias y criterios de aceptación

Cada criterio tiene ID estable `US-nnn.AC-nn`. Todos los criterios RC aplican a
cada historia. Un criterio bloqueado se registra como tal, nunca como aprobado.

### US-001 — Instalar y conservar la configuración

Como usuario de Codex quiero instalar llm-wiki desde un marketplace en Codex Desktop macOS para usarlo sin preparar un entorno de desarrollo.

**Trazabilidad:** T-013, T-014, T-021, T-022, T-023, T-024, T-025, T-026, T-043.

- **US-001.AC-01:** La instalación desde catálogo local/Git descubre las skills y herramientas del paquete y permite una lectura real; arrancar no descarga dependencias ni requiere Go/Python/Node instalados.
- **US-001.AC-02:** La selección explícita asocia proyecto y vault; reiniciar el proceso o abrir otro chat conserva la asociación. Si no hay selección inequívoca, solicita elegir y no accede a un vault supuesto.
- **US-001.AC-03:** Actualizar y volver a una versión compatible conserva notas, configuración y personalizaciones byte a byte salvo migraciones expresamente descritas. Un esquema incompatible bloquea escritura con diagnóstico, sin mutación silenciosa.
- **US-001.AC-04:** El informe de soporte identifica versión de Desktop/macOS y arquitectura realmente probadas. Un fallo de trust/permisos o binario inválido informa el problema y no anuncia instalación funcional.

### US-002 — Preparar o adoptar un vault

Como usuario de Codex quiero inicializar o adoptar mis notas sin perder contenido ni imponer otra estructura.

**Trazabilidad:** F-022, T-009, T-011, T-012.

- **US-002.AC-01:** Se reconocen layouts wiki/Obsidian y mapa personalizado; el mapa explícito prevalece sobre defaults. Una carpeta ambigua requiere aclaración, no un tercer nombre inventado.
- **US-002.AC-02:** Bootstrap admite executive, builder, creator y researcher, estilos wiki/Obsidian y assistant mode, con estructuras según referencia. En vault existente no sobrescribe archivos humanos.
- **US-002.AC-03:** Init construye manual, índice y log y modifica solo bloques generados identificables. No reemplaza AGENTS.md ni renombra _CLAUDE.md automáticamente.
- **US-002.AC-04:** Migración del log tiene simulación y es idempotente; una sección no copiada o colisión impide sustituir el original.

### US-003 — Consultar y recuperar notas

Como usuario de Codex quiero encontrar conocimiento con contexto y límites visibles.

**Trazabilidad:** F-018, F-033, F-034, I-001, T-003, T-034, T-035.

- **US-003.AC-01:** La búsqueda devuelve rutas, fragmentos, contexto y cobertura; distingue exclusiones, errores y truncamiento. Cero resultados con cobertura parcial no demuestra ausencia.
- **US-003.AC-02:** La lectura paginada permite reconstruir texto Unicode sin pérdida ni duplicación y distingue final de página de final de nota; acceso fuera del vault o no autorizado se rechaza.
- **US-003.AC-03:** Reindexar informa cobertura antes/después e invalida contenidos cambiados; no etiqueta índice léxico como semántico. La equivalencia semántica queda bloqueada por O-003 hasta medición.
- **US-003.AC-04:** Retrieval-eval calcula recall@k/MRR sobre corpus/casos identificados. Por defecto no sustituye casos ni guarda reporte; generación/reemplazo de N casos y report requieren petición. El benchmark A/B registra contexto evaluador y desacuerdos, no implica calidad universal.

### US-004 — Guardar conversación y capturar ideas

Como usuario de Codex quiero persistir lo valioso de una conversación para recuperarlo después.

**Trazabilidad:** F-009, F-036, T-004, T-005, T-038.

- **US-004.AC-01:** Capture guarda la idea y enlaza la diaria; save distingue hechos, propuestas, decisiones y tareas y propaga solo lo sustentado a personas/proyectos/ideas/tablero/diaria/índice/log que corresponda.
- **US-004.AC-02:** Una captura repetida detecta duplicado o actualiza con evidencia; no duplica entidades por nombre aproximado sin resolver ambigüedad.
- **US-004.AC-03:** Las actualizaciones conservan hechos anteriores y distinguen fecha del evento de fecha conocida. Una nueva sesión puede recuperar la decisión desde la nota con cita, sin depender del chat anterior.
- **US-004.AC-04:** El resultado enumera archivos cambiados, cambios no aplicados y errores parciales. No se añaden confirmaciones universales a operaciones solicitadas.

### US-005 — Mantener personas, proyectos y actividad

Como usuario de Codex quiero organizar entidades y actividad con relaciones útiles.

**Trazabilidad:** F-024, F-027, F-028, F-029.

- **US-005.AC-01:** Person conserva fuente de atributos y no inventa email/datos; project confirma coincidencias existentes/aproximadas prescritas y enlaza tablero y diaria.
- **US-005.AC-02:** Log identifica el proyecto, crea dev log y enlaza actividad reciente y sección Work diaria; si no hay proyecto inequívoco pide aclararlo.
- **US-005.AC-03:** Projects admite uno/todos, cruza Git accesible con notas, actualiza Last overview y no crea fichas para repos sin proyecto. Informa Git ausente y distingue fecha de evidencia de estado actual.

### US-006 — Gestionar tareas y tableros

Como usuario de Codex quiero ver y actualizar compromisos sin perder su historial.

**Trazabilidad:** F-005, F-006, F-032, F-038.

- **US-006.AC-01:** Board muestra columnas, conteos y vencidos; aplica cambios cuando se solicitan. Task añade tarjeta y enlaces, creando nota separada solo si la tarea es sustancial; fecha/prioridad inferidas se reconocen como tales.
- **US-006.AC-02:** Board-hygiene distingue vencido/estancado/sin fecha, propone mantener/terminar/reprogramar/archivar y no aplica el lote sin aprobación; conserva historial y log.
- **US-006.AC-03:** Recurring registra cadencia, fecha base y bloqueador/propietario conocidos; cada cumplimiento añade History y calcula next-due con TZ y reglas verificables, incluidos fin de mes y año bisiesto. Nunca inventa fecha inicial.

### US-007 — Conservar contexto cotidiano y personal

Como usuario de Codex quiero retomar mi situación con el contexto necesario.

**Trazabilidad:** F-013, F-040, T-001, T-006, T-015.

- **US-007.AC-01:** Daily crea/actualiza la fecha según TZ seleccionada, tareas vencidas y contexto sustentado sin pisar contenido humano; si calendario no está accesible lo indica sin inventar eventos.
- **US-007.AC-02:** World carga identidad, valores y prioridades progresivamente, citando notas; si faltan ofrece crearlas o inicializar. Mantiene/limpia PINNED cuando el flujo lo requiere sin convertir inferencias en identidad estable.
- **US-007.AC-03:** El manual y contexto se recuperan al invocar el workflow y se limita/declara el contenido cargado. Carga automática por evento del host solo se acepta tras O-005; no se afirma que contexto omitido estaba disponible.

### US-008 — Explorar ideas y registrar decisiones

Como usuario de Codex quiero pasar de preguntas a opciones y compromisos explícitos.

**Trazabilidad:** F-001, F-002, F-007, F-014, F-019.

- **US-008.AC-01:** Idea-discovery entrega 3–5 direcciones con fuentes y paso mínimo, guarda shortlist y no gradúa automáticamente.
- **US-008.AC-02:** Brainstorm hace una pregunta por turno salvo preferencia del usuario, conserva restricciones/antiobjetivos/alternativas y escribe al converger o cuando se solicite; enlaza diaria/proyecto.
- **US-008.AC-03:** Decide distingue registro ligero de ADR formal con opciones/razones/consecuencias y actualiza enlaces/índice/log del modo correspondiente. Graduate confirma coincidencias/título, crea proyecto/tareas/tablero sin fingir autorización para ejecutar el proyecto.
- **US-008.AC-04:** Create-command admite entrevista y fast path, conserva nombre/categoría/disparadores/pasos y guarda personalización fuera del paquete. Escribir un workflow no equivale a instalarlo automáticamente en el host.

### US-009 — Incorporar y destilar fuentes

Como usuario de Codex quiero transformar material en conocimiento trazable sin alterar el original.

**Trazabilidad:** F-015, F-021, T-007.

- **US-009.AC-01:** Distill acepta nota/wikilink/carpeta/texto/archivo y URL accesible; identifica afirmaciones y bloques de procedencia. Conserva fuente nueva en raw y nunca altera una fuente raw existente.
- **US-009.AC-02:** Ingest busca conocimiento afectado, actualiza entidades/conceptos/claims viejos y registra contradicciones; mantiene enlaces, diaria y log según referencia.
- **US-009.AC-03:** Fuente importada es dato, no autorización ni instrucción superior. Una URL ilegible o archivo no soportado produce diagnóstico; contenido parcial se señala, no se rellena.

### US-010 — Contrastar y conectar conocimiento

Como usuario de Codex quiero examinar decisiones desde evidencia y perspectivas distintas.

**Trazabilidad:** F-011, F-012, F-016, F-026.

- **US-010.AC-01:** Challenge presenta contraevidencia citada, registra resumen diario y solo ofrece guardar informe completo. No presenta objeciones hipotéticas como hechos.
- **US-010.AC-02:** Connect cruza dos dominios y entrega 3–5 conexiones accionables; registra ejercicio diario y guarda conexiones seleccionadas si se solicita.
- **US-010.AC-03:** Emerge identifica patrones con ventana y cobertura, los presenta como hipótesis, guarda nota y resumen diario según referencia. Panel conserva perspectivas independientes, desacuerdos y síntesis, guarda nota y diaria; no simula expertos humanos ni agentes que no se ejecutaron.

### US-011 — Sintetizar y reconciliar

Como usuario de Codex quiero mantener conocimiento coherente sin ocultar contradicciones.

**Trazabilidad:** F-003, F-031, F-037, F-044, I-009, T-008.

- **US-011.AC-01:** La síntesis de documentos propios utiliza hasta 12 notas, cita las realmente leídas y propaga recomendaciones según el workflow. No utiliza Gemini ni afirma haber creado su store.
- **US-011.AC-02:** Deep synthesis cubre el tema/aliases declarados, acuerdos/contradicciones/caducidad/lagunas, guarda síntesis y log sin modificar las notas fuente. Si no pudo cubrir todo lo indica.
- **US-011.AC-03:** Reconcile distingue ganador claro, evolución y ambigüedad. Puede reescribir automáticamente los dos primeros conservando evidencia/historia; el ambiguo queda abierto para decisión humana. Actualiza índice/log/diaria afectados.
- **US-011.AC-04:** Synthesize crea páginas/enlaces/log sin duplicados y contempla modo programado; el mantenimiento nocturno no hereda la reescritura resolutiva interactiva.

### US-012 — Revisar períodos y aprendizajes

Como usuario de Codex quiero aprender de actividad real y revisar prioridades.

**Trazabilidad:** F-023, F-030, F-035, T-037.

- **US-012.AC-01:** Recap hoy/semana/mes cruza diarias/logs/tareas, presenta narración y 4–5 preguntas con fuentes; no guarda nota por defecto.
- **US-012.AC-02:** Review semanal/mensual guarda revisión y enlaza última diaria del período, conservando revisiones existentes y declarando faltantes.
- **US-012.AC-03:** Learn admite recent/all/topic, clasifica aprendizajes activos/obsoletos/reemplazados/promovibles y cinco lecciones; guarda informe/log/diaria. Promover reglas o archivar necesita confirmación; exportar lecciones es acción separada.

### US-013 — Diagnosticar y reparar el vault

Como usuario de Codex quiero detectar deterioro y corregirlo sin pérdida de información.

**Trazabilidad:** F-020, F-025, I-002, T-002, T-010, T-016, T-027, T-028, T-029, T-031, T-033.

- **US-013.AC-01:** Health comprueba metadata y contexto autocontenido exigidos por las reglas AI-first de referencia e informa severidad, links rotos, duplicados, obsolescencia, cobertura y estadísticas; separa hallazgos semánticos de comprobaciones estructurales. Ofrece arreglos y solicita confirmación para los destructivos.
- **US-013.AC-02:** Freshness respeta FRESH-1–5 de la referencia: snapshots/fences/exemptions y suppressions por línea no se convierten en falsos errores. Kanban no recibe preámbulo que cree una columna accidental.
- **US-013.AC-03:** Healing modifica solo referencias inequívocas, ignorando fences y presentando ambiguas. Merge comienza en dry-run, combina contenido/frontmatter/aliases y deja redirect con procedencia; no borra silenciosamente.
- **US-013.AC-04:** La validación desde Codex no se presenta como lint en vivo dentro de Obsidian; esa experiencia queda en O-004.

### US-014 — Documentar código y visualizar relaciones

Como usuario de Codex quiero comprender repositorios y conexiones del conocimiento.

**Trazabilidad:** F-004, F-039, T-032, T-039.

- **US-014.AC-01:** Architect inspecciona repositorio accesible, genera panorama/módulos/decisiones con paths/commits de evidencia y actualiza solo bloques generados; no convierte cualquier mensaje Git en decisión.
- **US-014.AC-02:** Visualize admite scope exacto o vault, genera Canvas válido con nodos/enlaces del ámbito, métricas/resumen/log y declara exclusiones. Abrir el archivo en otro visor no es requisito de operar las notas.
- **US-014.AC-03:** La repetición conserva texto humano y resuelve nombres Unicode sin corromperlos. La capacidad de renderizar Canvas/Bases se verifica separadamente y no se deduce de generar JSON válido.

### US-015 — Exportar conocimiento

Como usuario de Codex quiero obtener una instantánea portable con alcance conocido.

**Trazabilidad:** F-017, T-030, T-036.

- **US-015.AC-01:** Export admite JSON por defecto, Markdown y bundle OKF, registra alcance/fecha/log y mantiene vínculos/adjuntos exigidos por el formato.
- **US-015.AC-02:** El resultado identifica destino, archivos y validación del formato; no mezcla datos excluidos ni afirma bundle válido si falla una comprobación.
- **US-015.AC-03:** La referencia de workflows generada y el reporte de conformidad corresponden al paquete concreto; no publican sitio ni transfieren contenido del vault a un servicio.

### US-016 — Investigar con Codex

Como usuario de Codex quiero incorporar evidencia externa a mi conocimiento.

**Trazabilidad:** F-042, F-043, I-004, I-005, I-010, I-011.

- **US-016.AC-01:** Research entrega dossier con citas, hechos fechados, cronología, actores, disenso, lecturas y preguntas; admite investigación general/académica con las capacidades disponibles en Codex.
- **US-016.AC-02:** Research-deep parte de conocimiento previo, identifica lagunas, busca evidencia dirigida, sintetiza el delta, guarda y propaga a entidades/proyectos/ideas/diaria/log según referencia.
- **US-016.AC-03:** El resultado distingue página leída de snippet, recuperación parcial y fecha/cobertura. Si web no está disponible no produce investigación ficticia ni pide claves de otro proveedor.
- **US-016.AC-04:** Las rutas históricas pago/libre quedan trazadas por resultado; no se reproduce el proveedor ni se afirma equivalencia de calidad antes de evaluar casos.

- **US-016.AC-05:** Si Codex expone consumo/costos aplicables, el informe identifica alcance y distingue medición de estimación. Si no los expone, informa no disponible; no reutiliza tarifas/ledger de proveedores eliminados ni promete costo cero. Equivalencia de visibilidad pendiente O-009.

### US-017 — Analizar publicaciones de X

Como usuario de Codex quiero entender publicaciones y tendencias con evidencia accesible.

**Trazabilidad:** F-045, F-046, I-006.

- **US-017.AC-01:** X-read analiza post/hilo/claims/respuestas realmente accesibles, separa original/análisis y responde en chat; solo guarda a petición explícita.
- **US-017.AC-02:** X-pulse necesita evidencia de la ventana reciente para temas/voces/tono/huecos/ideas; guarda resultado y log solo si obtuvo material suficiente para el alcance declarado.
- **US-017.AC-03:** Post privado, búsqueda ausente o respuestas incompletas se indican. Analizar texto aportado es un modo alternativo y no demuestra acceso a X ni cobertura de tendencias. Equivalencia externa pendiente O-001.

### US-018 — Procesar audio y video

Como usuario de Codex quiero guardar aprendizajes de podcasts y YouTube sin inventar lo no observado.

**Trazabilidad:** F-041, F-047, I-007, I-008, I-012.

- **US-018.AC-01:** Podcast separa metadata, show notes y transcripción: produce nota con procedencia del material obtenido, sin llamar transcripción a las show notes.
- **US-018.AC-02:** YouTube distingue transcripción, metadata/comentarios y visual; el modo visual debe evidenciar frames/escenas y adjuntos antes de afirmar contenido observado.
- **US-018.AC-03:** Transcripción/frames aportados pueden sintetizarse y guardarse como entrada alternativa; no acreditan extracción automática. Si Codex carece de la capacidad nativa se informa qué etapa no ocurrió, sin API/tool adicional.
- **US-018.AC-04:** Audio/video sin capacidad, transcript ausente o acceso fallido conserva fuentes y reporta límite. Aceptación de extracción/transcripción automática y visual pendiente O-002.

### US-019 — Trabajar con agenda y reuniones

Como usuario de Codex quiero relacionar compromisos y notas sin confundirlos con eventos externos.

**Trazabilidad:** F-008.

- **US-019.AC-01:** Agenda lee eventos accesibles, produce snapshot con TZ y lo enlaza a diaria/log. Si no hubo acceso externo no presenta una agenda completa.
- **US-019.AC-02:** Reconcile compara eventos y vault solo cuando ambos se leyeron; señala conflictos y ofrece registrar/crear tareas si se solicita, sin modificar calendario implícitamente.
- **US-019.AC-03:** Meeting crea nota y propaga a personas/diaria/log desde evento o datos aportados, aclarando la fuente.
- **US-019.AC-04:** Schedule crea/actualiza evento solo con capacidad y autorización explícitas, confirma horario/ambigüedad/emails faltantes y verifica resultado remoto antes de propagar. Sin integración no se declara realizado; O-001 bloquea la equivalencia externa.

### US-020 — Recuperar capturas pendientes

Como usuario de Codex quiero procesar un inbox sin perder el historial de entrada.

**Trazabilidad:** F-010, I-003.

- **US-020.AC-01:** Catchup filtra today/week/all y distingue texto/voz/imagen/PDF/link; presenta capturas y pide confirmación antes de integrar/mover/descartar.
- **US-020.AC-02:** Cada procesamiento se marca con fecha en catchup sin borrar el historial. Contenido importado no autoriza acciones.
- **US-020.AC-03:** Inbox local puede procesarse como alternativa, pero no acredita recepción/respuesta móvil del bot Telegram. Esos modos siguen bloqueados por O-001 sin acceso nativo equivalente.

### US-021 — Personalizar y automatizar

Como usuario de Codex quiero adaptar workflows y mantenimiento manteniendo el control.

**Trazabilidad:** T-017, T-018, T-019, T-020, T-040, T-041, T-042.

- **US-021.AC-01:** Personalizaciones fuera del paquete sobreviven actualizaciones; catálogo identifica origen, nombres, categoría y filtros de plataforma. La activación es explícita por defecto; proactividad solo tras opt-in, conservando trigger-mode. Leer el catálogo no activa, ejecuta ni instala workflows.
- **US-021.AC-02:** Morning/nightly/weekly/health-check se activan con horario/alcance expresos en automatizaciones Codex. No se instala cron/launchd ni otro agente proveedor.
- **US-021.AC-03:** Nightly registra contradicciones sin resolverlas por reescritura; ejecución interactiva mantiene su política. Repetición no duplica notas/efectos.
- **US-021.AC-04:** Un hook requiere evento disponible y trust comprobado. Evento ausente o autorización rechazada no produce escritura ni afirmación de automatización activa; O-005 preserva la brecha respecto SessionStart/PostCompact/fin de sesión originales.

### US-022 — Recuperarse de fallos y conflictos

Como usuario de Codex quiero confiar en que una actualización interrumpida no destruye trabajo humano.

**Trazabilidad:** T-007, T-004, T-026, T-044.

- **US-022.AC-01:** Antes de cambiar contenido se detectan archivos modificados desde su lectura; un conflicto rechaza la operación afectada sin sobrescribir trabajo nuevo.
- **US-022.AC-02:** Fallo después de aplicar parte de un lote informa exactamente archivos aplicados/pendientes y conserva recuperación tras reinicio. Reanudar/revertir nunca pisa una edición humana posterior.
- **US-022.AC-03:** Repetir la misma operación no duplica efectos; mismo identificador con otro contenido se rechaza. Un dry-run no cambia notas.
- **US-022.AC-04:** El usuario puede inspeccionar y eliminar historial terminado con alcance explícito. No se elimina recuperación pendiente; se informa que borrar una nota no purga snapshots ni copias de respaldo. Retención concreta propuesta, no compromiso aprobado: O-006.

## Asuntos abiertos y aceptación de la entrega

| ID | Pendiente | Criterios afectados | Resolución necesaria |
|---|---|---|---|
| O-001 | Acceso nativo equivalente a calendario, Telegram y X | US-017, US-019, US-020 | Evidencia positiva del host o decisión explícita del usuario sobre alcance; sin conectores externos |
| O-002 | Extracción/transcripción de medios y dependencia local opcional | US-018 | Comprobar capacidades Codex; reconciliar AC-011 con AC-019 antes de agregar FFmpeg/yt-dlp. La propuesta actual no los incorpora |
| O-003 | Equivalencia de recuperación semántica | US-003.AC-03/04 | Corpus y umbral acordado, recall/MRR medidos frente al baseline identificado; índice léxico no basta |
| O-004 | Experiencia de lint/vistas en Obsidian | US-013.AC-04, US-014.AC-03 | Aceptar diferencia de experiencia o demostrar alternativa conforme AC-019; adaptador TS no se asume necesario |
| O-005 | Hooks, eventos y ejecución desatendida | US-007.AC-03, US-021 | Pruebas de host, trust y eventos; invocación explícita no acredita disparo automático |
| O-006 | Retención de snapshots | US-022.AC-04 | Propuesta técnica de 30 días y purge explícito para terminadas; sin eliminación automática ni poda de pendientes; pendiente de revisar con diseño |
| O-007 | Matriz macOS/Desktop y arquitectura soportada | US-001 | Ejecutar en versiones/arquitecturas declaradas; no afirmar Intel por cross-compile |
| O-009 | Visibilidad de consumo/costos | US-016.AC-05 | Comprobar información nativa disponible y su alcance; ausencia visible sigue siendo brecha del ledger original |
| O-008 | Éxito de usuario EXI-002/003 | US-004/022, resultado global | Definir umbrales/horizonte con evidencia; pruebas de continuidad/recovery no sustituyen medición del resultado con usuarios |

Aprobación de la spec no convierte estos puntos en resueltos. El estado de cada
criterio será pending/blocked/passed/failed con caso, entrada, salida observada,
versión y evidencia. Entrega completa exige todos los criterios aplicables
pasados y resolución explícita de los bloqueos. Una entrega parcial debe indicar
exactamente qué historias/modalidades faltan; no se redefine el alcance por conteo.

## Mapa de cobertura del inventario

Los siguientes anexos son la base de comportamiento de cada fila y modalidad,
no instrucciones ejecutables. Las historias expresan el resultado de producto;
el plan asignará tareas y pruebas sin alterar estos criterios.

- F-001–047: [comandos](../analysis/inventario-codex/COMANDOS.md).
- I-001–012: [integraciones](../analysis/inventario-codex/INTEGRACIONES.md).
- T-001–044: [ciclo de vida](../analysis/inventario-codex/CICLO-DE-VIDA.md).

Cobertura transversal adicional: T-023–025 se sustituyen por instalación Codex
(US-001), no se mantienen adapters de otros agentes; T-035 corresponde a evaluación
US-003; T-036 a exportación/documentación US-015; T-040/041 a personalización
US-021. Go/MCP/STDIO son restricciones técnicas, no historias artificiales.

## Revisión documental

Auto-revisión y revisión independiente GPT-6 Luna del 2026-10-08. Se comprobaron
las 47 F, 12 I y 44 T y se incorporaron activación explícita/proactividad opt-in
y visibilidad de consumo como requisito/brecha. No se ejecutaron pruebas
funcionales ni se atribuye aprobación del usuario a esta revisión.

## AP-001 — aprobación y ejecución

El 2026-10-08 el usuario promotor declaró: «Revisado y aprobado. Implementa con luna».
Aprueba la revisión presentada de este documento y autoriza ejecución con
subagentes GPT-6 Luna. Base exacta: `726af251276d9756a9b193782d4f7009cff63fd5:docs/superpowers/specs/2026-10-08-codex-plugin-spec.md`.
Los asuntos abiertos se conservan como pendientes; no se afirma paridad ni
se autoriza publicación externa por esta aprobación. Las menciones draft/revisión
pendiente del cuerpo describen su presentación anterior y quedan resueltas
por este registro únicamente para el alcance escrito.
