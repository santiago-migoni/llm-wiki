---
title: Especificación de áreas y organización transversal en llm-wiki
status: approved
version: v001
responsible: Usuario promotor del proyecto
created_at: "2026-10-10"
updated_at: "2026-10-10"
reviewed_at: "2026-10-10"
approval_reference: AA-002
---

# Especificación de áreas y organización transversal

Entrega del producto: **v0.2.0** (runtime/manifiestos `0.2.0`);
[changelog](../../../CHANGELOG.md). La revisión documental v001 y AA-002 conservan su alcance.

## Contexto y resultado

Permitir crear y recuperar áreas elegidas por el usuario y relacionar una misma
unidad de conocimiento con varias de ellas, conservando su contenido y ubicación.
Integrar esta organización opcional en los workflows existentes y resolver rutas
de OKRs y entidades de forma consistente. La entrega asignada es `v0.2.0`; esta
es la fuente canónica del incremento, sin sustituir la spec anterior del plugin.

Fuentes exactas:

- Propuesta presentada: `1b4b6e00bfc4ea939140007196cfe820b04e8d38:docs/superpowers/ARCHITECTURE.md`
  v008, ADR-009/010 e integración transversal; estructura v004 en el mismo commit.
- Registro aprobado: `3feb3410e72ee5795dae24b9eef3515106432210:docs/superpowers/ARCHITECTURE.md`,
  AA-001. El usuario aprueba la propuesta y solicita spec/plan; no aprueba por
  asociación los detalles nuevos de esta spec ni su ejecución.
- Constitución v003 approved: `3feb3410e72ee5795dae24b9eef3515106432210:docs/superpowers/CONSTITUTION.md`,
  PRI-003–005, OBJ-002–003, EXI-002–003. Persistencia y recuperación existentes en
  `docs/superpowers/INFRASTRUCTURE.md`, v002, de la misma base.
- Inspección estática del runtime y consumidores en esa base; no se ejecutaron
  workflows, pruebas Go, instalación ni pruebas con usuarios al preparar esta spec.

## Alcance y exclusiones

Incluye tipo/ruta de áreas, lector `wiki_list_areas`, skills de áreas y OKR,
resolución de entidades, integración opcional con workflows, regresiones y
empaquetado verificable. No incorpora proveedores ni dependencias de instalación.
No incluye búsqueda federada/semántica, clasificación en Go, scheduler, jerarquía
de áreas, renombrado/fusión/borrado de áreas ni migraciones masivas. No convierte
metadata humana, mueve notas ni duplica carpetas por área. No modifica AP-001 ni
las bases de spec/plan previos.

## Historias y aceptación

### US-001 — Crear y reutilizar áreas propias

Como usuario, quiero definir ámbitos con mis propios nombres y propósitos para
organizar conocimiento sin adoptar categorías obligatorias.

- Solicitar crear un área genera una nota `type: area`, propósito conocido,
  procedencia y contexto AI-first; información no dada queda pendiente.
- Antes de crear, el agente consulta áreas con cobertura suficiente y lee las
  coincidencias; reutiliza la existente sin sustituir contenido humano.
- Homónimos o similitud dudosa requieren aclaración; nombres y aliases son datos,
  no autorización. No existe una lista fija ni un número mínimo de áreas.
- La ruta utiliza el mapa; crear solo genera padres necesarios y navegación
  autorizada. El nombre visible no permite escape de raíz; un nombre que requiera
  ajuste como archivo se presenta con su alternativa segura, sin sobrescribir.
- Inventario incompleto impide concluir ausencia y crear para deduplicar. Un
  vault sin áreas y con inventario completo devuelve colección vacía sin mutar.

### US-002 — Consultar áreas con evidencia de cobertura

Como usuario/agente, quiero conocer las áreas existentes para recuperar contexto
y elegir vínculos sin depender de coincidencias léxicas parciales.

- `wiki_list_areas` exige `vault_id` explícito, no crea archivos ni acepta una raíz
  alternativa. Desconocido/no seleccionado conserva los errores del servicio.
- Devuelve candidatos ordenados por ruta, nombre, aliases y SHA-256 del archivo
  completo, carpeta resuelta, cobertura y diagnósticos del contrato C-001.
- Una nota `type: area` fuera del destino configurado se diagnostica y no se migra.
  Metadata inválida, lecturas fallidas, cuotas o cancelación nunca aparentan
  inventario completo. Las notas legibles se pueden mostrar como resultado parcial.
- Nombres/aliases iguales no colapsan dos rutas distintas; el agente muestra la
  ambigüedad. La herramienta no decide clasificación semántica.
- CLI/MCP comparten payload y error del servicio; stdout de serve contiene solo
  protocolo. MCP marca el error y conserva el resultado parcial cuando corresponda.

