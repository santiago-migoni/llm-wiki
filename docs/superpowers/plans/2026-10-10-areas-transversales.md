---
title: Plan de implementación de áreas y organización transversal
status: approved
version: v001
responsible: Usuario promotor del proyecto
created_at: "2026-10-10"
updated_at: "2026-10-10"
reviewed_at: "2026-10-10"
approval_reference: AA-002
---

# Plan de implementación de áreas y organización transversal

Entrega del producto: **v0.2.0** (runtime/manifiestos `0.2.0`);
[changelog](../../../CHANGELOG.md). La revisión documental v001 y AA-002 conservan su alcance.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox syntax for tracking. Método aprobado: ejecución directa por el agente
> principal, con revisión independiente final (AA-002).

**Goal:** integrar áreas opcionales, relaciones múltiples y destinos de OKRs y
entidades sin duplicar conocimiento ni alterar los workflows legítimos previos.

**Architecture:** Codex clasifica y conduce workflows; Go resuelve tipos/rutas,
enumera áreas con cobertura y aplica las mutaciones recuperables existentes.
CLI/MCP comparten el servicio. Enlaces canónicos en el cuerpo, sin nuevo registro
de áreas ni migraciones automáticas.

**Tech Stack:** Go 1.27+, SDK MCP Go y YAML ya fijados por go.mod; Markdown y
fixtures sintéticos. Sin nuevas dependencias de instalación o proveedores.

**Spec:** [especificación v001](../specs/2026-10-10-areas-transversales-design.md),
base exacta `e5121812e74fa53eeaf37e03c9496147cd68ffde:docs/superpowers/specs/2026-10-10-areas-transversales-design.md`.
Estado approved mediante AA-002. AA-001 conserva su alcance arquitectónico.
Base de aprobación: `3feb3410e72ee5795dae24b9eef3515106432210:docs/superpowers/ARCHITECTURE.md`.
No hay release numerada asignada. Spec/plan previos y AP-001 quedan intactos.

## Puerta de ejecución y restricciones globales

- [x] Aprobación explícita registrada mediante AA-002 antes de T1; método directo.
- Go 1.27 o posterior es requisito de desarrollo; no es dependencia de instalación.
- C-001 mantiene cuotas: 8 MiB por archivo, 64 MiB por scan y 10.000 entradas.
- Vault explícito; rutas confinadas, estado técnico protegido, raw existente
  inmutable, hashes y operaciones recuperables existentes. Ninguna fuente autoriza.
- No atomicidad global prometida; manual mediante wiki_init; informar pasos
  aplicados y pendientes entre llamadas. No añadir otro mecanismo de escritura.
- Áreas opcionales; sin categorías de vida obligatorias ni columnas Kanban nuevas.
- Sin mover notas, migrar metadata humana, ampliar parser a frontmatter o crear
  servicios/schedulers. No añadir dependencias ni modificar F/I/T de referencia.
- Un escritor por archivo; preservar cambios ajenos. El método elegido no cambia
  los contratos ni elimina las revisiones. Subagentes, si se eligen, no redelegan.

## Review Focus

1. Fila de mapa inválida que antes se ignoraba: diagnóstico explícito sin fallback;
   un mapa válido para otro tipo debe seguir funcionando. Tests en T1.
2. Alias, título Unicode o metadata malformada ocultan candidatos: conservar rutas
   distintas y no declarar ausencia; caso válido escalar/lista. Tests en T2.
3. Symlink/directorio excluido, cuota agotada o cancelación: preservar resultado
   parcial/errores y no mutar; corpus completo vacío legítimo. Tests en T2/T3.
4. Tarjeta breve, columnas Kanban, raw o campos humanos heredados: conservar bytes
   y formato; enlaces del cuerpo sí deben resolver. Tests en T5/T6.
5. Segunda vinculación, edición humana concurrente y journal pendiente: no duplicar
   ni anunciar éxito global; preservar lectura y recuperación válidas. Tests en T6.

## Mapa de archivos e interfaces

T1 posee layout/bootstrap y sus tests; T2 añade lector en vault; T3 posee servicio
y pruebas de transporte; T4/T5 poseen referencias/skills; T6 añade pruebas de
integración sin reescribir código ajeno; T7 posee paquete/documentación de entrega.
La aprobación de este plan acepta las interfaces internas propuestas siguientes:

- `Layout.ResolveFolder(kind string) (string, error)`, método puro en layout.go.
  Campo adicional `FolderErrors map[string]string` con JSON `folder_errors,omitempty`;
  conservar la representación existente de Style/Folders. Clone copia ambos mapas.
