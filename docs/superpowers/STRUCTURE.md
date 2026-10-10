---
title: Estructura de llm-wiki
status: draft
version: v006
responsible: Usuario promotor del proyecto
created_at: "2026-10-08"
updated_at: "2026-10-10"
reviewed_at: null
approval_reference: null
---

# Estructura de llm-wiki

Entrega relacionada: **v0.2.1**; runtime/manifiestos `0.2.1`.
[Changelog](../../CHANGELOG.md), [informe 0.2.1](../../tests/acceptance/release-0.2.1.json) y [aceptación Desktop](evidence/2026-10-10-desktop-areas.md).

Estructura vigente para `v0.2.1`: resolución e inventario de áreas, dos skills nuevas
y paquete universal verificado. Base histórica `43ee802`, ensayo 0.1.0-dev.2.
La dirección aplicable está en
[CONSTITUTION.md](CONSTITUTION.md), v004, y [ARCHITECTURE.md](ARCHITECTURE.md),
AC-019. Base previa: `f27672f7a63039805784efd8ccd5478b089dda2e`.
El inventario original se conserva intacto. No se utiliza `.repos` en ejecución.

## Organización implementada

| Ruta | Responsabilidad |
|---|---|
| `go.mod`, `go.sum` | Módulo `llm-wiki`, toolchain y dependencias fijadas; sin SDKs de proveedores |
| `cmd/llm-wiki/main.go` | Entrada `serve`, `doctor`, `vault`, `call`, `recover`; STDIO solo en serve |
| `internal/config/` | Registro de proyectos/vaults, versión de esquema, preferencias y rutas de estado |
| `internal/vault/` | Acceso confinado, lectura paginada, operaciones de notas y layouts |
| `internal/journal/` | Precondiciones, historial recuperable, bloqueo y reanudación de lotes |
| `internal/search/` | Recuperación léxica, cobertura e índice reconstruible; sin embeddings externos |
| `internal/graph/` | Wikilinks, aliases, backlinks, movimientos y exportación Canvas |
| `internal/health/` | Esquema, frescura y hallazgos; no decide verdad semántica |
| `internal/export/` | JSON, Markdown y OKF; validación del paquete exportado |
| `internal/service/` | Dispatch común a MCP/CLI; contexto, catálogo y operaciones del vault |
| `internal/packaging/` | Validación del paquete, inventario y checksums |
| `internal/version/` | Versión compartida por servidor y evaluación |
| `internal/mcp/` | Contratos JSON y traducción a operaciones comunes; SDK MCP Go |
| `internal/catalog/` | Descubrimiento de workflows del paquete y personalizaciones del vault |
| `cmd/package/main.go` | Construcción, validación y ensamblado de distribución; Go solo para desarrollo |
| `plugins/llm-wiki/plugin.json` | Identidad portable, metadata de versión y presentación |
| `plugins/llm-wiki/.codex-plugin/plugin.json` | Manifiesto de compatibilidad generado, si el host lo requiere |
| `plugins/llm-wiki/mcp.json` | Ejecuta binario empaquetado por STDIO, sin rutas personales |
| `plugins/llm-wiki/bin/llm-wiki` | Ejecutable macOS universal; nunca `go run` en la instalación |
| `plugins/llm-wiki/skills/llm-wiki/SKILL.md` | Reglas comunes, selección del vault y dispatch de workflows |
| `plugins/llm-wiki/skills/<workflow>/SKILL.md` | Workflow por resultado con política de escritura propia |
| `plugins/llm-wiki/references/` | Mapas, formatos, políticas y templates; atribución del material adaptado |
| `plugins/llm-wiki/capabilities.json` | Cada F/I/T y modalidad: implementación, evidencia y brechas |
| `.agents/plugins/marketplace.json` | Catálogo del proyecto, source local relativo para checkout Git |
| `tests/fixtures/` | Vaults sintéticos wiki/Obsidian, nunca datos personales |
| `tests/acceptance/` | Casos de workflows y evidencia de integración instalada |
| `docs/superpowers/analysis/equivalencias-codex/` | Matrices por registro original y límites de equivalencia |
| `docs/superpowers/plans/2026-10-08-codex-plugin.md` | Dependencias de tareas, pruebas y estado de ejecución |

