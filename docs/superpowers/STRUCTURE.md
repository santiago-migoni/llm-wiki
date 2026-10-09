---
title: Estructura de llm-wiki
status: draft
version: v002
responsible: Usuario promotor del proyecto
created_at: "2026-10-08"
updated_at: "2026-10-08"
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