- `parseFolderMapDetailed(manual, style string) (map[string]string, map[string]string)`:
  destinos válidos y errores por clave. Wrapper parseFolderMap existente conserva
  su firma para consumidores legítimos; detección/selección usan el resultado detallado.
- `AreaEntry { Path, Name, Hash string; Aliases []string }` y
  `AreaDiagnostic { Path, Code, Message string }`, tags JSON de C-001.
- `AreaListResult { Folder string; Areas []AreaEntry; Coverage Coverage;
  Diagnostics []AreaDiagnostic; Complete bool }`, tags JSON de C-001.
- `(*Vault).ListAreas(ctx context.Context) (AreaListResult, error)` en areas.go;
  `listAreasWithLimits(ctx context.Context, maxBytes, maxEntries int)` privado para
  pruebas con presupuestos pequeños, usando scanWithLimits existente.
- Service.Definitions añade lector sin campos aparte de vault_id;
  Runtime.Call devuelve AreaListResult y error. No se añade tipo de transporte propio.

## T1 — Resolver rutas de áreas, OKRs y entidades

**Archivos:** modificar `internal/vault/layout.go`, `bootstrap.go`, `notes.go`;
tests `vault_test.go`, `bootstrap_test.go`, `notes_test.go`.
**Consume:** C-002 y ADR-010. **Produce:** ResolveFolder y FolderErrors del mapa anterior.

- [x] Añadir `TestLayoutAreaEntityOKRPrecedence`: tabla wiki/Obsidian con claves de
  C-002; comprobar subtipo explícito > entity explícito > default; company/companies
  resuelven organization; sin entity nuevo, person conserva comportamiento anterior.

  Aserción del caso wiki sin mapa explícito, con vault ya abierto por el test:

  ```go
  folder, err := v.Layout().ResolveFolder("organization")
  if err != nil || folder != "wiki/entities" {
      t.Fatalf("folder=%q err=%v", folder, err)
  }
  ```
- [x] Añadir `TestLayoutInvalidExplicitFolderDoesNotFallback`: ruta protegida,
  absoluta, traversal, layout ambiguo y override válido posterior; comprobar error
  del tipo afectado y resolución legítima del otro tipo. Clone no comparte mapas.
- [x] Ejecutar `go test ./internal/vault -run 'TestLayout(AreaEntityOKRPrecedence|InvalidExplicitFolderDoesNotFallback)'`;
  observar fallos por las claves/diagnósticos ausentes antes de implementar.
- [x] Extender normalización/defaults y parser detallado. Combinar _CLAUDE primero,
  manual después por clave; una fila efectiva inválida conserva error. Aplicar la
  precedencia de C-002 después de resolver las entradas explícitas por origen.
- [x] Usar ResolveFolder para daily en PlanCapture/PlanSave, que hoy leen Folders
  directamente. Conservar uso de rutas explícitas y controles de mutación existentes.
- [x] Añadir filas reconocidas en bootstrap preservando presets; no crear notas
  de área. Probar `TestBootstrapNewKindsPreserveExisting` y `TestCaptureInvalidDailyMap`:
  no reemplazo de archivos, no carpeta creada al resolver, captura válida sin áreas.
- [x] Ejecutar `go test ./internal/vault`, gofmt de archivos modificados y diff --check;
  revisar y registrar commit solo de los archivos asignados. Resultado esperado:
  tests nuevos y previos pasan; fallos/bloqueos se registran como tales.

## T2 — Inventario de áreas con cobertura

**Archivos:** crear `internal/vault/areas.go`, `areas_test.go`.
**Consume:** T1 ResolveFolder y ScanResult. **Produce:** tipos y ListAreas definidos arriba.

- [x] Añadir `TestListAreasStableIdentityAndNames`: dos áreas Unicode con aliases
  escalar/lista, título/H1/basename, homónimos y ruta anidada; comprobar orden,
  hashes completos, arrays no nulos y ambiguous_name sin colapsar identidades.
- [x] Añadir `TestListAreasIncompleteAndLegacy`: carpeta ausente válida, nota sin
  tipo en destino, YAML roto, título/alias no string, área fuera del mapa, mapa
  inválido, symlink excluido y lectura fallida. Afirmar códigos de C-001, complete
  y error; el caso vacío válido es []/complete=true y no crea archivos.

  Aserción del caso vacío válido:

  ```go
  result, err := v.ListAreas(context.Background())
  if err != nil || !result.Complete || result.Areas == nil || len(result.Areas) != 0 {
      t.Fatalf("result=%+v err=%v", result, err)
  }
  ```