## Límites de módulos

`mcp` y CLI dependen de los mismos servicios. `vault` no importa transportes,
proveedores ni capacidades del host. El razonamiento vive en skills y ocurre en
Codex. Las referencias importadas son datos, no autorizaciones. Personalizar un
workflow no modifica el paquete instalado: se conserva en el área del vault y
se invoca mediante el catálogo. Una skill local no se declara automáticamente
instalada globalmente en Codex.

## Archivos de usuario

La asociación proyecto-vault vive fuera del paquete. Dentro del vault se utiliza
`.llm-wiki/` para convenciones, workflows del usuario y un manual de continuidad.
Los índices y journals se guardan en el directorio de estado por vault definido
en INFRASTRUCTURE.md. No se sobrescribe AGENTS.md existente; el onboarding ofrece
un bloque delimitado o un enlace al manual, solo en el proyecto seleccionado.
El manual `_CLAUDE.md` se puede leer como convención heredada, sin importarlo como
instrucción superior ni renombrarlo automáticamente.

## Convenciones

- No código ejecutable propio Python/Node/shell. TypeScript no se requiere para
  este diseño Codex; la interfaz dentro de Obsidian permanece como brecha separada.
- Los tests viven junto a cada paquete; fixtures de integración en `tests/`.
- Toda operación que muta usa `journal`; lecturas y diagnósticos no escriben notas.
- Las dependencias de desarrollo (Go, lipo) no se instalan al arrancar el plugin.
- Las licencias y avisos de la referencia MIT se conservan para material adaptado.
- No se cambia `.gitignore` preexistente hasta que una tarea lo requiera y preserve
  su contenido; no se versionan caches, configuración personal ni fixtures reales.

Evidencia de pruebas y límites: [registro de implementación](evidence/2026-10-08-implementation.md). Desktop, modalidades y soporte Intel siguen pendientes.

## Cambios propuestos para áreas definidas por el usuario

La decisión canónica es [ARCHITECTURE.md](ARCHITECTURE.md), v008, ADR-009
(approved, AA-001). Esta sección localiza los cambios futuros; no los describe como
implementados ni autoriza ejecutarlos. La dirección aplicable permanece en
[CONSTITUTION.md](CONSTITUTION.md), v004, PRI-003–005 y OBJ-002–003.
Base exacta de archivos inspeccionados y constitución:
`e703e1f9d898ccc797cb2d7eeed1ee4e68a3ee04`.

| Ruta | Diferencia propuesta respecto de la base inspeccionada | Evidencia necesaria |
| --- | --- | --- |
| `internal/vault/layout.go` y pruebas cercanas | Reconocer tipo `area`, defaults por layout y mapa personalizado; no crear notas al resolverlo | Regresiones wiki/Obsidian, mapa explícito, ruta inválida y comportamiento previo sin áreas |
| `internal/vault/` | Ubicar la enumeración determinista y lectura limitada de candidatos a áreas junto al acceso confinado | Cuotas, cancelación, metadata malformada y cobertura incompleta; ubicación final de archivos pendiente de especificación |
| `internal/service/service.go`, `internal/mcp/` y consumidores CLI | Exponer el lector propuesto `wiki_list_areas` desde el servicio común | Contrato y errores equivalentes en CLI/MCP; no hay lector específico implementado en esta base |
| `plugins/llm-wiki/skills/wiki-obsidian-area/SKILL.md` (nuevo, propuesto) | Workflow explícito de crear, consultar y vincular áreas | Descubrimiento en catálogo y ejecución sobre vault sintético; presencia del archivo no acredita el resultado |
| `plugins/llm-wiki/references/folder-map.md` | Describir destino `area` respetando mapas existentes | Correspondencia con resolución del runtime y ausencia de migración implícita |
| `plugins/llm-wiki/references/write-rules.md`, `ai-first-rules.md` y referencia de áreas propuesta | Convenciones de relaciones y autorización conforme a ADR-009 | Enlaces del cuerpo recuperables; contenido humano y políticas de cada workflow preservados |
| `plugins/llm-wiki/skills/llm-wiki/SKILL.md` y skills que guardan contenido relacionado | Dispatch y reutilización opcional de áreas, sin categorías codificadas | Pruebas de activación explícita y de flujos anteriores sin áreas |
| `plugins/llm-wiki/capabilities.json`, catálogo/referencias derivados y `README.md` | Registrar la extensión y sus límites cuando exista entrega; conservar intacta la identidad de F/I/T de referencia | Estado acorde a evidencia por componente, workflow y host; no declarar paridad por añadir una skill |
| `tests/fixtures/`, `tests/acceptance/` y tests de paquetes afectados | Casos sintéticos y de continuidad definidos por ADR-009 | Resultados identificables por revisión y entorno; sin notas ni rutas personales |

