---
title: Plan de implementación de llm-wiki basado en especificación funcional
status: approved
version: v003
responsible: Usuario promotor del proyecto
created_at: "2026-10-08"
updated_at: "2026-10-08"
reviewed_at: null
approval_reference: "AP-001 — aprobación y ejecución"
---

# Plan de implementación del plugin llm-wiki para Codex

> **Ejecución actual:** directa, sin subagentes, por instrucción del usuario del 2026-10-08. Usar `superpowered:executing-plans`; verificar cada entrega y actualizar sus checkboxes después de leer el resultado. Las revisiones anteriores se conservan como evidencia.

**Goal:** construir y verificar un plugin local instalable en Codex Desktop macOS, conservando los resultados del inventario mediante Codex y un núcleo Go, con brechas explícitas.

**Architecture:** skills para razonamiento y workflows; MCP STDIO/CLI sobre servicios Go comunes para archivos, búsqueda y recuperación. Ningún cliente de IA, conector externo ni clave adicional. Marketplace Git con binario precompilado.

**Tech Stack:** Go, SDK oficial MCP Go (versión a fijar y probar en tarea 1), JSON, Markdown, binario universal macOS ensamblado con lipo de desarrollo.

**Spec:** [Especificación funcional v001](../specs/2026-10-08-codex-plugin-spec.md), base exacta `e6daa9743e15ef6bb332e8057100d32eafe86bb6:docs/superpowers/specs/2026-10-08-codex-plugin-spec.md`. Aprobada mediante AP-001. Las brechas O-001–009 siguen delimitando los criterios que requieren evidencia adicional.

**Diseño de apoyo:** arquitectura v005, estructura/infraestructura v001 en `80ae6c8b5849301acf6b5a46166b34cbc79a2aed`. La arquitectura decide organización técnica; la spec es la fuente de comportamiento y aceptación. Dirección aprobada: constitución v003/AC-019 en `f27672f7a63039805784efd8ccd5478b089dda2e`.

**Base previa:** `b7e81a038de4c189fb293e96cae92010ed79fbf4:docs/superpowers/plans/2026-10-08-codex-plugin.md`. Este v002 sustituye aquel borrador no aprobado, sin borrar su historia. No hay release numerada ni otro plan canónico.

**Estado:** en implementación, 2026-10-08. Spec y plan aprobados mediante AP-001; el usuario cambió la ejecución a trabajo directo sin subagentes. Se trabaja en la rama `codex/llm-wiki-plugin`. Un checkbox marcado acredita el paso concreto, no todos los criterios de la tarea.

## Avance visible

| Tarea | Implementación | Aceptación del producto |
|---|---|---|
| T1 — configuración/MCP inicial | Implementada y revisada; schema v2 para identidad por vault | Pendiente de interacción Desktop |
| T2 — lectura/layouts | Implementada y revisada | Contrato de lectura aceptado con fixtures |
| T3 — mutaciones/recuperación | Implementada y revisada | Integración MCP/CLI y recuperación tras salida de proceso verificadas; retención O-006 pendiente |
| T4 — grafo/edición | Implementada y revisada directamente | Canvas/scope aceptado; render y prueba integral de repetición pendientes |
| T5 — búsqueda/salud | Búsqueda léxica y salud implementadas; revisión directa verificada | Pendiente; O-003 semántica abierta |
| T6 — skills/continuidad | Assets, catálogo, contexto y operaciones MCP/CLI implementados; continuidad del núcleo probada | Pendiente |
| T7 — bootstrap/export | Bootstrap con plantillas de los 8 presets, init, migración y export implementados y probados | Contratos locales aceptados con fixtures |
| T8 — capacidades del host | Investigación general y síntesis 12/13 verificadas; resto de modalidades pendiente | Pendiente/brechas O-001–009 |
| T9 — marketplace | Paquete universal generado, validado y probado sin toolchain; instalación CLI realizada | Pendiente |
| T10 — instalación real | Instalación CLI y descubrimiento verificados; Desktop pendiente | Pendiente |

Evidencia: [registro de implementación](../evidence/2026-10-08-implementation.md), `tests/acceptance/results.json` y commits de la rama. Actualizar esta tabla junto con cada cierre verificable.

## Global Constraints

RC-001–008 de la spec son normativos para todas las tareas. No copiar otra versión
independiente de esos requisitos aquí. O-001–009 limitan el cierre de las tareas
afectadas; no se completan sustituyendo un criterio positivo por diagnóstico.
Las decisiones técnicas nuevas (CLI, SDK/versiones, binario universal, retención
30 días y estrategia sin dependencias locales de medios) son propuestas del
diseño. No ampliar AC-019 para revocar silenciosamente AC-010/011.

## Review Focus

1. Nombres Unicode, espacios, alias y duplicados en dos layouts: ningún enlace se reescribe por coincidencia ambigua (tareas 2, 4).
2. Cambio humano entre lectura/escritura y fallo tras primer archivo: sin sobrescribir contenido reciente ni anunciar lote completo (tarea 3).
3. Lectura parcial, permisos denegados, exclusiones y corpus grande: no afirmar ausencia/exhaustividad con truncamiento (tareas 2, 5).
4. Capability ausente, enlace privado o multimedia sin transcripción: no inventar contenido ni introducir proveedor (tareas 6, 8).
5. Proceso GUI sin variables de shell, actualización y segunda sesión: asociación de vault/personalizaciones sobreviven sin rutas privadas en paquete (tareas 1, 9, 10).

## Secuencia y criterio de cierre

