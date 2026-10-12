---
title: Ejecución y distribución de LLM Wiki
status: draft
version: v006
responsible: Usuario promotor del proyecto
created_at: "2026-10-08"
updated_at: "2026-10-12"
reviewed_at: null
approval_reference: null
---

# Ejecución y distribución de LLM Wiki

Entrega relacionada: **v0.2.2**; runtime/manifiestos `0.2.2`.
Nombre visible **LLM Wiki**; identificador técnico `llm-wiki` conservado.
[Changelog](../../CHANGELOG.md), [informe 0.2.2](../../tests/acceptance/release-0.2.2.json) y [verificación Desktop 0.2.2](evidence/2026-10-12-desktop-0.2.2.md).

Estado de entrega: paquete `v0.2.2`; identidad, procedencia, checksums y estado
de publicación/instalación se registran en el informe de esta entrega. La prueba
MCP del paquete sin toolchain cubre los ocho cruces de layout/preset. La aceptación
Desktop de dos chats sobre 0.2.0 se conserva histórica. En 0.2.2 se verificaron
la carga de esquemas y el flujo MCP probado en un chat nuevo; refrescar sesiones
existentes, ejecutar Intel y renderizar vistas Obsidian requieren evidencia propia.
Fuentes: [arquitectura](ARCHITECTURE.md) y
[estudio de empaquetado](analysis/equivalencias-codex/EMPAQUETADO.md).

## Entornos

| Entorno | Función | Evidencia actual |
|---|---|---|
| Desarrollo macOS | Go, SDK MCP, tests, generación de paquete | Go 1.27.2 aislado, SDK MCP Go 1.8.0; race/vet pasan |
| Paquete v0.2.2 | Binario Go universal, 50 skills y manifiestos | Procedencia y aceptación MCP sin toolchain en el informe 0.2.2; aceptación Desktop 0.2.0 separada |
| Codex Desktop macOS | Carga de skills y proceso MCP local STDIO | 0.2.2: 30 herramientas expuestas, 28 llamadas a 12 herramientas en un chat nuevo; aceptación de dos chats 0.2.0 histórica, refresco de sesiones existentes pendiente |
| Distribución Git | Catálogo y paquete por referencia versionada | Repositorio GitHub `santiago-migoni/llm-wiki`; referencia `v0.2.2`; estado remoto en informe 0.2.2. PR #1 previo integrado en main |

Se compilan arm64 y amd64 y se ensamblan con `lipo` en desarrollo. Esto no prueba
que ambas arquitecturas funcionen: cada una exige ejecución real antes de figurar
como soportada. La versión mínima de macOS/Desktop se fijará con esa evidencia,
no por suposición a partir del host del desarrollador.

## Configuración y persistencia

- Configuración implementada: `os.UserConfigDir()/llm-wiki/registry.json`, esquema 2; override `--state-dir`.
- Estado por vault: `os.UserConfigDir()/llm-wiki/state/<vault-id>/`.
- `vault-id`: identificador persistente asociado a raíz canónica, no título mutable.
- Asociación de proyecto: ruta canónica del proyecto a vault-id. Selección explícita;
  si hay más de un candidato se devuelve `vault_not_selected`, no se adivina.
- En herramientas MCP se pasa vault-id; nunca se permite un path fuera de su raíz.
- No se heredan rutas del shell ni se empaquetan paths personales en el manifiesto.
- No se guardan claves de servicios: no existen clientes propios de esos servicios.
- Estado con permisos restringidos; journals incluyen contenido previo sensible y
  se conservan localmente. Propuesta: retención de operaciones terminadas de 30 días,
  con purge explícito que primero muestra el alcance. Nunca podar operaciones
  pendientes/fallidas. El diagnóstico informa tamaño y snapshots vencidos; no hay
  borrado automático. Borrar una nota no borra su historial. Purge no promete
  borrado físico seguro en SSD ni de copias de respaldo.

## Arranque y recursos

Los manifiestos portable y compatible declaran `./bin/llm-wiki serve` con
`cwd: "./"`, relativo a la raíz instalada del plugin. Ambos descriptores se
validan y la aceptación del paquete los ejecuta desde una ruta con espacios,
sin requerir expansión de variables en el comando. La disponibilidad en los
chats exige además comprobar la carga real del host. STDOUT contiene únicamente
MCP; STDERR diagnósticos sin contenido de
notas ni credenciales. No daemon HTTP, puertos, servicios de nube, telemetry ni
descargas en arranque. El proceso Go no llama a LLMs ni endpoints externos.
Codex puede utilizar sus capacidades de red disponibles para los workflows; esto
no es una promesa de funcionamiento offline del agente completo.

## Actualización y recuperación

El artefacto contiene fuente de versión y checksums. El catálogo Git referencia
contenido reproducible. Cambiar paquete no mueve vault/config/estado ni workflows
personales. Esquema de estado versionado: una versión que no lo entiende rechaza
arranque de escritura; no lo transforma silenciosamente. Las primeras versiones
mantienen compatibilidad hacia atrás antes de habilitar downgrade.

