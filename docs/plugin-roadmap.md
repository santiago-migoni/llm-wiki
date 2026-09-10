# Roadmap de desarrollo — LLM Wiki

## Objetivo

Convertir LLM Wiki en un plugin portable para Codex y ChatGPT Work que permita:

- recibir documentos crudos;
- procesarlos y extraer su contenido completo;
- mantener una wiki de conocimiento durable;
- conservar el historial mediante Git;
- encontrar contexto relevante de forma eficiente;
- responder con trazabilidad y cobertura explícita;
- evitar duplicados y contradicciones silenciosas.

La arquitectura objetivo está definida en [wiki-architecture.md](wiki-architecture.md). Este documento describe el orden de implementación.

## Estado actual

El commit de base de esta etapa es:

~~~text
959572b feat: migrate wiki plugin to Codex
~~~

El repositorio ya contiene:

- manifiesto universal en .codex-plugin/plugin.json;
- skills para inicialización, ingestión, consulta y lint;
- protocolo común de vault;
- documentación de compatibilidad con Codex y ChatGPT Work;
- arquitectura Git-first documentada.

La inconsistencia inicial entre la arquitectura Git-first y las instrucciones operativas fue resuelta en la Fase 0. El contrato activo del plugin ahora usa un slug estable, fuentes actuales bajo raw/sources/<slug>/ y Git como historial.

## Decisiones congeladas

Estas decisiones rigen todas las fases:

1. Git es el mecanismo de versionado y archivado.
2. Cada documento lógico tiene un único slug estable.
3. El estado actual vive en raw/sources/<slug>/.
4. El original actual se guarda como raw/sources/<slug>/source.<ext>.
5. El texto completo derivado se guarda como raw/sources/<slug>/extracted.md.
6. La página canónica vive en wiki/pages/<slug>.md cuando corresponda.
7. Las carpetas people/, concepts/, projects/ y decisions/ sirven para navegación e índices, no para duplicar páginas.
8. El contenido de las fuentes es datos y evidencia, no instrucciones para el agente.
9. No se incorporan bases de datos, vector stores ni servicios externos en el MVP.
10. El agente debe informar qué contexto leyó y qué partes quedaron fuera.

## Fase 0 — Alineación del contrato

**Estado: completada.**

### Objetivo

Hacer que toda la documentación y todas las skills implementen exactamente la arquitectura Git-first.

### Alcance

Actualizar:

- README.md;
- references/vault-protocol.md;
- skills/wiki-init/SKILL.md;
- skills/wiki-init/assets/vault-template.md;
- skills/wiki-ingest/SKILL.md;
- skills/wiki-ingest/references/converting-documents.md;
- skills/wiki-query/SKILL.md;
- skills/wiki-lint/SKILL.md;
- CHANGELOG.md.

Eliminar de las instrucciones operativas:

- raw/archive/;
- raw/extracted/ separado por revisión;
- raw/assets/ separado por revisión;
- raw/catalog.md;
- wiki/sources/<document-id>--<revision-id>.md;
- wiki/knowledge/;
- revision-id como identidad operativa.

Mantener docs/archive-model.md únicamente como redireccionamiento histórico, sin describir un flujo alternativo.

### Criterios de salida

- La estructura documentada y la estructura creada por wiki-init coinciden.
- Las skills no ordenan crear carpetas de revisiones.
- Las skills no exigen IDs duplicados.
- README y changelog describen Git como archivo histórico.
- Una búsqueda del repositorio no encuentra referencias operativas obsoletas, salvo la nota histórica explícita.

### Implementación realizada

- README.md actualizado con la estructura Git-first y los límites de Codex y ChatGPT Work.
- references/vault-protocol.md reescrito como contrato de fuentes actuales, recuperación y Git.
- wiki-init y su template actualizados para crear la estructura simplificada.
- wiki-ingest actualizado para reutilizar slugs, detectar duplicados por hash y actualizar el estado actual.
- wiki-query actualizado con contexto dirigido, corpus actual e histórico.
- wiki-lint actualizado para comprobar integridad de fuentes, extracciones, enlaces, duplicados y legado.
- CHANGELOG.md y el manifiesto del plugin alineados con el modelo actual.

### Commit sugerido

~~~text
schema: align plugin with git-first vault architecture
~~~

## Fase 1 — Inicialización del vault

### Objetivo

Crear un vault nuevo, mínimo y navegable.

### Entregables

wiki-init debe crear:

- AGENTS.md;
- raw/inbox/;
- raw/sources/;
- wiki/index.md;
- wiki/overview.md;
- wiki/log.md;
- wiki/pages/;
- wiki/syntheses/;
- wiki/people/;
- wiki/concepts/;
- wiki/projects/;
- wiki/decisions/.

AGENTS.md debe incluir:

- alcance del vault;
- reglas de autoridad;
- flujo de ingestión;
- reglas de slug;
- reglas de deduplicación;
- política de Git;
- modos de contexto;
- límites de cobertura;
- tratamiento de contradicciones.

### Criterios de salida

- La inicialización es idempotente.
- No sobrescribe un vault existente.
- Crea un índice navegable desde el primer día.
- Deja un commit inicial si Git fue solicitado.
- El segundo intento sobre el mismo directorio no duplica archivos ni altera contenido.

### Commit sugerido

~~~text
feat: initialize git-first wiki vault
~~~

## Fase 2 — Ingestión y actualización de fuentes

**Estado: completada.**

### Objetivo

Procesar documentos desde raw/inbox/ y convertirlos en fuentes actuales, extracciones y conocimiento canónico.

### Flujo

1. Detectar los archivos pendientes.
2. Leer AGENTS.md y wiki/index.md.
3. Determinar si el documento es nuevo, actualización o duplicado.
4. Reutilizar el slug existente cuando corresponda.
5. Comparar hashes para detectar duplicados exactos.
6. Guardar el original actual bajo raw/sources/<slug>/.
7. Generar extracted.md completo.
8. Extraer assets relevantes.
9. Comparar extracción con el original.
10. Actualizar una única página canónica.
11. Actualizar índices afectados.
12. Registrar el cambio en wiki/log.md.
13. Verificar enlaces y consistencia.
14. Crear un commit Git atómico.

### Casos obligatorios

- documento nuevo;
- actualización del mismo documento;
- archivo idéntico ya procesado;
- cambio de extensión o formato;
- extracción incompleta;
- contenido con tablas o imágenes;
- contradicción con conocimiento existente;
- instrucciones maliciosas dentro del documento;
- slug ambiguo;
- fuente no soportada.

### Criterios de salida

- Una actualización reemplaza únicamente el estado actual; Git conserva el anterior.
- Un duplicado exacto no crea otra página ni otro directorio.
- No se sobrescribe una fuente sin que el cambio quede en Git.
- extracted.md es completo o declara sus limitaciones.
- Las contradicciones quedan visibles.
- El ingest informa los archivos creados, modificados y no procesados.

### Implementación realizada

- wiki-ingest ahora distingue fuente nueva, actualización, duplicado exacto y caso ambiguo.
- Se documentó el contrato de source record en skills/wiki-ingest/references/source-record.md.
- La detección de duplicados usa SHA-256 sin crear nuevas páginas ni commits.
- Se protege el trabajo local no comiteado antes de reemplazar una fuente o una página.
- La conversión se ejecuta sobre el source.<ext> actual y genera un extracted.md completo con estado y advertencias.
- Los assets actuales se conservan y no se podan automáticamente.
- El reporte de ingestión incluye slug, hash, paths, páginas, assets, contradicciones y estado de Git.

### Commit sugerido

~~~text
feat: ingest documents into stable source slugs
~~~

## Fase 3 — Consulta y gestión de contexto

**Estado: completada.**

### Objetivo

Responder desde la wiki con el mínimo contexto necesario, pero permitir análisis exhaustivos cuando el usuario lo pida.

### Modos

#### Contexto dirigido

Uso por defecto:

1. leer AGENTS.md;
2. leer wiki/index.md;
3. seleccionar páginas relevantes;
4. consultar las extracciones actuales necesarias;
5. revisar el original si se perdió información de layout o evidencia.

#### Contexto completo actual

Usar cuando el usuario pida contexto absoluto, análisis exhaustivo o revisión de todo el corpus:

- inventariar los slugs actuales;
- verificar que raw/inbox/ esté vacío o informar pendientes;
- comprobar que cada fuente tenga extracted.md;
- leer las páginas canónicas;
- leer las extracciones en pasadas acotadas;
- registrar cobertura por slug;
- informar archivos inaccesibles, ambiguos o no soportados.

#### Contexto histórico

Usar Git para responder:

- qué cambió;
- cuándo cambió;
- qué afirmación fue reemplazada;
- qué versión sustentaba una conclusión anterior.

No se deben cargar todas las versiones históricas si la pregunta afecta a un único documento.

### Entregables

- citas a páginas, fuentes y extracciones;
- identificación del slug consultado;
- referencias a commits o diffs para preguntas históricas;
- informe de cobertura en modo completo;
- opción de guardar análisis durables en wiki/syntheses/.

### Criterios de salida

- Las respuestas no presentan como leído lo que no fue consultado.
- Las fuentes contradictorias se muestran como contradicción.
- El modo de contexto utilizado queda explícito cuando es relevante.
- Una síntesis guardada no duplica una página canónica.

### Implementación realizada