### Organización propuesta del vault

Para estilo wiki sin mapa alternativo, después de una creación expresamente
solicitada podría existir `wiki/areas/<nombre-seguro>.md`. La nota relacionada
permanece en la carpeta de su tipo, con enlaces de áreas en el cuerpo. Son rutas
de ejemplo del vault del usuario, no archivos nuevos del repositorio ni una
plantilla con categorías fijas. En Obsidian o con mapa propio se usa el destino
resuelto por ADR-009.

El manual conserva la convención y los enlaces permiten recuperar relaciones;
no se introduce un registro paralelo de áreas en el directorio técnico ni una
copia por ámbito. El runtime sigue usando la persistencia y recuperación de
[INFRASTRUCTURE.md](INFRASTRUCTURE.md). No se añade un paquete ejecutable,
proveedor, daemon ni dependencia de instalación.

### Alcance de esta revisión y referencias relacionadas

Solo se presentan cambios documentales en arquitectura y estructura. Runtime,
skills, catálogo, manifiestos, binario y criterios de aceptación instalados no
se modifican. La spec y el plan existentes mantienen sus bases y aprobaciones;
la propuesta requiere adopción explícita antes de incorporarse a una entrega.
Las descripciones iniciales ligadas al paquete dev.2 conservan su carácter
histórico; esta inspección acotada no actualiza toda la evidencia de distribución.

### Historial de cambios

2026-10-09, v003: localiza la integración propuesta de ADR-009 y sus requisitos
de evidencia. Base anterior preservada:
`e703e1f9d898ccc797cb2d7eeed1ee4e68a3ee04:docs/superpowers/STRUCTURE.md`.
Estado draft, aprobación pendiente; no se altera el mapa de módulos implementados.

## Localización de la propuesta transversal

Extiende el mapa anterior con la integración de ADR-009 y la resolución de tipos
de ADR-010, ambos aprobados por AA-001 en [ARCHITECTURE.md](ARCHITECTURE.md), v008.
Fuente de esta revisión: selección del usuario de los cuatro puntos de la
verificación de carpetas. Base exacta:
`16da4583ab65a124d461cfdd41470020cb8a8073:docs/superpowers/STRUCTURE.md`.

