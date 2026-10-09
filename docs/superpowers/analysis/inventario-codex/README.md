---
title: Inventario funcional y diagnóstico de adaptación a Codex
status: draft
version: v001
created_at: "2026-10-08"
updated_at: "2026-10-08"
approval_reference: null
---

# Inventario funcional y diagnóstico de adaptación a Codex

El plugin contiene 47 comandos, 12 herramientas MCP y capacidades adicionales de
contexto, automatización, investigación, integración y mantenimiento. Conservarlo
completo exige trasladar esas capacidades además de sus instrucciones. El build
actual recomendado para Codex genera 46 comandos y una skill de soporte: excluye
calendario y omite componentes que algunos scripts incluidos necesitan.

La adaptación resulta viable como dirección de trabajo, pero la paridad completa
no está demostrada. La principal dificultad es conservar el comportamiento entre
sesiones, proveedores, herramientas y escrituras coordinadas. Este análisis no
selecciona arquitectura, lenguaje, release ni calendario de implementación.

## Base y alcance

- Dirección aprobada: [Constitución v001](../../CONSTITUTION.md), base exacta
  `eaa8e3075c482dc3fd9cb9864b50b34e668d97a0:docs/superpowers/CONSTITUTION.md`.
  Aplican PRI-001–PRI-005, OBJ-001–OBJ-003 y EXI-001–EXI-003.
- Fuente inspeccionada: `.repos/obsidian-second-brain`, commit
  `b0089f7666d0c1e3db41c9d587b6f3c997665235`, versión declarada `0.17.0`.
- Su remoto real es `https://github.com/santiago-migoni/fork.eugeniughelbur-obsidian-second-brain.git`.
  La constitución identifica el proyecto de origen `eugeniughelbur/obsidian-second-brain`;
  este diagnóstico corresponde específicamente al fork local y al commit indicado.
  No se verificó equivalencia con la versión actual del upstream remoto.
- Inspección estática repartida entre tres subagentes GPT-6 Luna, con contraste
  posterior de efectos de escritura, dependencias y salidas generadas.
- Verificación ejecutada: build en copia temporal del commit y dos pruebas de ayuda
  CLI sin datos de usuario. No instalación, activación, vault real ni llamadas a
  proveedores. Las pruebas existentes se inspeccionaron selectivamente, no se ejecutó
  su suite. Esto no es una auditoría de seguridad.

## Cómo leer el inventario

| Anexo | Cobertura | Uso |
|---|---|---|
| [Comandos](COMANDOS.md) | F-001–F-047, una fila por cada archivo de comando | Resultados, modos, dependencias y efectos de cada flujo |
| [Integraciones](INTEGRACIONES.md) | I-001–I-012, familias de integración e investigación | MCP, Obsidian, Telegram, proveedores, medios y guardado |
| [Ciclo de vida](CICLO-DE-VIDA.md) | T-001–T-044, reglas, utilidades y operación | Contexto, presets, hooks, mantenimiento, instalación y evaluación |
| [Fuentes](FUENTES.json) | 366 archivos versionados con hash y clasificación | Delimitación reproducible de la fuente inspeccionada |
| [Verificación](VERIFICACION.json) | Comandos, salidas y códigos de retorno | Evidencia del build y del fallo de dependencia reproducido |

Los 103 registros F/I/T se solapan deliberadamente: un comando puede usar una
integración y varias reglas. No representan 103 funcionalidades independientes ni
una medición de paridad. Los IDs permiten seguir cada capacidad al diseñar y probar
la adaptación. Las sugerencias de la última columna son propuestas, no cambios
aprobados al comportamiento original.

La cobertura es completa respecto del catálogo de 47 comandos y de las familias
identificadas en esta base. No significa que todos los casos internos, errores o
resultados posibles hayan sido ejercitados. Los archivos de sitio, ejemplos, tests
y mantenimiento del repositorio se clasifican separadamente y no añaden funciones
de usuario por el mero hecho de existir.

## Mapa funcional