T1 → T2 → T3; T4/T5 consumen T2/T3 y pueden desarrollarse en paralelo con ownership separado. T6 depende de T4/T5; T7 de T3/T4/T5; T8 integra T6/T7 y evidencia del host. T9 empaqueta; T10 verifica instalación real.

La ejecución actual es directa, sin subagentes. Se conserva la revisión de entregas y la verificación por tarea; las revisiones independientes ya realizadas permanecen registradas. No se termina la entrega declarando paridad si una modalidad no pasó su prueba. Las brechas de servicios pueden impedir el cierre completo sin impedir el núcleo local; deben escalarse con alternativa concreta, no borrarse del plan.

### T1 — Toolchain, configuración y proceso MCP

**Criterios propietarios:** US-001.AC-02. **Fuente:** SPEC v001; RC-001–008.

**Files:** `go.mod`, `go.sum`, `cmd/llm-wiki/main.go`, `internal/config/{config.go,config_test.go}`, `internal/mcp/{server.go,server_test.go}`.
**Interfaces:** `config.Open(stateRoot string) (*Registry,error)`; `(*Registry).Select(projectPath,vaultPath string) (VaultRef,error)`; `VaultRef{ID,Root,Layout,Timezone string}`; `mcp.Run(ctx context.Context,in io.Reader,out,diagnostics io.Writer,registry *config.Registry) error`.

- [x] Preparar toolchain Go de desarrollo en ubicación aislada; fijar versión compatible con SDK oficial, registrar checksums/versiones en go.mod/evidencia. No alterar PATH/config global ni instalar runtime para usuarios.
- [x] Escribir `TestSelectPersistsAcrossRestart`, `TestMissingVaultSelection`, `TestNoShellEnvironmentRequired` y `TestMCPInitializeAndList`: registro sobre raíz temporal, segunda instancia devuelve mismo ID, root inexistente falla, handshake valida protocolo.
- [x] Ejecutar `go test ./internal/config ./internal/mcp`, observar fallo antes de implementar.
- [x] Implementar registro atómico versionado; host inicia sin vault y permite selección explícita. `serve` reserva stdout; `doctor` no lee contenido personal.
- [x] Repetir tests y prueba de proceso con stdin/stdout real; verificar que diagnóstico no corrompe protocolo.
- [x] Revisar y commit acotado con evidencia de versiones y comandos.

- [ ] **Aceptación T1:** Persistir selección, reiniciar proceso y ejecutar handshake MCP sin variables de shell; comprobar mismo vault-id y error ante selección ausente. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T2 — Lectura, layouts y límites del vault

**Criterios propietarios:** US-003.AC-02. **Fuente:** SPEC v001; RC-001–008.

**Files:** `internal/vault/{vault.go,paths.go,layout.go,vault_test.go}`, `tests/fixtures/{wiki,obsidian}/`.
**Interfaces:** `vault.Open(root string) (*Vault,error)`; `(*Vault).Read(path string,offset,limit int) (ReadResult,error)`; `ReadResult{Text,Hash string; NextOffset int; Truncated bool}`; `(*Vault).Scan(ctx context.Context) (ScanResult,error)` con paths, hashes y Coverage.

- [x] Escribir casos de traversal, symlink fuera/dentro del vault, Unicode paginado, carpetas con espacios, destino no existente y permiso denegado. Límite default 20000 caracteres, hash sobre bytes completos.
- [x] Ejecutar `go test ./internal/vault` y comprobar fallo de casos nuevos.
- [x] Implementar resolución confinada, layout explícito/heredado, exclusión de `.git` y estado técnico; reportar cada exclusión/error sin ocultarlo. Sin mutar `_CLAUDE.md` ni AGENTS.md existentes.
- [x] Verificar lectura de ambas fixtures y archivo mayor que página sin pérdida/duplicación; `go test -race ./internal/vault`.
- [x] Revisar y commit; fixtures nunca incluyen notas del usuario.

- [x] **Aceptación T2:** Reconstruir una nota Unicode desde páginas, comparar bytes y hash; traversal, symlink externo y permiso denegado deben fallar sin filtrar contenido. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T3 — Lotes, mutaciones y recuperación

**Criterios propietarios:** US-004.AC-04, US-022.AC-01, US-022.AC-02, US-022.AC-03, US-022.AC-04. **Fuente:** SPEC v001; RC-001–008.

**Files:** `internal/journal/{journal.go,recovery.go,journal_test.go}`, `internal/vault/{changes.go,changes_test.go}`.
**Interfaces:** `Change{Path,Content,ExpectedHash,Destination string; Kind string}`; `(*Vault).Apply(ctx context.Context,operationID string,changes []Change,dryRun bool) (OperationResult,error)`; `(*Vault).Recover(ctx context.Context,operationID,action string) (OperationResult,error)`.

- [x] Escribir `TestCreateOnly`, `TestRawImmutable`, `TestConflictBeforeApply`, `TestFailureAfterFirstFile`, `TestRollbackRefusesHumanEdit`, `TestIdempotencyRejectsDifferentPayload`, `TestConcurrentWriters`, `TestDryRunNoVaultWrites` y `TestPurgeKeepsPendingOperations`. Purge de operaciones terminadas vencidas (30 días) exige alcance explícito; dry-run conserva snapshots y no promete borrado físico seguro.
- [x] Ejecutar suite de tarea y confirmar fallos; introducir fallo de filesystem determinista mediante interfaz interna, no sleeps.
- [x] Implementar snapshots/journal previos, bloqueo por vault y reemplazo atómico por archivo; creación exige ausencia, modificación hash exacto, recovery condicional. Resume usa journal y hashes, nunca interpreta instrucciones del contenido.
- [x] Ejecutar `go test -race ./internal/journal ./internal/vault`; inspeccionar hashes de fixture antes/después/recovery y JSON de estados.
- [x] Revisar y commit; describir límite frente a editor externo que no respeta locks.