- [x] Añadir `TestListAreasBudgetsAndCancellation`: presupuestos pequeños mediante
  helper privado y contexto cancelado; complete=false, errores conservados y
  hashes parciales nunca presentados como completos. Comparar archivos antes/después.
- [x] Ejecutar `go test ./internal/vault -run TestListAreas`; confirmar que falla
  antes de implementar por API ausente, no por fixtures incorrectos.
- [x] Implementar ListAreas con un único scan, Parse existente y reglas C-001;
  chequear cancelación en el análisis posterior al scan. Sin nuevo índice o caché.
- [x] Ejecutar `go test ./internal/vault`, gofmt y diff --check; revisar y commit
  de areas.go/areas_test.go. Solo procede T3 cuando contrato y tests estén conformes.

## T3 — Exponer el lector en servicio, CLI y MCP

**Archivos:** modificar `internal/service/service.go`, `service_test.go`,
`internal/mcp/server_test.go` y `cmd/llm-wiki/main_test.go`.
**Consume:** T2 ListAreas. **Produce:** contrato público wiki_list_areas de C-001.

- [x] Añadir `TestListAreasContractAndPartialError` en service: vault-id ausente/
  desconocido, entrada legítima, layout ambiguo, mapa inválido y scan parcial.
  Comparar result/error y asegurar que no se aplica ninguna mutación.
- [x] Añadir `TestMCPListAreasMatchesService` y `TestCLIListAreasMatchesService`:
  mismo fixture/resultado, MCP IsError con parcial y CLI error; serve no imprime
  cuerpos de notas/diagnósticos ajenos al protocolo en stdout.
- [x] Ejecutar `go test ./internal/service ./internal/mcp ./cmd/llm-wiki -run ListAreas`;
  verificar fallos por herramienta ausente antes de registrar la definición.
- [x] Registrar definición y dispatch al lector antes del scan genérico del servicio
  para no duplicar el recorrido; transporte MCP actual consume Definitions sin
  otro adaptador. Conservar resultado parcial en el envoltorio result/error.
- [x] Ejecutar `go test ./internal/service ./internal/mcp ./cmd/llm-wiki`, gofmt,
  diff --check, revisión y commit de archivos asignados.

## T4 — Skills de áreas y OKR con reglas compartidas

**Archivos:** crear `plugins/llm-wiki/skills/wiki-obsidian-area/SKILL.md`,
`wiki-obsidian-okr/SKILL.md`, `plugins/llm-wiki/references/areas.md`;
modificar `folder-map.md`, `ai-first-rules.md`, `write-rules.md` y router llm-wiki;
extender `internal/catalog/catalog_test.go` y registro de capacidades de extensión.
**Consume:** T1 rutas/T3 lector. **Produce:** flujos US-001/006 y reglas para T5.

- [x] Definir escenario fallido de aceptación antes de añadir skills: solicitar
  área arbitraria y OKR incompleto; falta del lector/skill se reporta, no se simula.
- [x] Escribir flujos de crear/consultar/vincular área y crear/consultar/actualizar
  OKR; lectura completa, reutilización, aprobación de nuevas áreas y pendientes
  explícitos según spec. Consulta sola no escribe; manual usa wiki_init.
- [x] Registrar extensiones sin reutilizar F/I/T originales; si hace falta una
  sección extensions, usar IDs locales EXT-AREA/EXT-OKR y estados acordes a evidencia,
  sin cambiar el inventario original. Documentar la ubicación en el catálogo.
- [x] Añadir `TestCatalogDiscoversAreaAndOKRExtensions`: skill parseable y
  descubrible, personalización del usuario preservada. Este test acredita catálogo,
  no clasificación ni resultados del agente.
- [x] Ejecutar `go test ./internal/catalog ./internal/packaging`; validar YAML,
  JSON/enlaces y revisar políticas antes del commit de archivos asignados.

## T5 — Integración en workflows pertinentes

**Archivos:** modificar skills project/projects/person/capture/save/task/calendar/
daily/log/review/synthesize/distill; referencia areas.md de T4; crear
`tests/acceptance/areas-scenarios.md` y matriz `areas-cases.json`.
**Consume:** T4 reglas comunes. **Produce:** US-004 sin ampliar permisos de cada workflow.

- [x] Escribir casos de matriz antes de editar skills: por cada fila de arquitectura,
  modo con clasificación autorizada y modo previo sin áreas; expected paths,
  relaciones, escrituras permitidas y fuentes. Ninguna skill nueva sustituye otras.
