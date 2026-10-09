# Equivalencias de llm-wiki centrado en Codex

Análisis propuesto del 2026-10-08, ejecutado por tres subagentes GPT-6 Luna y
revisado por el coordinador. No es evidencia de implementación ni de paridad.
Dirección aprobada: CONSTITUTION v003 y AC-019 de ARCHITECTURE, base
`f27672f7a63039805784efd8ccd5478b089dda2e`. Inventario de referencia: commit
`22fd2d36ea33415a8f44211ae346f2a775021553`, fuente obsidian-second-brain v0.17.0
`b0089f7666d0c1e3db41c9d587b6f3c997665235`.

## Documentos y cobertura

- [Comandos](COMANDOS.md): 47 registros F-001–F-047, con modos y efectos.
- [Integraciones y ciclo](INTEGRACIONES-Y-CICLO.md): 12 I y 44 T.
- [Empaquetado](EMPAQUETADO.md): catálogo Git, manifiestos, Go/macOS y comprobaciones.
- [Verificación estructural](VERIFICACION.json): recuento automático de IDs.

Los grupos se solapan: no sumarlos como funcionalidades únicas ni como porcentaje
de paridad. Cada fila mantiene resultado, propuesta, brecha y evidencia pendiente.

## Decisiones técnicas recomendadas

1. Skills de Codex para razonamiento/workflows; Go para operaciones deterministas
   del vault, MCP STDIO y herramientas internas. Sin runtime propio Python/Node.
2. Lint en Go, formatos Obsidian conservados; sin adaptador de editor en esta
   implementación centrada en Codex. La experiencia en vivo no está reemplazada.
3. Sin claves, adaptadores de proveedores, Telegram bot, Calendar MCP, Ollama,
   FFmpeg o yt-dlp. La exclusión es coherente con AC-019, pero deja resultados
   originales cuya equivalencia no está demostrada.
4. Binario universal macOS precompilado, sin descargas en arranque. Instalación
   desde Git y pruebas reales de Desktop todavía pendientes.

## Brechas que impiden afirmar cobertura completa

| Capacidad original | Qué puede diseñarse dentro de Codex | Qué sigue sin demostrar |
|---|---|---|
| Calendar: agenda/reconcile/meeting/schedule | Notas de reunión y agenda a partir de datos aportados | Leer/modificar un calendario externo sin integración |
| Captura Telegram y catchup | Procesar un inbox local en Codex | Recepción móvil, polling del bot y respuestas remotas |
| Investigación X | Analizar texto o páginas públicas accesibles por Codex | Búsqueda dedicada, respuestas completas y ventana de tendencias |
| Podcast/YouTube visual | Síntesis de transcripción/frames disponibles | Extraer/transcribir audio, comentarios y escenas sin herramientas adicionales |
| Índice semántico | Índice léxico, links, consultas ampliadas por Codex | Recall por paráfrasis equivalente a los embeddings originales |
| Lint dentro de Obsidian y vistas | Validación Go y archivos Canvas/Bases | Vista/lint en tiempo real dentro de otro editor |
| Hooks y tareas desatendidas | Workflows explícitos y automatizaciones Codex autorizadas | Eventos/trust, entorno y ejecución real tras instalación |

Cambiar un proveedor no es por sí mismo perder una función: la síntesis de hasta
12 notas, antes Gemini File Search, puede rediseñarse como síntesis citada en Codex.
Lo que se evalúa es el resultado y sus límites, no reproducir el store de Gemini.
Igualmente, ofrecer texto manual no equivale a capturar un mensaje de Telegram.

La etapa implementable localmente puede avanzar sin declarar resueltas estas
brechas. Cualquier reducción definitiva del objetivo completo necesita decisión
explícita; no se considera aceptada por aprobar una implementación incremental.

## Evidencia del host y límites

Se inspeccionó el catálogo de herramientas de esta sesión: web y automatizaciones
Codex están disponibles; Apple Calendar aparece como plugin independiente, por lo
que no se usa como capacidad nativa. No se identificó transcripción general nativa.
Esta observación no demuestra ausencia universal ni disponibilidad en otras cuentas.
Go no está en PATH; compilar exige preparar un toolchain de desarrollo. No se
instaló nada, no se modificó el vault personal ni la referencia. La ayuda de CLI
inspeccionada no sustituye una instalación Desktop.

## Handoff

Diseño: [ARCHITECTURE](../../ARCHITECTURE.md), [STRUCTURE](../../STRUCTURE.md),
[INFRASTRUCTURE](../../INFRASTRUCTURE.md).
Especificación funcional: [spec v001](../../specs/2026-10-08-codex-plugin-spec.md).
Planificación (draft, dependiente de la spec): [plan](../../plans/2026-10-08-codex-plugin.md).