| Familia | Capacidades que deben conservarse | Registros principales |
|---|---|---|
| Preparación y personalización | Inicializar vault existente o nuevo, cuatro presets, layouts wiki/Obsidian, modo asistente, manual y creación de comandos propios | F-001, F-022; T-009–T-014 |
| Captura y organización | Conversación, ideas, personas, proyectos, decisiones, tareas, obligaciones recurrentes, diaria, devlog y tableros | F-005–F-009, F-013–F-014, F-019, F-024, F-027–F-029, F-032, F-036, F-038 |
| Consulta y continuidad | Búsqueda con evidencia, contexto progresivo de identidad/prioridades, recap y estado cruzado con Git | F-018, F-029–F-030, F-040; I-001; T-005–T-006 |
| Pensamiento y aprendizaje | Brainstorm, crítica, panel, descubrimiento de ideas, conexiones, patrones, revisión de aprendizajes y revisiones periódicas | F-002, F-007, F-011–F-012, F-016, F-023, F-026, F-035, F-037, F-044 |
| Ingestión y evolución | Fuentes originales, destilación, actualización multientidad, reconciliación, síntesis y fusión con redirecciones | F-015, F-021, F-025, F-031, F-037, F-044; T-004–T-008, T-029 |
| Investigación externa y corpus propio | Web general/académica, brechas respecto del vault, X, YouTube con frames, podcast y Gemini File Search | F-003, F-041–F-043, F-045–F-047; I-004–I-012 |
| Calendario y captura remota | Agenda, reconciliación, reuniones, programación/reprogramación de eventos; texto, voz, imágenes, PDF y enlaces por Telegram | F-008, F-010, F-013; I-003 |
| Representación e intercambio | Canvas/grafo, JSON/Markdown/OKF, notas de arquitectura y extracción de decisiones de commits | F-004, F-017, F-039; T-032–T-033 |
| Calidad y mantenimiento | Salud, frescura, reparación de enlaces, migración de logs, estadísticas, índice semántico y evaluaciones | F-020, F-025, F-033–F-034; I-002; T-027–T-035 |
| Operación automática | Carga inicial, recall, validación post-escritura, propagación tras compactación y cuatro rutinas programadas | T-015–T-020 |
| Distribución y evolución | Instalar, configurar, actualizar, restaurar respaldo, generar skills y comprobar su consistencia | T-013–T-014, T-021–T-026, T-036, T-041, T-043 |

### Rutinas programadas

Estas rutinas se describen en `SKILL.md`, sección Scheduled Agents; no son cuatro
ejecutables autónomos instalados por el build. Los horarios son ejemplos de la
referencia, no programación solicitada ni creada en este proyecto.

| Rutina | Resultado y límites del original |
|---|---|
| Morning | Diaria, pendientes/vencidos y proyectos activos sin actividad reciente |
| Nightly | Cierre del día, tareas completadas a Done, conflictos abiertos, síntesis, enlaces de huérfanas e índice/log; no resuelve contradicciones reescribiendo páginas durante la noche |
| Weekly | Recap semanal y nota de review enlazada desde la diaria |
| Health check | Informe de salud y severidad; no aplica correcciones autónomas |

La reconciliación interactiva F-031 sí reescribe automáticamente cuando identifica
un ganador claro o evolución; los casos ambiguos se dejan abiertos. Esta diferencia
con Nightly es parte del comportamiento a conservar, no una inconsistencia que
deba borrarse al unificar instrucciones.

## Diagnóstico de adaptación