- [x] Integrar consulta/vinculación opcional, reutilización y manejo de ausencia/
  ambigüedad. No replicar reglas completas: enlazar referencia canónica areas.md.
- [x] Añadir casos `ShortTaskKeepsCard`, `KanbanKeepsColumns`, `RawBytesUnchanged`,
  `HumanMetadataPreserved`, `UnknownOKRValuesPending` y `GeneralIntentDoesNotWrite`;
  verificar manualmente cambios de políticas contra los originales de la base.
- [x] Añadir `RepeatedLinkRequestWithCurrentHash`: el usuario vuelve a solicitar
  vincular la misma nota a las mismas áreas; el agente relee la nota/hash vigente
  y determina que los enlaces ya existen. Afirmar nota única, misma ruta y un solo
  enlace por área; si hay otra actualización autorizada, debe usar un nuevo ID y
  hash vigente sin duplicar vínculos. No aceptar replay del ID anterior como
  sustituto de esta prueba del workflow. Ejecutar también el caso legítimo de
  añadir una segunda área previamente ausente, conservando la primera.
- [x] Validar sintaxis JSON/YAML y enlaces; ejecutar `go test ./internal/catalog`.
  Ejecutar casos con agente/host si está disponible, registrando por caso entrada,
  operaciones y diffs observados. Si no lo está, marcar aceptación pendiente:
  la lectura de skills no cuenta como ejecución aprobada.
- [x] Revisar y commit del diff asignado y matriz; no modificar raw reales ni
  mantener categorías del usuario en fixtures versionados.

## T6 — Regresiones de relaciones, continuidad y recuperación

**Archivos:** crear `tests/fixtures/areas-wiki/`, `areas-obsidian/`,
`tests/acceptance/areas_test.go`, `internal/service/areas_integration_test.go`;
extender `internal/graph/graph_test.go`, `internal/vault/process_recovery_test.go`.
**Consume:** T1–T5. **Produce:** evidencia técnica y matriz de aceptación distinguibles.

- [x] Añadir `TestAreaRelationsSingleNoteAndRepeatedLink`: una nota con dos áreas,
  grafo/backlinks desde ambas, ruta sin cambio y una nota canónica; repetir el lote
  con mismo operation_id/payload es idempotente. Otro ID con hash stale debe fallar.
- [x] Añadir `TestAreaLinksInBodyNotFrontmatter`: cuerpo resuelto, YAML preservado
  y no interpretado; múltiples aliases Unicode conservan reglas existentes.

  Aserción con scan de fixture que contiene un proyecto único enlazado a dos áreas:

  ```go
  g := graph.Build(scan)
  for _, target := range []string{"wiki/areas/Exploración.md", "wiki/areas/Comunidad.md"} {
      refs := g.Backlinks(target)
      if len(refs) != 1 || refs[0].Source != "wiki/projects/Nexo.md" {
          t.Fatalf("target=%q backlinks=%+v", target, refs)
      }
  }
  ```
- [x] Añadir `TestAreaSecondProcessAndRecovery`: reiniciar proceso con mismo estado,
  leer áreas y backlinks; inducir fallo parcial, edición humana y journal pendiente;
  reanudar/rollback solo donde hashes permiten. No afirmar aislamiento global.
- [x] Correr primero los tests nuevos para observar fallos antes de agregar soporte
  o fixtures definitivos; si las primitivas ya satisfacen un caso, conservarlo como
  prueba de compatibilidad sin fabricar un fallo. No ampliar el grafo a YAML.
- [x] Ejecutar `go test ./internal/vault ./internal/graph ./internal/service ./tests/acceptance`;
  cerrar casos de formato/coverage de Review Focus. Tests con llamadas prediseñadas
  acreditan runtime, no decisiones semánticas de un agente.
- [x] Ejecutar escenario de dos chats Codex sobre fixtures sintéticos; registrar
  host/plugin, entrada, operaciones/diffs y fuentes recuperadas. Ausencia del host
  deja este criterio pendiente sin sustituirlo por dos procesos.
- [x] Ejecutar `go test -race ./...`, `go vet ./...`, gofmt y `git diff --check`;
  revisión final del diff y commit de archivos propios. No repetir suite sin cambios.

## T7 — Paquete y documentación de entrega

**Archivos:** README, capabilities y referencias derivados; documentos técnicos
afectados, evidencia nueva `docs/superpowers/evidence/2026-10-10-areas.md`;
artefactos generados por cmd/package. Versiones de internal/version y manifiestos
solo después de asignar release. **Consume:** T6 y versión elegida antes de distribuir.