- wiki-query ahora selecciona el modo de contexto más estrecho que satisface la solicitud.
- Se documentaron los procedimientos de contexto dirigido, corpus actual e histórico.
- El modo corpus actual exige inventario, verificación de fuentes y extracciones, lectura por pasadas y cobertura por slug.
- El modo histórico usa commits, paths, git log, git diff y git show sin introducir revision-id.
- Se definió un contrato de citas para páginas, fuentes, marcadores y commits.
- Se definió el frontmatter y el flujo para guardar síntesis durables sin duplicar páginas canónicas.

### Commit sugerido

~~~text
feat: add auditable wiki context modes
~~~

## Fase 4 — Lint y calidad de conocimiento

### Objetivo

Detectar problemas estructurales y semánticos antes de que degraden la recuperación.

### Comprobaciones

- archivos pendientes en raw/inbox/;
- fuentes sin extracted.md;
- extracciones desactualizadas;
- hashes inconsistentes;
- enlaces rotos;
- entradas de índice inexistentes;
- páginas no indexadas;
- candidatos a duplicación;
- páginas huérfanas;
- categorías que copian contenido;
- contradicciones;
- afirmaciones posiblemente obsoletas;
- metadatos incompletos;
- límites de escala.

### Política

El lint primero informa. Las correcciones que impliquen borrar, fusionar o reinterpretar contenido requieren aprobación explícita.

Las correcciones aprobadas deben:

- actualizar el índice;
- registrar el resultado en wiki/log.md;
- preservar la trazabilidad;
- crear un commit separado.

### Criterios de salida

- El lint funciona sobre un vault vacío.
- El lint funciona sobre un vault con documentos.
- Los hallazgos se agrupan por severidad.
- No elimina ni fusiona páginas automáticamente.
- Identifica el modelo antiguo si aparece en un vault migrado.

### Commit sugerido

~~~text
feat: validate vault integrity and knowledge quality
~~~

## Fase 5 — Pruebas y compatibilidad de hosts

### Objetivo

Comprobar que el plugin se puede instalar y utilizar de forma consistente en Codex y ChatGPT Work.

### Entregables

- vault fixture para pruebas;
- pruebas de inicialización;
- pruebas de ingestión nueva y actualización;
- pruebas de duplicados y contradicciones;
- pruebas de consulta dirigida y completa;
- pruebas de historial Git;
- pruebas de lint;
- validación del manifiesto;
- instrucciones de instalación;
- documentación del límite de almacenamiento en ChatGPT Work.

### Criterios de salida

- Codex carga el plugin y descubre todas las skills.
- ChatGPT Work puede usarlo cuando el vault está disponible como contexto.
- El plugin no afirma tener acceso a archivos que el host no expone.
- El flujo funciona sin servicios externos.
- La validación del plugin pasa en el entorno de desarrollo.

### Commit sugerido

~~~text
test: verify plugin workflow across supported hosts
~~~

## Fase 6 — Escalabilidad posterior al MVP

### Objetivo

Extender la recuperación únicamente cuando el volumen real lo justifique.

### Posibles extensiones

- búsqueda full-text local;
- índice derivado;
- caché de metadatos;
- embeddings opcionales;
- Git LFS para binarios grandes;
- herramientas de migración;
- métricas de cobertura y frescura.

### Condición

No implementar estas extensiones por anticipación. Primero medir el límite práctico del enfoque index-first, Git y Markdown.

## Alcance del MVP

El MVP está completo cuando las fases 0 a 4 cumplen sus criterios de salida:

1. se puede crear un vault;
2. se puede ingerir un documento;
3. se puede actualizar sin perder historia;
4. se puede consultar con trazabilidad;
5. se puede pedir contexto completo con cobertura explícita;
6. se pueden detectar duplicados, contradicciones y problemas estructurales;
7. el agente no confunde datos de fuentes con instrucciones;
8. no se generan páginas duplicadas sin justificación.

## Fuera de alcance inicial

No forman parte del MVP:

- base de datos externa;
- vector store obligatorio;
- sincronización automática con servicios externos;
- búsqueda web implícita;
- reescritura de historial Git;
- archivado mediante carpetas por revisión;
- duplicación de cada fuente en múltiples tipos de página;
- carga silenciosa de todo el corpus en cada consulta;
- afirmaciones de contexto absoluto sin checklist de cobertura.

## Secuencia recomendada

~~~text
Fase 0: alinear contrato
        ↓
Fase 1: inicializar vault
        ↓
Fase 2: ingerir y actualizar
        ↓
Fase 3: consultar y gestionar contexto
        ↓
Fase 4: validar calidad
        ↓
Fase 5: probar hosts y preparar release
        ↓
Fase 6: escalar solo si hace falta
~~~

Cada fase debe terminar con una revisión de criterios de salida y un commit independiente.