- [ ] **Aceptación T3:** Inducir fallo tras primer archivo, reiniciar, inspeccionar/aplicar recovery y comparar hashes; probar idempotencia y purge sin eliminar pendientes. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T4 — Grafo, edición y bookkeeping

**Criterios propietarios:** US-014.AC-02, US-014.AC-03. **Fuente:** SPEC v001; RC-001–008.

**Files:** `internal/graph/{links.go,canvas.go,graph_test.go}`, `internal/vault/{notes.go,notes_test.go}`.
**Interfaces:** `graph.Build(scan vault.ScanResult) Graph`; `Graph.Backlinks(target string) []Reference`; `Graph.Canvas(scope []string) ([]byte,error)`; servicios de save/update/replace/move/capture producen Change[] para Apply. `vault.Coverage{Scanned int; Excluded,Errors []string; Truncated bool}` se define en T2 y se reutiliza sin reinterpretar exclusiones como lecturas.

- [x] Escribir fixtures para wikilinks con alias, heading, block IDs, embeds, nombres duplicados, links Markdown, code fences y backlinks; mover nota actualiza solo referencias inequívocas.
- [x] Escribir `TestCaptureIncludesDaily`, `TestMoveRefusesExistingDestination`, `TestReplaceRequiresSingleMatch`, `TestPreserveHumanSections` y `TestCanvasScope`; observar fallos.
- [x] Implementar cambios relacionados sobre nota/index/log/diaria según workflow, con fechas/TZ explícitas, preservando historia y contenido humano.
- [x] Ejecutar `go test -race ./internal/graph ./internal/vault`; validar JSON Canvas y referencias existentes.
- [x] Revisar y commit. La visualización dentro de Obsidian sigue separada de la validez del archivo.

- [ ] **Aceptación T4:** Validar Canvas/scope y links con fixtures; ejecutar dos veces conservando bloques humanos y declarar por separado cualquier render no probado. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T5 — Recuperación, salud y evaluación

**Criterios propietarios:** US-003.AC-01, US-003.AC-03, US-003.AC-04, US-013.AC-01, US-013.AC-02. **Fuente:** SPEC v001; RC-001–008.

**Files:** `internal/search/{search.go,index.go,eval.go,search_test.go}`, `internal/health/{health.go,freshness.go,health_test.go}`.
**Interfaces:** `search.Query(scan vault.ScanResult,q string,limit int) Result`; `Result{Matches []Match; Coverage vault.Coverage}`; `search.Evaluate(cases []Case,k int) Metrics`; `health.Check(scan vault.ScanResult,now time.Time) []Finding`.

- [x] Escribir ranking determinista y consultas Unicode, corpus con notas ilegibles, exclusiones, límite de resultados vs cobertura; recall@k/MRR sobre casos manualmente conocidos.
- [x] Escribir fixtures freshness FRESH-1–5, snapshots, fechas inválidas, fences y suppressions, metadata de notas vs excepción kanban; observar fallos.
- [x] Implementar índice léxico reconstruible con invalidación por hash, health y métricas; no embeddings ni claims de paridad semántica.
- [x] Ejecutar `go test ./internal/search ./internal/health`; guardar resultados numéricos del corpus sintético con revisión del corpus. No inventar baseline universal.
- [x] Revisar y commit. La brecha F-033 permanece hasta evidencia equivalente, aunque los tests léxicos pasen.

- [ ] **Aceptación T5:** Ejecutar fixtures ranking/cobertura/frescura y benchmark con recall@k/MRR esperado a mano; semántica permanece bloqueada hasta O-003. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T6 — Skills y continuidad sobre el vault

**Criterios propietarios:** US-004.AC-01, US-004.AC-02, US-005.AC-01, US-005.AC-02, US-005.AC-03, US-006.AC-01, US-006.AC-02, US-006.AC-03, US-007.AC-01, US-007.AC-02, US-008.AC-01, US-008.AC-02, US-008.AC-03, US-008.AC-04, US-009.AC-01, US-009.AC-02, US-009.AC-03, US-010.AC-01, US-010.AC-02, US-010.AC-03, US-011.AC-02, US-011.AC-03, US-012.AC-01, US-012.AC-02, US-012.AC-03, US-014.AC-01, US-021.AC-01. **Fuente:** SPEC v001; RC-001–008.

**Files:** `plugins/llm-wiki/skills/`, `plugins/llm-wiki/references/`, `internal/catalog/`, `tests/acceptance/workflows.json`, `plugins/llm-wiki/capabilities.json`.
**Interfaces:** `catalog.List(vaultRoot,pluginRoot string) ([]Skill,error)`; `catalog.Get(vaultRoot,pluginRoot,name string) (Skill,error)`; `Skill{Name,Description,Body,Source string}`. Frontmatter de cada SKILL.md válido y referencias relativas dentro del paquete.

