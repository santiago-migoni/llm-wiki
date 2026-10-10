---
title: Ejecución y distribución de llm-wiki
status: draft
version: v003
responsible: Usuario promotor del proyecto
created_at: "2026-10-08"
updated_at: "2026-10-10"
reviewed_at: null
approval_reference: null
---

# Ejecución y distribución de llm-wiki

Entrega relacionada: **v0.2.0**; runtime/manifiestos `0.2.0`.
[Changelog](../../CHANGELOG.md) y [evidencia de áreas](evidence/2026-10-10-areas.md).

Estado de entrega: paquete `v0.2.0` verificado en macOS arm64, con firma ad hoc,
checksums y procedencia; MCP con PATH del sistema y persistencia comprobados.
Instalación CLI histórica: 0.1.0-dev.2. Instalación de 0.2.0/Desktop e Intel pendientes.
Fuentes: [arquitectura](ARCHITECTURE.md) y
[estudio de empaquetado](analysis/equivalencias-codex/EMPAQUETADO.md).

## Entornos

| Entorno | Función | Evidencia actual |
|---|---|---|
| Desarrollo macOS | Go, SDK MCP, tests, generación de paquete | Go 1.27.2 aislado, SDK MCP Go 1.8.0; race/vet pasan |
| Paquete v0.2.0 | Binario Go universal, 50 skills y manifiestos | Fuentes `2219440`; firma, checksums, procedencia y MCP aislado verificados. Ensayos dev.1/dev.2 históricos |
| Codex Desktop macOS | Carga de skills y proceso MCP local STDIO | Evidencia CLI histórica de 48 skills; 50 empaquetadas en 0.2.0. Instalación/aceptación Desktop de esta versión pendientes |
| Distribución Git | Catálogo y paquete por referencia versionada | Repositorio GitHub `santiago-migoni/llm-wiki`; tag `v0.2.0` autorizado por el usuario. PR #1 hacia main |

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

Codex ejecuta `${PLUGIN_ROOT}/bin/llm-wiki serve` según manifiesto compatible
validado. STDOUT contiene únicamente MCP; STDERR diagnósticos sin contenido de
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
procedencia en `plugins/llm-wiki/build.json`, hashes en `checksums.json`.
La [evidencia de áreas](evidence/2026-10-10-areas.md) y el
[informe de entrega](../../tests/acceptance/release-0.2.0.json) registran pruebas
y pendientes. El tag `v0.2.0` incluye documentación/changelog y esas fuentes;
no sustituye las pruebas pendientes de instalación, dos chats reales o Intel.