- [x] Registrar resultados, omisiones y límites por componente/workflow/host;
  mantener EXI-002/003 como resultados de usuario pendientes cuando corresponda.
- [x] Asignar versión de entrega con el usuario antes de editar versiones; no
  inventarla durante ejecución. Conservar licencias/atribuciones y no crear changelog
  por este plan. No instalar, publicar ni hacer push por asociación.
- [x] Construir en salida nueva:

  ```sh
  package_parent=$(mktemp -d)
  go run ./cmd/package --output "$package_parent/marketplace"
  LLM_WIKI_PACKAGE="$package_parent/marketplace/plugins/llm-wiki" go test ./tests/acceptance -run TestPackagedBinaryWithoutDevelopmentToolchain -count=1
  ```

- [x] Extender aceptación empaquetada para descubrir wiki_list_areas/skills nuevas
  y ejecutar lectura tras reinicio, conservando las pruebas previas. Comprobar
  manifests, identidad/versiones, firma, arquitecturas, checksums y source fingerprint
  contra las fuentes reales; build con cambios locales se registra como tal.
- [x] Copiar únicamente artefactos comprobados al paquete versionado y registrar
  procedencia. Compilar Intel no acredita ejecución Intel. Confirmar diff y commit
  de artefactos/evidencia propios; no marcar completa la aceptación de host pendiente.

## Trazabilidad y revisión propia

| Spec | Tareas |
| --- | --- |
| US-001 | T1, T2, T4, T5, T6 |
| US-002, C-001 | T2, T3, T6, T7 |
| US-003 | T4, T5, T6 |
| US-004 | T4, T5, T6 |
| US-005, C-002 | T1, T3, T4, T6 |
| US-006 | T1, T4, T5, T6 |
| C-003 y distribución | T1–T7 |

Orden T1 → T2 → T3 → T4 → T5 → T6 → T7; no escribir archivos compartidos en
paralelo. Cada tarea termina con revisión/commit limitado. La preparación de este
plan verificó cobertura de historias, consistencia de firmas, comandos y foco de
revisión; no ejecutó sus tests. Los checkboxes se actualizan con la evidencia de ejecución; aceptación del host queda pendiente.

## Handoff y pendientes

AA-002 aprueba las bases y ejecución directa. Versión elegida: 0.2.0.
Resultados y decisiones: [evidencia de áreas](../evidence/2026-10-10-areas.md).
T1–T7 completadas técnicamente. Revisión independiente final: tres Important
corregidos en una pasada, sin Critical ni Minor diferidos. Suite con carreras,
vet y paquete 0.2.0 recompilado/verificado pasan. La aceptación del host
instalado/dos chats queda pendiente y conserva su casilla abierta.

2026-10-10, v001: borrador solicitado sobre AA-001 y spec v001 draft fijada arriba.
No ejecutado ni aprobado; no cambia el plan anterior del plugin.

## AA-002 — Aprobación y ejecución directa

- Fecha: 2026-10-10. Responsable: usuario promotor del proyecto.
- Fuente: mensaje humano en esta conversación: «Apruebo la especificación y el plan. Avanza con la implementación directa y verifica los cambios según el plan.»
- Spec presentada: `e5121812e74fa53eeaf37e03c9496147cd68ffde:docs/superpowers/specs/2026-10-10-areas-transversales-design.md`.
- Plan presentado: `bc9f6fa2c7d333c3a1e2d42fd840b5c1c8e7e527:docs/superpowers/plans/2026-10-10-areas-transversales.md`.
- Alcance: especificación v001, tareas T1–T7 y ejecución directa con verificación; no autoriza publicación, instalación ni push. Conserva AA-001 y AP-001.

## Cierre de aceptación Desktop — 2026-10-10

Los estados pendientes de host anteriores son históricos. Tras corregir los
manifiestos y reiniciar Desktop, ambos chats existentes reciben 30 herramientas
MCP. Se ejecutaron 24 casos pareados de los 12 workflows y 14 escenarios de
guardas con los límites de interacción detallados en la matriz. La segunda pasada
del chat lector recuperó los mismos hashes, valores pendientes de OKR y 12
backlinks por área, con cobertura completa de 20 archivos.

Evidencia: [informe Desktop](../evidence/2026-10-10-desktop-areas.md) y
[observaciones](../../../tests/acceptance/desktop-areas-observed.json).
El renderizado Obsidian, calendario remoto, ejecución Intel y resultados de
usuario EXI-002/003 no quedan acreditados por esta aceptación sintética.