| Ruta afectada en una futura implementación | Cambio localizado y límite |
| --- | --- |
| `internal/vault/layout.go`, `bootstrap.go` y tests | Resolución de OKR y entidades; filas canónicas/alias y precedencia de mapas de ADR-010. Preservar presets, notas existentes y carpeta wiki compartida |
| `plugins/llm-wiki/references/folder-map.md`, `write-rules.md`, `ai-first-rules.md` | Reglas transversales opcionales, rutas y enlaces según arquitectura; sin autoridad nueva sobre el vault |
| `plugins/llm-wiki/skills/wiki-obsidian-project/`, `wiki-obsidian-projects/`, `wiki-obsidian-person/`, `wiki-obsidian-capture/`, `wiki-obsidian-task/`, `wiki-obsidian-save/` | Consulta/reutilización de áreas durante las acciones autorizadas; conservar deduplicación y políticas actuales |
| `plugins/llm-wiki/skills/wiki-obsidian-calendar/` | Relación opcional de notas en modo meeting; no ampliar integración externa ni permisos de schedule |
| `plugins/llm-wiki/skills/wiki-obsidian-daily/`, `wiki-obsidian-log/`, `wiki-obsidian-review/`, `wiki-obsidian-synthesize/`, `wiki-obsidian-distill/` | Relaciones sustentadas en entradas y notas derivadas, sin diarios por área ni modificación de raw |
| `plugins/llm-wiki/skills/wiki-obsidian-okr/SKILL.md` (nuevo, propuesto) y referencias de plantilla | Workflow de objetivos/resultados explícitos; valores desconocidos pendientes, sin métricas o fechas inventadas |
| `internal/graph/` y pruebas cercanas | Verificar enlaces múltiples en el cuerpo y backlinks; no se propone ampliar el parser a frontmatter |
| `internal/catalog/`, metadatos de capacidades y referencias derivados | Descubrimiento y trazabilidad de extensiones; conservar F/I/T originales y separar entrega de validación por modalidad |
| `tests/fixtures/`, `tests/acceptance/` y tests de paquetes afectados | Casos wiki/Obsidian, precedencia, tarjetas, raw, múltiples áreas, repetición, conflictos y continuidad de arquitectura v008 |

La matriz de arquitectura fija comportamiento y decisiones; esta tabla solo
indica ubicación. Los directorios listados existen salvo las skills declaradas
nuevas. No se generaron skills, plantillas, fixtures ni código del producto en
esta tarea. Un área no crea subcarpetas en los destinos anteriores.

Constitución v003 e infraestructura v002 conservan sus bases. Spec/plan y
evidencias previas no se repuntan a este borrador ni se amplía AP-001.
2026-10-09, v004: añade localización del alcance transversal y referencias a
ADR-010, con estado draft y aprobación pendiente.

### Aprobación del alcance propuesto

AA-001, en [ARCHITECTURE.md](ARCHITECTURE.md), registra la aprobación explícita
del usuario del 2026-10-10 sobre esta v004 presentada en
`1b4b6e00bfc4ea939140007196cfe820b04e8d38:docs/superpowers/STRUCTURE.md`,
secciones de cambios propuestos y localización transversal. Se conserva la
revisión y el estado global draft: esta aprobación no cubre todos los pendientes
históricos ni describe código nuevo como implementado. Se autoriza preparar
spec y plan, pendientes de su propia revisión. El registro canónico de la decisión
es AA-001; las menciones previas a aprobación pendiente son históricas.

## Estado de implementación del incremento de áreas — 2026-10-10

AA-002 aprobó spec/plan v001 y ejecución directa. La versión asignada por el
usuario es 0.2.0. Rutas: `Layout.ResolveFolder` con errores de mapa; inventario:
`internal/vault/areas.go`, lector público `wiki_list_areas` en el servicio compartido.
Skills `wiki-obsidian-area` y `wiki-obsidian-okr`, referencia `areas.md` e integración
opcional en doce workflows. Sin cambios de parser de enlaces ni nuevos journals.

Las secciones de propuesta anteriores conservan su contexto histórico. La
[evidencia de entrega](evidence/2026-10-10-areas.md) distingue pruebas del runtime,
catálogo, propuestas de aplicación y criterios del host pendientes. No acredita
aceptación total, instalación ni dos chats por la existencia de estos archivos.

## Corrección descriptiva v0.2.1

Esquemas array en `internal/service/service.go`; regresiones de catálogo y lotes
STDIO en `internal/mcp/server_test.go`. Validación de ambos manifiestos en
`internal/packaging/`; lanzamiento y alternancia con identidad en
`tests/acceptance/package_test.go`. La evidencia Desktop 0.2.0 permanece histórica
y separada del [informe de paquete 0.2.1](../../tests/acceptance/release-0.2.1.json).