- [x] Crear casos de aceptación por cada F/mode a partir de matriz: captura/save/person/project/task/daily/decide/log; board/hygiene/recurring; síntesis/challenge/panel/connect/emerge/review/recap; world/learn; architect/init/graduate y autoría de workflows. Mantener políticas específicas de guardar/ofrecer/dry-run.
- [x] Probar que catálogo no permite traversal ni instalación implícita y que personalizaciones sobreviven cambio del paquete; tests fallan antes de código.
- [x] Adaptar instrucciones al namespace llm-wiki, conservar procedencia y atribución, sin scripts ni herramientas inexistentes. Core carga manual/layout/contexto explícitamente; una segunda sesión recupera lo persistido, no depende del historial del chat.
- [ ] Verificar cada modalidad con fixtures y transcript de interacción/evidencia; grep de referencias eliminadas ayuda pero no valida funcionamiento.
- [x] Revisar skills con checklist de resultados del inventario, no solo conteo. Commit acotado.

- [ ] **Aceptación T6:** Ejecutar cada caso de interacción indicado abajo sobre vault sintético, capturar respuesta y diff real, comprobar política de escritura y activación explícita/opt-in. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T7 — Bootstrap, migración y exportación

**Criterios propietarios:** US-002.AC-01, US-002.AC-02, US-002.AC-03, US-002.AC-04, US-013.AC-03, US-015.AC-01, US-015.AC-02. **Fuente:** SPEC v001; RC-001–008.

**Files:** `internal/vault/{bootstrap.go,migrate.go,merge.go}`, `internal/export/{export.go,export_test.go}`, tests correspondientes.
**Interfaces:** `BootstrapSpec{Root,Name,Style,Preset,Timezone string; Assistant bool}`; `vault.PlanBootstrap(spec BootstrapSpec) ([]Change,error)`; `export.Build(scan vault.ScanResult,format string) (Bundle,error)`; `Bundle{Files map[string][]byte}`.

- [x] Escribir fixtures de estilos wiki/Obsidian, presets executive/builder/creator/researcher, assistant mode y vault existente; bootstrap jamás pisa contenido previo.
- [x] Escribir merge dry-run con redirect y backlinks, migración de log repetida/colisión y exports JSON/Markdown/OKF; observar fallos.
- [x] Implementar planes de cambio por Apply, preservando fuentes y archivos de usuario; validar OKF contra la especificación de referencia y casos con adjuntos/rutas especiales.
- [x] Ejecutar `go test ./internal/vault ./internal/export`; dos ejecuciones idempotentes y verificaciones de hashes.
- [x] Revisar y commit.

- [x] **Aceptación T7:** Para US-013.AC-03 ejecutar dos subcasos obligatorios: healing exacto/slug/ambiguo con el módulo de T4 (ignorar fences, no arreglar ambiguos) y merge dry-run/apply con redirect/procedencia. Registrar ambas evidencias; el criterio no pasa solo por merge. Ejecutar además ambos layouts/presets y migración/merge/export; comparar archivos esperados, dry-run y segunda ejecución. Validar formatos y rechazo de colisiones. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T8 — Investigación y capacidades dependientes del host

**Criterios propietarios:** US-007.AC-03, US-011.AC-01, US-011.AC-04, US-013.AC-04, US-016.AC-01, US-016.AC-02, US-016.AC-03, US-016.AC-04, US-016.AC-05, US-017.AC-01, US-017.AC-02, US-017.AC-03, US-018.AC-01, US-018.AC-02, US-018.AC-03, US-018.AC-04, US-019.AC-01, US-019.AC-02, US-019.AC-03, US-019.AC-04, US-020.AC-01, US-020.AC-02, US-020.AC-03, US-021.AC-02, US-021.AC-03, US-021.AC-04. **Fuente:** SPEC v001; RC-001–008.

**Files:** skills research/deep/synthesis/media/calendar/catchup; `tests/acceptance/capabilities.json`; actualizar `capabilities.json` y matrices solo con evidencia.
**Interfaces:** skills reciben capacidades realmente expuestas en la sesión y resultados con fuentes/fechas; el núcleo no llama a esas APIs. Ausencia devuelve diagnóstico `capability_unavailable`.

- [x] Probar investigación web y síntesis de hasta 12 notas con fuentes/cobertura, propagación y límite; ningún envío a proveedor propio.
- [ ] Probar negativos: sin web, post privado, podcast sin transcript, video sin frames, calendario sin acceso y cola Telegram ausente. Calendar tiene cuatro casos independientes: agenda sin lectura externa, reconcile sin comparar eventos inaccesibles, meeting con datos aportados y schedule sin afirmar creación externa. Ningún resultado simulado ni conector instalado para sortear AC-019.
- [ ] Separar modo manual de captura de bot y lectura de transcript de transcripción de audio. Preservar pruebas por modalidad en matriz; no llamarlos equivalencia completa.
- [ ] Probar hooks/automatizaciones únicamente con horarios/eventos expresamente configurados en entorno de ensayo; rutina nocturna no resuelve conflictos mediante reescritura.
- [ ] Registrar brechas con evidencia. Si no hay ruta nativa, solicitar decisión concreta de alcance antes de declarar modalidad entregada; resto de tareas independientes puede continuar.
- [ ] Revisar y commit. T8 no puede quedar completa solo porque todos los negativos fallan de forma clara.

- [ ] **Aceptación T8:** Ejecutar cada modalidad con capacidad presente y ausente. Guardar fuentes/salida/archivos y evidencia del host; donde no exista ruta nativa marcar blocked con O asociado, sin simular éxito. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T9 — Paquete y marketplace

**Criterios propietarios:** US-001.AC-01, US-015.AC-03. **Fuente:** SPEC v001; RC-001–008.

**Files:** `cmd/package/main.go`, manifiestos de `plugins/llm-wiki`, `.agents/plugins/marketplace.json`, `tests/acceptance/package_test.go`, README/avisos MIT.
**Interfaces:** herramienta Go `go run ./cmd/package --output <dir>` genera paquete autocontenido; executable `bin/llm-wiki serve`; metadatos/checksums de build.