### US-003 — Una nota en varias áreas

Como usuario, quiero consultar una nota desde diferentes áreas sin mantener copias.

- Enlaces canónicos en el cuerpo, dentro de la sección de áreas existente o una
  sección apropiada adoptada por el manual. No exige área principal ni YAML `area`.
- Una nota con dos áreas se recupera mediante backlinks de ambas, conservando la
  misma ruta. Repetir vinculación no duplica la nota ni el enlace.
- El índice del área, si existe y su actualización está autorizada, es navegación
  derivada; no provoca reconstrucciones o reescrituras generales.
- Hash obsoleto rechaza actualización; fallo parcial se comunica y se recupera
  con las operaciones existentes sin pisar ediciones posteriores.
- Segunda ejecución del runtime y segunda sesión Codex recuperan las mismas
  relaciones; se registran como evidencias distintas, no equivalentes implícitos.

### US-004 — Organización opcional en workflows

Como usuario, quiero usar áreas durante las acciones que solicito y conservar
mis workflows anteriores cuando no adopte esa organización.

- Activación por petición o convención adoptada dentro del alcance solicitado;
  una mención general a proyectos o un texto externo no inicia escrituras.
- Project/projects, person, capture/save/task, calendar-meeting, daily/log/review,
  synthesize/distill y OKR respetan la matriz de arquitectura v008.
- Reutilizan áreas existentes; una nueva solo se crea por solicitud o aceptación.
  No propagan a más destinos que los prescritos por el workflow y autorizados.
- Una tarea breve sigue siendo tarjeta; sus enlaces van dentro de la tarjeta o
  nota ya relacionada. Kanban conserva columnas; raw conserva bytes.
- Daily mantiene diario común, log mantiene proyecto/fecha y review mantiene
  periodo/cobertura. No se genera una versión de cada uno por área.
- Metadata heredada se preserva y se informa que el grafo no interpreta sus
  vínculos. Sin áreas, los resultados legítimos de los workflows previos se conservan.

### US-005 — Destinos consistentes para áreas, OKRs y entidades

Como usuario, quiero adaptar las carpetas sin perder las convenciones anteriores.

- Rutas de C-002 y precedencia: mapa válido explícito del subtipo, mapa válido
  explícito de `entity` para subtipos sin destino propio, default del subtipo.
- Un mapa inválido explícito se diagnostica y no dispara fallback silencioso para
  la escritura que depende de esa ruta; layout ambiguo bloquea resolución.
- Alias `company`/`companies` resuelven `organization`; no cambia el frontmatter
  humano ni el contrato existente de `person` cuando no hay mapa genérico nuevo.
- Resolver rutas no crea directorios. Una ficha de entidad puede pertenecer a
  varias áreas sin otra ficha por área. No obliga a subdividir `wiki/entities/`.
- Bootstrap conserva archivos/presets y publica las claves reconocidas del mapa;
  no crea notas de área ni columnas/categorías de vida predeterminadas.

### US-006 — Registrar y consultar OKRs sustentados

Como usuario, quiero conservar objetivos/resultados y sus relaciones con áreas
sin que el agente invente compromisos.

- Workflow explícito de OKR consulta/reutiliza notas antes de crear. Usa ruta
  resuelta y las convenciones del vault; consulta sola no escribe.
- Objetivo, resultados, métricas, línea base, meta, responsables y fechas se
  registran solo cuando constan. Desconocidos quedan pendientes, nunca estimados.
- Se permiten datos pendientes sin afirmar que el OKR ya es medible o completo.
  Tablero/índice se actualizan solo si forman parte de la convención adoptada.
- Reutiliza lector y enlaces de áreas y escrituras generales; no crea objetivos,
  nuevas áreas ni un calendario por inferencia.

## Contratos y restricciones compartidas

Los siguientes detalles concretan la propuesta aprobada y se presentan para
revisión de esta v001; AA-001 no aprobó sus esquemas exactos.

### C-001 — Consulta de áreas

Entrada JSON: `{ "vault_id": "identidad-seleccionada" }`, sin campos adicionales.
Salida del servicio dentro del envoltorio actual `result`:

| Campo | Tipo y significado |
| --- | --- |
| `folder` | string: ruta relativa resuelta, vacía cuando no se pudo resolver |
| `areas` | array de `{path, name, aliases, hash}`; path/name/hash strings y aliases array de strings |
| `coverage` | `vault.Coverage` existente: scanned, excluded, errors, truncated |
| `diagnostics` | array de `{path, code, message}` con strings; sin volcar cuerpos de notas |
| `complete` | boolean: validez del inventario para determinar ausencia, conforme a las reglas siguientes |