| ID | Hallazgo y evidencia | Consecuencia |
|---|---|---|
| B-001 | `commands/obsidian-calendar.md` excluye `agent-skills` y todos los destinos no-Claude. Build comprobado: 46 comandos + `obsidian-core`, falta calendario | El paquete generado no cumple por sí solo la meta de conservar todas las capacidades; calendar exige un equivalente real con sus cuatro modos |
| B-002 | `_ask_emit_core` copia scripts/references/pyproject, pero no `integrations/`. `semantic_search.py:64` importa `vault_ops` desde esa carpeta. `--help` pasa en source y falla en el paquete con `ModuleNotFoundError` | F-033 y la cadena de evaluación que también importa `vault_ops` necesitan cierre de dependencias; compilar skills no demuestra que sus herramientas arranquen |
| B-003 | Build sin `hooks/`, integración MCP ni rutinas programadas; setup y hook PostCompact siguen invocando/configurando Claude | Contexto automático, validación, ejecución desatendida y MCP necesitan integración Codex propia y validación por superficie |
| B-004 | Conviven adaptador `codex-cli` deprecado, `agent-skills` recomendado y wrappers que lanzan `codex exec` | Hay tres rutas con supuestos diferentes; se debe acordar una experiencia de instalación/uso sin confundir nueva ejecución con la sesión actual |
| B-005 | Playbooks invocan otras instrucciones, subagentes y propagación; devolver texto con `obsidian_get_skill` no ejecuta el workflow | Trasladar solo el MCP no conserva el producto; deben verificarse también los pasos de agente y sus efectos |
| B-006 | Investigación y captura remota usan xAI, Perplexity, Google, OpenAI, Groq, Telegram y fuentes web según modo; `/notebooklm` es Gemini File Search, no NotebookLM de consumo | Exclusividad para Codex no significa exclusividad de proveedor ni funcionamiento offline; sustituir proveedores requiere demostrar equivalencia |
| B-007 | Writer de investigación usa carpetas `Research/*` y autoapertura Obsidian; otras rutas dependen del folder map. Validación existe en prompts, shell, Python y TypeScript | Hay riesgo de comportamiento divergente entre caminos de escritura/layout; definir equivalencia con fixtures de ambos estilos y conservar integraciones de Obsidian |
| B-008 | README MCP enumera 10 herramientas mientras `server.py` registra 12; contiene checklist antiguo de tres herramientas | El código y los cuerpos de comandos son la base del inventario; cantidades promocionales no sirven como prueba de cobertura |
| B-009 | Descripciones de Agent Skills se construyen con `triggers_en`, aunque fuentes contienen también español, alemán, portugués y chino | Conservar selección semántica multilingüe exige evaluación; no se demuestra que otros idiomas fallen, pero sus disparadores no se trasladan literalmente |
| B-010 | Scripts de setup/update presuponen instalación mutable; `create-command` escribe nuevos comandos y pide recompilar | La extensibilidad del usuario necesita conservarse sin depender accidentalmente de editar una caché instalada o perder personalizaciones al actualizar |

Los hallazgos B-001 y B-002 tienen comprobación ejecutada en copia temporal. Los
demás son diagnóstico estático de la base identificada. No se infiere exposición
de seguridad actual a partir de revisiones históricas ni se certifica seguridad por
la existencia de pruebas.

### Encaje actual de Codex