Cada mutación verifica hashes, registra el estado previo y publica archivos con
reemplazo atómico individual. No se afirma atomicidad global del lote. Un fallo
parcial bloquea nuevas escrituras del vault hasta inspeccionar/reanudar o revertir
con precondiciones. Rollback no pisa ediciones posteriores de otra aplicación.
No se fuerza un bloqueo abandonado sin inspección del journal y proceso dueño.

## Verificación operativa exigida

1. Paquete desde un checkout limpio sin `.repos`, Go ni herramientas de desarrollo
   en PATH; validar identidad, archivos, ejecutable y rutas relativas.
2. Instalar desde catálogo local/Git en Desktop, abrir chat nuevo, descubrir skills
   y MCP, seleccionar un vault sintético y ejecutar lectura inocua.
3. Guardar una decisión, cerrar chat/proceso, recuperar en otro chat citando la nota.
4. Actualizar versión conservando bytes del vault y personalizaciones; ensayar volver
   a una versión previa con estado compatible. Registrar versiones de host y plugin.
5. Interrumpir un lote, reiniciar, recuperar y verificar hashes de todos los archivos.
6. Probar proceso real por cada arquitectura declarada compatible. Compilar no basta.

El siguiente registro de instalación y autorización v0.2.0 es histórico.
El alta de marketplace y la instalación local forman parte de la etapa solicitada
por el usuario, y se ejecutaron mediante Codex CLI 0.135.0. La interfaz Desktop no pudo verificarse por restricción de la herramienta de control visual. La entrega `v0.2.0` tiene autorización explícita para commit, tag y push; una
GitHub Release, el directorio público y la instalación de esta versión quedan pendientes. No se reinicia la app durante un chat activo
sin coordinar el momento; si exige reinicio, la comprobación se registra pendiente.

## Automatizaciones y Obsidian

Usar automatizaciones de Codex con horarios expresamente elegidos por el usuario;
no crear cron, launchd ni otro scheduler. Los hooks son opt-in y necesitan probar
el evento y trust del host. Sin evento equivalente se ofrece invocación explícita,
registrada como diferencia funcional; no se anuncia persistencia automática probada.
Se generan formatos `.canvas`/`.base` compatibles como datos; operar esos archivos
no exige abrir Obsidian. La vista en vivo/lint del editor no se declara reemplazada.

## Evidencia histórica de la revisión dev.2

Fuente del paquete dev.2: `43ee802`; macOS 27.0.1 arm64. Se probó alternar procesos dev.1/dev.2 sin cambios en hashes del vault/estado y recuperar un lote tras salida del proceso, conservando edición humana posterior. El [registro de implementación](evidence/2026-10-08-implementation.md) detalla comandos, criterios y límites. No prueba downgrade del marketplace, dos chats Desktop, trust/permisos ni ejecución Intel. En ese ensayo no había publicación remota.

## Evidencia de la entrega v0.2.0

Fuentes del binario: `22194406f7d9d7137b3f8d38fcf4824b98e6897a`;
procedencia histórica en el informe 0.2.0; build.json/checksums.json identifican ahora el paquete vigente.
La [evidencia de áreas](evidence/2026-10-10-areas.md) y el
[informe de entrega](../../tests/acceptance/release-0.2.0.json) registran pruebas
y pendientes. El tag `v0.2.0` incluye documentación/changelog y esas fuentes;
La aceptación posterior de dos chats está registrada separadamente; Intel sigue pendiente.

## Entrega local v0.2.1

Fuentes y artefactos en [release-0.2.1.json](../../tests/acceptance/release-0.2.1.json).
La revisión usa origin/main actualizado por fetch; commit y tag son locales.
No se reinstala Desktop ni se publica por preparar esta entrega. Un proceso nuevo
del cargador vio los schemas corregidos en 0.2.0; las sesiones existentes seguían
mostrando arrays de strings. La prueba pendiente es refrescar su declaración.

## Entrega v0.2.2

El [informe 0.2.2](../../tests/acceptance/release-0.2.2.json) identifica el paquete
reconstruido y sus comprobaciones. La aceptación MCP usa registros y bóvedas
sintéticos, PATH sin toolchain y procesos reales del binario empaquetado. Las
[regresiones de plantillas](evidence/2026-10-12-template-regressions.md) cubren
60 plantillas, 24 Bases y dos colas en los ocho escenarios; parsear Bases no
demuestra su ejecución en Obsidian. Actualizar el plugin preserva las plantillas
humanas y no migra bóvedas existentes.

La [verificación Desktop 0.2.2](evidence/2026-10-12-desktop-0.2.2.md) y su
[registro de llamadas](../../tests/acceptance/desktop-0.2.2.json) cierran la carga
de esquemas en un chat nuevo y el funcionamiento del flujo probado. Vault, proyecto
y journals por vault son sintéticos; el registro global del servidor usa su
configuración compartida. Bootstrap `obsidian`/`researcher` quedó committed y su
repetición sin cambios; se inspeccionó una operación committed, sin forzar caída,
resume ni rollback. Manifiestos, build y checksum corroboran la versión instalada;
las respuestas de las herramientas probadas no informan la versión del proceso
y no se observó su ruta. No se acredita refresco de sesiones existentes ni se
repite la aceptación completa de dos chats de 0.2.0.