- [x] Tests de manifiesto/rutas, skills faltantes, symlinks externos, ejecutable ausente, namespace y ausencia de secretos/paths privados; comprobar fallo antes de generar paquete.
- [x] Generar binarios arm64 y amd64 con CGO_ENABLED=0/trimpath; ensamblar universal con lipo. Fijar versión SDK/Go y fuente de build; no fetch en arranque.
- [ ] Validar esquemas portable y compatibilidad Codex; usar `${PLUGIN_ROOT}` solo tras prueba host. El catálogo referencia árbol que contiene binario, no assets remotos que Codex no descarga.
- [x] Probar copia aislada sin `.repos`, Go, Python o npm en PATH, handshake y herramientas; verificar checksums, permisos y arquitecturas del binario.
- [x] Revisar y commit del source y paquete según estrategia Git reproducible. No publicar remoto.

- [ ] **Aceptación T9:** Instalar paquete de ensayo completo sin toolchain de desarrollo, descubrir skills/MCP y hacer lectura; conformance y documentación coinciden con ese artefacto. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

### T10 — Instalación Desktop, actualización y recuperación real

**Criterios propietarios:** US-001.AC-03, US-001.AC-04, US-004.AC-03. **Fuente:** SPEC v001; RC-001–008.

**Files:** `tests/acceptance/desktop.md`, `docs/superpowers/analysis/equivalencias-codex/RESULTADOS.md`, actualizar plan/infra con hechos.
**Interfaces:** instalación soportada de marketplace local; vault sintético. Registrar version Desktop, arquitectura macOS, versión plugin, commit y logs sin notas personales.

- [x] Ejecutar suite completa `go test -race ./...`, `go vet ./...` y validadores de paquete; fallos bloquean instalación de ensayo.
- [ ] Instalar desde marketplace mediante flujo soportado, abrir nuevo chat y comprobar descubrimiento y llamada real. No sustituir por solo probar CLI. Probar además proyecto sin trust, rechazo de lectura/escritura MCP y hook no disponible: no mutar ni presentar ejecución exitosa, informar el permiso/evento faltante.
- [ ] Ejecutar continuidad entre dos chats con captura/decisión/proyecto, edición humana entre sesiones y citas de recuperación.
- [ ] Actualizar paquete y volver a versión previa compatible, comparar vault/config/workflows antes/después; si no hay segundo build, caso queda pendiente.
- [ ] Fallar lote y reiniciar proceso; demostrar recovery sin pérdida humana. Ejecutar binario en cada arquitectura declarada soportada; cross-compile no cuenta.
- [ ] Revisar directamente sin subagentes por instrucción del usuario, corregir hallazgos y consolidar evidencias por F/I/T/mode. Registrar cualquier intervención/reinicio Desktop aún necesaria.
- [ ] Commit final acotado; no marcar completo si quedan pruebas o brechas. No publicar ni cambiar visibilidad del repositorio.

- [ ] **Aceptación T10:** Ejecutar dos chats Desktop, actualización/downgrade y rechazo de trust/permisos con vault sintético; registrar plataforma/versiones y hashes antes/después. Registrar un caso por cada criterio propietario en `tests/acceptance/results.json` con criterion_id, entrada, salida observada, resultado pending/blocked/passed/failed, versión/commit y ruta a evidencia. Los criterios concretos están en la tabla de aceptación de este plan.

## Verificación del plan antes de ejecutar

Auto-revisión: contratos T1–T5 preceden consumidores, cada bloque tiene pruebas negativas y evidencia de ejecución; matrices cubren todas las filas sin convertir bloqueos en éxitos. T6/T8 requieren evidencia de comportamiento del agente, no únicamente tests Go. T8 y soporte Intel pueden bloquear la paridad final. La ejecución directa sin subagentes sustituye el uso de Luna por la última instrucción del usuario; las brechas conservan sus requisitos de evidencia.

## Trazabilidad completa de comandos a tareas y casos

Cada celda lista familias de casos en `tests/acceptance/workflows.json`; cada
modo tiene resultado esperado y política de escritura del inventario. Los casos
negativos y alternativos no certifican equivalencia del modo original.