La documentación oficial contempla skills y MCP dentro de plugins, y hooks para
instalación manual en Codex desktop. Los hooks requieren confianza explícita y sus
scripts disponibles en el entorno; tener el mismo evento nominal no prueba que el
payload y la salida de un hook de Claude sean compatibles. La documentación también
excluye plugins con hooks del directorio público. Esto afecta la elección futura
del canal de distribución, no obliga a eliminar sus capacidades.
[Empaquetado oficial](https://developers.openai.com/plugins/build/plugins).

OpenAI describe la conversión de comandos y procedimientos a skills, conservando
helpers y reemplazando configuración específica de Claude. La conversión formal
no verifica resultados de usuario.
[Guía oficial de conversión](https://developers.openai.com/plugins/guides/submit-claude-plugin).

Las tareas programadas pueden invocar skills y plugins; el trabajo sobre archivos
locales requiere el equipo encendido y la aplicación activa. CLI e IDE no ofrecen
la interfaz de gestión Scheduled. Debemos delimitar la superficie soportada antes
de prometer equivalencia de mantenimiento desatendido.
[Tareas programadas](https://learn.chatgpt.com/docs/automations?surface=app).

Fuentes oficiales consultadas el 2026-10-08. Son condiciones documentadas, no una
validación de este plugin en la aplicación instalada.

## Comprobaciones para demostrar equivalencia

Estos escenarios son insumos propuestos para el diseño posterior, no una
especificación aprobada ni un plan de ejecución.

| Área | Comprobación propuesta |
|---|---|
| Instalación y actualización | Instalación limpia, rutas con espacios, cambio de carpeta, nuevo vault y vault existente; actualización conserva preferencias y comandos personalizados |
| Captura y propagación | Una conversación produce las notas, enlaces, diaria, índice y log prescritos; repetir no pierde texto humano ni inventa decisiones |
| Recuperación | Consulta entre sesiones, exacta y semántica, corpus mayor al límite, índice ausente/desactualizado y backend caído; declarar cobertura |
| Evolución | Diferenciar evolución temporal, contradicción inequívoca y conflicto ambiguo; conservar historial; comparar ejecución interactiva y nocturna |
| Escritura y recuperación | Archivos largos, aliases, UTF-8, modificaciones concurrentes y fallo parcial; verificar qué cambió y cómo recuperarlo |
| Investigación y medios | Modos pago/libre/académico, claves ausentes, proveedor caído, texto parcial, audio sin transcript y video con frames; procedencia y propagación correctas |
| Calendario | Cuatro modos, zona horaria, asistentes sin email, conflicto de agenda y reejecución con event ID; mismo resultado funcional que la referencia |
| Telegram | Texto, voz, foto, PDF, enlaces, cola catchup, reinicio/reintento y reubicación; acceso limitado a chats configurados |
| Contexto y hooks | Arranque dentro/fuera del vault, manual grande, reglas locales, selección multilingüe, compactación y escritura; contrastar eventos y permisos reales de Codex |
| Representación y calidad | Canvas abierto en Obsidian, exportaciones consumibles, health/lint coherentes, eval reproducible y lectura de índices sin módulos faltantes |

## Decisiones necesarias para pasar a diseño

1. **Superficies y distribución:** Codex desktop, CLI y/o IDE; instalación local o
   aspiración al directorio público. La decisión afecta hooks, scheduling y acceso
   al filesystem, y no debe resolver las brechas eliminando funcionalidades.
2. **Experiencia con Obsidian:** preservar formatos, canvas, bases y plugin lint;
   acordar qué operaciones deben poder usarse sin abrir la aplicación.
3. **Integraciones externas:** conservar las capacidades de calendario, Telegram,
   X, audio/video y corpus propio; decidir continuidad o sustitución de proveedores
   con evidencia de equivalencia.
4. **Autorización y automatización:** trasladar la distinción original entre acciones
   automáticas, ofertas opcionales, dry-run y confirmaciones. Cualquier cambio de
   política se presenta expresamente; este inventario no impone confirmación nueva
   para cada escritura.
5. **Unidad de instalación y extensión:** acordar dónde viven reglas del vault,
   preferencias y comandos del usuario, y cómo sobreviven a una actualización.

Recomendación: usar este inventario para un diseño orientado a Codex que conserve
los resultados de usuario y reutilice las piezas verificables. Antes de cerrar ese
diseño, resolver B-001–B-003 y las decisiones anteriores en términos de alcance y
contratos. Elegir Go o Python, agrupar skills, diseñar persistencia o priorizar
releases permanece fuera de esta entrega de análisis.

## Cobertura reproducible

El registro de fuentes contiene 366 paths únicos: 47 comandos, 80 archivos de scripts
(40 de investigación y 40 restantes), 22 de integraciones, 12 hooks, 10 de adaptadores,
14 referencias, 82 tests, 49 documentos generados, 9 ejemplos, 7 medios, 8 de GitHub,
2 manifiestos Claude, 1 include del sitio y 23 archivos raíz.

La clasificación de tests, ejemplos, sitio y metadatos establece su presencia y
propósito; no afirma revisión exhaustiva de cada línea ni que cada recurso sea una
capacidad instalada. La fuente en `.repos` sigue sin cambios. La constitución
aprobada permanece intacta; no hay otros documentos de diseño, roadmap, spec o plan
del proyecto que reconciliar en el estado actual del checkout.