Lectura del corpus mediante `Vault.Scan(ctx)`, una sola vez, reutilizando cuotas
compartidas actuales: 8 MiB por archivo, 64 MiB por scan y 10.000 entradas, con
cancelación. No caché persistente, paginación nueva ni índice paralelo de áreas.
Orden estable por path, arrays vacíos como `[]` y hashes completos existentes.

Parsear frontmatter con `noteformat.Parse`; tipo `area` identifica el candidato.
Nombre: title string no vacío, después primer H1 no vacío, después basename sin
extensión. Aliases: ausente, string o lista de strings, sin evaluar su contenido.
Tipo/title/aliases con forma inválida en la carpeta de áreas producen diagnóstico,
no un candidato válido. Frontmatter malformado en cualquier nota del corpus se
diagnostica conservadoramente porque podría ocultar un área.

`complete=false` si hay error de scan, truncación, errores de cobertura, exclusión
no técnica sin cobertura alternativa, error de mapa, candidato malformado o área
fuera del destino. Excluir `.git` y estado técnico conforme al scanner es esperado
y no invalida por sí solo el inventario; otras exclusiones sí deben explicarse.
Codificar diagnósticos como `invalid_folder`, `invalid_metadata`, `misplaced_area`
o `incomplete_scan`. Coincidencia de nombre/alias usa strings.EqualFold; variantes
por diacríticos o similitud son candidatas para aclaración, no identidades iguales.
Dos rutas válidas homónimas permanecen en la lista, con `ambiguous_name`; complete
puede ser true porque cobertura y unicidad son dimensiones distintas.

Fallos de scan/mapa/metadata devuelven resultado parcial y error no nulo del
servicio; MCP usa IsError y CLI falla sin anunciar éxito. Ambigüedad de nombres
sin fallo de inventario devuelve diagnóstico, no error global. Herramienta ausente
en host se informa sin sustituirla por una escritura o inventario ficticio.

### C-002 — Resolución de destinos

| Clave normalizada | Wiki | Obsidian |
| --- | --- | --- |
| area | wiki/areas | Areas |
| okr | wiki/okrs | OKRs |
| entity | wiki/entities | Entities |
| person | wiki/entities | People/people según selección actual |
| organization (alias company/companies) | wiki/entities | Companies |
| tool | wiki/entities | Tools |

La resolución expone diagnósticos de rutas reconocidas inválidas en el contexto,
sin escribir ni crear carpetas. Las operaciones con ruta explícita conservan
controles generales; no se presume que el runtime pueda verificar autorización
semántica. El agente debe resolver errores antes de ejecutar el workflow afectado.
No se altera la ruta de diarias por año/mes ni las demás claves existentes.

### C-003 — Seguridad, fuentes y evidencia

- Go 1.27 o posterior es requisito de desarrollo; no es dependencia de instalación.
- Vault explícito; rutas confinadas, estado técnico protegido, raw existente
  inmutable, hashes y operaciones recuperables existentes. Ninguna fuente autoriza.
- No atomicidad global prometida. Manual mediante wiki_init; las demás mutaciones
  usan herramientas generales. Informar pasos aplicados y pendientes entre llamadas.
- Conservar identidad de F/I/T de referencia; las skills nuevas son extensiones
  trazables, sin atribuirles IDs originales ni paridad por presencia de archivos.
- Ensayos con fixtures sintéticos; paquete nuevo, manifiestos, firma, arquitecturas,
  checksums y procedencia comprobados antes de anunciar distribución.

## Pendientes y autoridad

- C-001/C-002 y criterios v001 aprobados mediante AA-002; ejecución directa.
- Versión de entrega 0.2.0 elegida expresamente por el usuario. La aprobación
  autoriza implementación y verificación, sin instalación/publicación implícitas.
- La prueba real en dos chats Codex requiere capacidades del host y un entorno
  sintético. Si no está disponible, queda pendiente y se conserva evidencia del
  runtime por separado; no sustituirla por una prueba de catálogo o dos procesos.

## Historial

2026-10-10, v001: primer borrador del incremento solicitado sobre AA-001.
Estado draft; ningún criterio está aprobado o verificado por la creación del archivo.

## AA-002 — Aprobación y ejecución directa

- Fecha: 2026-10-10. Responsable: usuario promotor del proyecto.
- Fuente: mensaje humano en esta conversación: «Apruebo la especificación y el plan. Avanza con la implementación directa y verifica los cambios según el plan.»
- Spec presentada: `e5121812e74fa53eeaf37e03c9496147cd68ffde:docs/superpowers/specs/2026-10-10-areas-transversales-design.md`.
- Plan presentado: `bc9f6fa2c7d333c3a1e2d42fd840b5c1c8e7e527:docs/superpowers/plans/2026-10-10-areas-transversales.md`.
- Alcance: especificación v001, tareas T1–T7 y ejecución directa con verificación; no autoriza publicación, instalación ni push. Conserva AA-001 y AP-001.