| ID | Workflow | Tarea propietaria | Casos/modalidades obligatorias |
|---|---|---|---|
| F-001 | create-command | T6 | entrevista, fast-path, descubrimiento posterior |
| F-002 | idea-discovery | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-003 | notebooklm | T8 | hasta 12 notas, límite, síntesis, propagación, sin Gemini |
| F-004 | obsidian-architect | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-005 | obsidian-board-hygiene | T6 | propuesta, lote aprobado, sin aprobación |
| F-006 | obsidian-board | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-007 | obsidian-brainstorm | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-008 | obsidian-calendar | T8 | agenda, reconcile, meeting, schedule; acceso ausente por modo |
| F-009 | obsidian-capture | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-010 | obsidian-catchup | T8 | today/week/all, texto/voz/imagen/PDF/link; cola ausente |
| F-011 | obsidian-challenge | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-012 | obsidian-connect | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-013 | obsidian-daily | T6 | crear/actualizar; calendario ausente |
| F-014 | obsidian-decide | T6 | ligero, formal |
| F-015 | obsidian-distill | T6 | nota, archivo, URL; raw intacto y procedencia |
| F-016 | obsidian-emerge | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-017 | obsidian-export | T7 | json, markdown, okf |
| F-018 | obsidian-find | T5 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-019 | obsidian-graduate | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-020 | obsidian-health | T5 | diagnóstico, arreglos ofrecidos, destructivo confirmado, exclusiones |
| F-021 | obsidian-ingest | T6 | fuente nueva, entidades afectadas, claims viejos, contradicción, raw intacto |
| F-022 | obsidian-init | T7 | existente/nuevo, manual delimitado, index/log |
| F-023 | obsidian-learn | T6 | recent/all/topic; promover/archivar solo confirmado |
| F-024 | obsidian-log | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-025 | obsidian-merge | T7 | dry-run, apply, redirect y conflictos |
| F-026 | obsidian-panel | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-027 | obsidian-person | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-028 | obsidian-project | T6 | resultado y escrituras de matriz; repetición, faltantes y conflicto |
| F-029 | obsidian-projects | T6 | uno/todos, Git disponible/ausente, Last overview |
| F-030 | obsidian-recap | T6 | hoy/semana/mes, respuesta sin guardar |
| F-031 | obsidian-reconcile | T6 | ganador claro, evolución, ambiguo; historial/index/log/diaria |
| F-032 | obsidian-recurring | T6 | alta, cumplimiento, next-due, fin de mes y TZ |
| F-033 | obsidian-reindex | T5 | reindex léxico; semántico bloqueado hasta equivalencia medida |
| F-034 | obsidian-retrieval-eval | T5 | evaluar, report, generar/reemplazar casos explícitos |
| F-035 | obsidian-review | T6 | semana/mes |
| F-036 | obsidian-save | T6 | decisiones, tareas, personas, proyectos, ideas; propagación |
| F-037 | obsidian-synthesize | T6 | interactivo, programado; duplicados y cobertura |
| F-038 | obsidian-task | T6 | tarea simple/sustancial, prioridad/fecha |
| F-039 | obsidian-visualize | T4 | scope exacto/vault; métricas/Canvas |
| F-040 | obsidian-world | T6 | niveles de contexto, PINNED crear/limpiar, faltantes |
| F-041 | podcast | T8 | transcript presente/ausente, audio sin capacidad |
| F-042 | research-deep | T8 | delta/vault/web, propagación, ruta sin proveedores |
| F-043 | research | T8 | dossier/citas/fechas/disenso, ruta sin proveedores |
| F-044 | vault-deep-synthesis | T6 | cobertura total declarada, contradicciones sin alterar fuentes |
| F-045 | x-pulse | T8 | ventana reciente, búsqueda X no disponible |
| F-046 | x-read | T8 | público/privado/inaccesible, chat-only, guardar explícito |
| F-047 | youtube | T8 | transcripción, metadatos/comentarios, visual por escenas; ausencia por modo |

## Trazabilidad de integraciones y ciclo de vida

Los criterios completos permanecen en la matriz I/T del commit de diseño. Los
siguientes son los propietarios y pruebas positivas necesarias; si el host no
puede producir el resultado, registrar blocked y no reemplazarlo por un negativo.

| Registros | Tarea | Prueba positiva o bloqueo explícito |
|---|---|---|
| I-001 | T1–T5 | Doce operaciones MCP originales mapeadas a wiki_* con lectura/paginación/escritura/backlinks/health/catálogo en fixture instalada |
| I-002 | T5/T8 | Lint Go por regla; experiencia editor Obsidian blocked sin adaptador |
| I-003 | T8 | Inbox texto/adjuntos procesable; recepción y respuesta Telegram blocked sin conexión |
| I-004 | T8 | Investigación general y académica con fuentes accesibles, cobertura/errores/citas; sin ruta pagada propia |
| I-005 | T8 | Baseline de vault, gap analysis, recuperación dirigida y delta propagado |
| I-006 | T8 | Publicación/hilo accesible con texto verificable y límites; tendencias/respuestas completas blocked si no accesibles |
| I-007 | T8 | Transcript/frames con procedencia → nota/adjuntos; extracción de comentarios/escenas blocked sin capacidad |
| I-008 | T8 | Feed/transcript accesible → nota citada; audio sin transcriptor nativo blocked |
| I-009 | T8 | Síntesis hasta 12 notas, citas y propagación sin subir a Gemini; calidad del resultado medida |
| I-010 | T8 | Leer página disponible en Codex y reportar truncamiento; sin confundir snippet con texto completo |
| I-011 | T8/T9 | No clientes/claves; errores nativos y guardado con procedencia; equivalencia de resultados, no de modelos |
| I-012 | T8 | Procesamiento de tipos de transcript disponibles y límites multimedia; pipeline extracción automática blocked si falta |
| T-001, T-002, T-003, T-004, T-005, T-006, T-008 | T6/T10 | Reglas, contexto, propagación e historia recuperables en nueva sesión |
| T-007, T-009, T-011, T-012 | T2/T3/T7 | raw intacto, dos layouts, presets, delimitadores y bootstrap conservador |
| T-010 | T5 | FRESH-1–5 con exclusiones, snapshots y suppressions |
| T-013, T-014, T-021, T-022, T-023, T-024, T-025, T-026 | T9/T10 | Sustituir instalación/adapters legacy por marketplace Codex; no construir otros harnesses; actualización con preservación |
| T-015, T-016, T-017, T-018, T-019, T-020 | T8/T10 | Contexto/validación al evento disponible, opt-in, locks y cuatro workflows morning/nightly/weekly/health-check; sin evento nativo queda blocked |
| T-027, T-028, T-029, T-030, T-031, T-032, T-033 | T4/T5/T7 | Salud, healing exacto, merge, migración, stats, scan/ADR/encoding y OKF con archivos esperados |
| T-034, T-035 | T5/T8 | Corpus reproducible, recall/MRR y diagnóstico; evaluación A/B por otro contexto Codex si disponible, no juez externo API |
| T-036 | T9 | Conformance por archivo; generación local de documentación de comandos, sin publicar sitio |
| T-037, T-038, T-039, T-040, T-041 | T6 | Skills por workflow, guards/sentinels, triggers y escritura específica verificadas |
| T-042 | T1/T8 | Persistencia MCP verificada; editor Obsidian y Telegram documentados como brechas |
| T-043, T-044 | T9/T10 | Personalización persiste, sin claves/telemetría; actualizar no sobrescribe datos |

