---
title: Estructura de llm-wiki
status: draft
version: v003
responsible: Usuario promotor del proyecto
created_at: "2026-10-08"
updated_at: "2026-10-09"
reviewed_at: null
approval_reference: null
---

# Estructura de llm-wiki

Estructura implementada en `43ee802`, con paquete de ensayo 0.1.0-dev.2. La dirección aplicable está en
[CONSTITUTION.md](CONSTITUTION.md), v003, y [ARCHITECTURE.md](ARCHITECTURE.md),
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

La decisión canónica es [ARCHITECTURE.md](ARCHITECTURE.md), v007, ADR-009
(proposed). Esta sección localiza los cambios futuros; no los describe como
implementados ni autoriza ejecutarlos. La dirección aplicable permanece en
[CONSTITUTION.md](CONSTITUTION.md), v003, PRI-003–005 y OBJ-002–003.
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