Ampliaciones concretas de tests por esta revisión: T4 incluye healing exacto/slug
y ambiguo sin arreglo; T5 prueba estadísticas con carpetas excluidas; T6 prueba
scan Git y extracción ADR con commits sintéticos/Unicode sin convertir mensajes
ambiguos en decisiones; T9 genera la referencia local de workflows y su conformance.
Estos son resultados del inventario, no funcionalidades nuevas.

## Revisión independiente

Tres subagentes GPT-6 Luna revisaron cobertura de comandos, integraciones/ciclo y
empaquetado. Se incorporaron casos por modalidad de calendario, trazabilidad F/I/T,
pruebas positivas condicionadas, arranque `serve`, persistencia del selector,
actualización/recovery, trust y retención explícita. FFmpeg/yt-dlp siguen fuera
de la propuesta inicial, como decisión técnica propuesta a reconciliar con AC-011,
no como prohibición histórica atribuida al usuario.

## Aceptación trazada a la spec

Cada fila es un caso independiente y está asignada al paso **Aceptación Tn**
de una tarea anterior. El identificador remite al criterio completo de la spec fijada;
ninguno está ejecutado. No crear un runner que declare passed por presencia del
archivo: se compara comportamiento observado con el resultado esperado completo.

| Criterio normativo de SPEC v001 | Tarea y paso que lo ejecutan | Caso de evidencia | Estado |
|---|---|---|---|
| US-001.AC-01 | T9, Aceptación T9 | `us-001.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-001.AC-02 | T1, Aceptación T1 | `us-001.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-001.AC-03 | T10, Aceptación T10 | `us-001.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-001.AC-04 | T10, Aceptación T10 | `us-001.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-002.AC-01 | T7, Aceptación T7 | `us-002.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-002.AC-02 | T7, Aceptación T7 | `us-002.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-002.AC-03 | T7, Aceptación T7 | `us-002.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-002.AC-04 | T7, Aceptación T7 | `us-002.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-003.AC-01 | T5, Aceptación T5 | `us-003.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-003.AC-02 | T2, Aceptación T2 | `us-003.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-003.AC-03 | T5, Aceptación T5 | `us-003.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-003.AC-04 | T5, Aceptación T5 | `us-003.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-004.AC-01 | T6, Aceptación T6 | `us-004.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-004.AC-02 | T6, Aceptación T6 | `us-004.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-004.AC-03 | T10, Aceptación T10 | `us-004.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-004.AC-04 | T3, Aceptación T3 | `us-004.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-005.AC-01 | T6, Aceptación T6 | `us-005.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-005.AC-02 | T6, Aceptación T6 | `us-005.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-005.AC-03 | T6, Aceptación T6 | `us-005.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-006.AC-01 | T6, Aceptación T6 | `us-006.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-006.AC-02 | T6, Aceptación T6 | `us-006.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-006.AC-03 | T6, Aceptación T6 | `us-006.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-007.AC-01 | T6, Aceptación T6 | `us-007.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-007.AC-02 | T6, Aceptación T6 | `us-007.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-007.AC-03 | T8, Aceptación T8 | `us-007.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-008.AC-01 | T6, Aceptación T6 | `us-008.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-008.AC-02 | T6, Aceptación T6 | `us-008.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-008.AC-03 | T6, Aceptación T6 | `us-008.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-008.AC-04 | T6, Aceptación T6 | `us-008.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-009.AC-01 | T6, Aceptación T6 | `us-009.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-009.AC-02 | T6, Aceptación T6 | `us-009.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-009.AC-03 | T6, Aceptación T6 | `us-009.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-010.AC-01 | T6, Aceptación T6 | `us-010.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-010.AC-02 | T6, Aceptación T6 | `us-010.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-010.AC-03 | T6, Aceptación T6 | `us-010.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-011.AC-01 | T8, Aceptación T8 | `us-011.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-011.AC-02 | T6, Aceptación T6 | `us-011.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-011.AC-03 | T6, Aceptación T6 | `us-011.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-011.AC-04 | T8, Aceptación T8 | `us-011.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-012.AC-01 | T6, Aceptación T6 | `us-012.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-012.AC-02 | T6, Aceptación T6 | `us-012.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-012.AC-03 | T6, Aceptación T6 | `us-012.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-013.AC-01 | T5, Aceptación T5 | `us-013.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-013.AC-02 | T5, Aceptación T5 | `us-013.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-013.AC-03 | T7, Aceptación T7 | `us-013.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-013.AC-04 | T8, Aceptación T8 | `us-013.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-014.AC-01 | T6, Aceptación T6 | `us-014.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-014.AC-02 | T4, Aceptación T4 | `us-014.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-014.AC-03 | T4, Aceptación T4 | `us-014.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-015.AC-01 | T7, Aceptación T7 | `us-015.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-015.AC-02 | T7, Aceptación T7 | `us-015.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-015.AC-03 | T9, Aceptación T9 | `us-015.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-016.AC-01 | T8, Aceptación T8 | `us-016.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-016.AC-02 | T8, Aceptación T8 | `us-016.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-016.AC-03 | T8, Aceptación T8 | `us-016.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-016.AC-04 | T8, Aceptación T8 | `us-016.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-016.AC-05 | T8, Aceptación T8 | `us-016.ac-05`; contrastar el criterio íntegro de la spec | pending |
| US-017.AC-01 | T8, Aceptación T8 | `us-017.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-017.AC-02 | T8, Aceptación T8 | `us-017.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-017.AC-03 | T8, Aceptación T8 | `us-017.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-018.AC-01 | T8, Aceptación T8 | `us-018.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-018.AC-02 | T8, Aceptación T8 | `us-018.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-018.AC-03 | T8, Aceptación T8 | `us-018.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-018.AC-04 | T8, Aceptación T8 | `us-018.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-019.AC-01 | T8, Aceptación T8 | `us-019.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-019.AC-02 | T8, Aceptación T8 | `us-019.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-019.AC-03 | T8, Aceptación T8 | `us-019.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-019.AC-04 | T8, Aceptación T8 | `us-019.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-020.AC-01 | T8, Aceptación T8 | `us-020.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-020.AC-02 | T8, Aceptación T8 | `us-020.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-020.AC-03 | T8, Aceptación T8 | `us-020.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-021.AC-01 | T6, Aceptación T6 | `us-021.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-021.AC-02 | T8, Aceptación T8 | `us-021.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-021.AC-03 | T8, Aceptación T8 | `us-021.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-021.AC-04 | T8, Aceptación T8 | `us-021.ac-04`; contrastar el criterio íntegro de la spec | pending |
| US-022.AC-01 | T3, Aceptación T3 | `us-022.ac-01`; contrastar el criterio íntegro de la spec | pending |
| US-022.AC-02 | T3, Aceptación T3 | `us-022.ac-02`; contrastar el criterio íntegro de la spec | pending |
| US-022.AC-03 | T3, Aceptación T3 | `us-022.ac-03`; contrastar el criterio íntegro de la spec | pending |
| US-022.AC-04 | T3, Aceptación T3 | `us-022.ac-04`; contrastar el criterio íntegro de la spec | pending |

## Bloqueos y dependencias de ejecución

| Spec | Tareas | Tratamiento |
|---|---|---|
| O-001/002 | T8 | Conservar cada modo de X/calendario/Telegram/medios como blocked hasta evidencia nativa o decisión de alcance; no añadir proveedores |
| O-003 | T5 | Tests léxicos ejecutables, aceptación semántica pendiente de corpus/umbral/equivalencia |
| O-004 | T4/T5/T8 | Archivo válido/lint Go no cierra criterio de experiencia del editor |
| O-005 | T8/T10 | Prueba real de evento/trust/scheduler; ruta manual no equivale a automatización |
| O-006 | T3 | Revisar retención/purge propuestos antes de implementar esa política; no impedir lectura y diseño del journal |
| O-007 | T9/T10 | Compilar no demuestra soporte; registrar plataformas sin hardware de ensayo como pendientes |
| O-008 | T10 | Separar pruebas técnicas de umbrales/horizonte del resultado con usuarios |
| O-009 | T8 | Verificar datos nativos de consumo y alcance; si no disponibles registrar brecha, sin tarifas externas inventadas |
| Spec/plan draft | Todas | Presentar ambos; aprobación pendiente, no iniciar implementación por existencia de estos archivos |

## Historial de esta revisión

v002, 2026-10-08: se sustituye la arquitectura como falsa fuente de spec por la
especificación funcional real. Se asignan todos los criterios a tareas con pasos
de verificación, se preservan los registros F/I/T y el método Luna y se explicitan
bloqueos. La revisión independiente descrita antes corresponde al plan previo;
no constituye aprobación ni verificación funcional del nuevo contenido.

Revisión independiente adicional del plan v002 por GPT-6 Luna: pin de spec y
79/79 criterios comprobados. Se explicitó la doble comprobación healing/merge
para US-013.AC-03 dentro de Aceptación T7. Verificación documental, no funcional.

## AP-001 — aprobación y ejecución

El 2026-10-08 el usuario promotor declaró: «Revisado y aprobado. Implementa con luna».
Aprueba la revisión presentada de este documento y autoriza ejecución con
subagentes GPT-6 Luna. Base exacta: `726af251276d9756a9b193782d4f7009cff63fd5:docs/superpowers/plans/2026-10-08-codex-plugin.md`.
Los asuntos abiertos se conservan como pendientes; no se afirma paridad ni
se autoriza publicación externa por esta aprobación. Las menciones draft/revisión
pendiente del cuerpo describen su presentación anterior y quedan resueltas
por este registro únicamente para el alcance escrito.

### Avance directo — T6

Catálogo, contexto paginado y continuidad persistida expuestos por MCP y CLI. `go test -race ./...` y `go vet ./...` pasan; el intercambio STDIO guarda/lee y rechaza un hash obsoleto. La aceptación de las 47 modalidades continúa pendiente de interacción y distribución; no se equipara con estas pruebas del núcleo.

### Avance directo — T7

Carpetas recuperables, layout explícito persistido, bootstrap de ambos layouts/cuatro presets, migración con archivo íntegro, merge conservador, exportación JSON/Markdown/OKF y autoría de workflows integrados en MCP/CLI. Suite completa con race y vet satisfactoria. Continúan pendientes la fidelidad completa de plantillas/seed notes, init sobre bloques existentes y aceptación de interacción por modalidad. La casilla global de implementación T7 permanece abierta por esas diferencias.
