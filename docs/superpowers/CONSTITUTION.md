---
title: Constitución de LLM Wiki
status: approved
version: v005
responsible: Usuario promotor del proyecto
created_at: "2026-10-08"
updated_at: "2026-10-12"
reviewed_at: "2026-10-08"
approval_reference: "D-005 (dirección sin cambios); D-006 (nombre y referencia); D-007 (Codex sin proveedores externos) — Historial de decisiones"
---

# Constitución de LLM Wiki

Entrega relacionada: **v0.2.2**; runtime/manifiestos `0.2.2`.
Nombre visible **LLM Wiki**; identificador técnico `llm-wiki` conservado.
[Changelog](../../CHANGELOG.md), [informe 0.2.2](../../tests/acceptance/release-0.2.2.json) y [aceptación Desktop](evidence/2026-10-10-desktop-areas.md).

## Misión

Permitir que los usuarios de Codex gestionen y hagan evolucionar su conocimiento
personal y profesional desde Codex, conservando el conjunto completo de capacidades
de obsidian-second-brain y adaptando su experiencia al entorno de Codex.

Estado de la misión: approved. D-004 conserva su aprobación inicial; D-005
registra la aprobación posterior de la constitución completa sobre la base presentada.

## Visión

Que los usuarios de Codex puedan convertir fuentes, conversaciones y experiencias
en conocimiento conectado, recuperable y actualizado, y utilizarlo para aprender,
decidir y dar continuidad a su trabajo entre sesiones. El horizonte temporal queda
pendiente de definición. Estado: approved (D-005).

## Usuarios y necesidades

| Usuario prioritario | Situación o necesidad | Evidencia o fuente | Estado |
|---|---|---|---|
| Usuarios de Codex en general | Acceder desde Codex al conjunto de capacidades del plugin de referencia | D-001 y D-002 | confirmed |
| Usuarios de Codex que gestionan conocimiento personal o profesional | Capturar, recuperar, relacionar y actualizar conocimiento entre sesiones | Misión y D-005; necesidades aún sin validación empírica con usuarios | approved |

## Principios del producto

### PRI-001 — Conservación integral de capacidades

- **Criterio:** conservar todas las funcionalidades del producto de referencia como dirección del producto.
- **Aplicación:** evaluar equivalencia por el resultado que obtiene el usuario. Una limitación de Codex o una dependencia externa debe registrarse como brecha pendiente; no permite eliminar silenciosamente la capacidad. La priorización por entregas no reduce la meta del producto. Las capacidades pueden adoptar mecanismos distintos si preservan su valor; cualquier reducción requiere una decisión explícita.
- **Status:** approved (D-005)
- **Fuente o decisión:** D-001 confirma la intención; D-005 aprueba esta formulación y sus reglas. R-001 identifica la referencia inicial.

### PRI-002 — Experiencia dedicada a Codex

- **Criterio:** priorizar una experiencia coherente para usuarios de Codex en instalación, uso, configuración y mantenimiento.
- **Aplicación:** tomar decisiones desde las posibilidades y restricciones reales de Codex. El soporte de otros agentes no será una obligación del producto. Esto no excluye servicios externos necesarios para conservar capacidades.
- **Status:** approved (D-005)
- **Fuente o decisión:** D-001 y consulta precedente sobre una adaptación exclusivamente para Codex; aprobación en D-005.

### PRI-003 — Conocimiento bajo control del usuario

- **Criterio:** conservar la capacidad del usuario de acceder, comprender, reutilizar y recuperar su conocimiento.
- **Aplicación:** preservar fuentes y procedencia, hacer explícitos los efectos de las modificaciones y mantener vías de recuperación proporcionales al impacto. Las automatizaciones operan dentro del alcance que el usuario autorice.
- **Status:** approved (D-005)
- **Fuente o decisión:** D-005 aprueba este principio para sostener el valor de un segundo cerebro confiable.

### PRI-004 — Continuidad y confiabilidad del conocimiento

- **Criterio:** priorizar información recuperable, conectada y verificable a lo largo del tiempo.
- **Aplicación:** distinguir fuentes, inferencias, cambios temporales y contradicciones. Comunicar las limitaciones de búsqueda y actualización cuando afecten la respuesta; no presentar información incompleta como exhaustiva.
- **Status:** approved (D-005)
- **Fuente o decisión:** D-005 aprueba el principio derivado del propósito y las capacidades descritas en R-001; no aprueba el código de referencia.

### PRI-005 — Capacidades demostradas

- **Criterio:** distinguir intención de conservación, implementación y funcionamiento comprobado en Codex.
- **Aplicación:** no declarar paridad a partir de nombres de comandos, conversión de instrucciones o empaquetado. Conservar evidencia de resultados de usuario y declarar dependencias, requisitos y brechas.
- **Status:** approved (D-005)
- **Fuente o decisión:** criterio de evaluación del producto aprobado en D-005.

## Objetivos del producto

### OBJ-001 — Disponer en Codex del valor funcional completo de la referencia

- **Resultado:** los usuarios pueden realizar en Codex las actividades cubiertas por obsidian-second-brain, con resultados funcionalmente equivalentes.
- **Usuarios o necesidades atendidas:** usuarios de Codex en general que desean utilizar el plugin.
- **Principios relacionados:** PRI-001, PRI-002, PRI-005.
- **Status:** approved (D-005)
- **Fuente o decisión:** D-001 y D-002; formulación aprobada en D-005.

### OBJ-002 — Dar continuidad al conocimiento personal y profesional

- **Resultado:** los usuarios recuperan y reutilizan conocimiento previo, incorporan nuevas fuentes y mantienen relaciones y cambios sin reconstruir el contexto en cada sesión.
- **Usuarios o necesidades atendidas:** gestión de conocimiento entre sesiones.
- **Principios relacionados:** PRI-003, PRI-004.
- **Status:** approved (D-005)
- **Fuente o decisión:** resultado duradero basado en la misión, aprobado en D-005.

### OBJ-003 — Mantener el conocimiento con confianza y control

- **Resultado:** los usuarios comprenden la procedencia y evolución de su conocimiento y pueden supervisar y recuperar cambios, incluyendo los producidos por mantenimiento automatizado.
- **Usuarios o necesidades atendidas:** usuarios que delegan captura, organización y mantenimiento de su conocimiento.
- **Principios relacionados:** PRI-003, PRI-004, PRI-005.
- **Status:** approved (D-005)
- **Fuente o decisión:** aprobación en D-005.

## Criterios de éxito

### EXI-001 — Cobertura funcional demostrada en Codex

- **Objetivo:** OBJ-001.
- **Métrica o señal observable:** proporción de capacidades de la referencia con resultados equivalentes demostrados en Codex, acompañada de las brechas identificadas. El inventario debe cubrir también integraciones y automatizaciones, no solo comandos.
- **Línea base:** pendiente de inventario y medición; la inspección estructural no demuestra paridad.
- **Meta:** cobertura completa de la referencia acordada. No existe compromiso de alcanzarla en la primera entrega.
- **Plazo:** pendiente de definición.
- **Fuente de evidencia:** inventario trazable y evidencias de uso en Codex que se producirán en etapas posteriores.
- **Criterion status:** approved (D-005)

### EXI-002 — Continuidad útil entre sesiones

- **Objetivo:** OBJ-002.
- **Métrica o señal observable:** usuarios que recuperan en otra sesión conocimiento previamente incorporado y lo utilizan con fuentes y contexto suficientes para resolver una necesidad real.
- **Línea base:** pendiente de medición.
- **Meta:** umbral y muestra pendientes de definición.
- **Plazo:** pendiente de definición.
- **Fuente de evidencia:** escenarios de uso y evaluación con usuarios; aún no realizados.
- **Criterion status:** approved (D-005)

### EXI-003 — Cambios comprensibles y recuperables

- **Objetivo:** OBJ-003.
- **Métrica o señal observable:** capacidad de los usuarios de identificar qué cambió y por qué, reconocer incertidumbres y recuperar el conocimiento afectado ante un cambio incorrecto.
- **Línea base:** pendiente de medición.
- **Meta:** umbral y escenarios pendientes de definición.
- **Plazo:** pendiente de definición.
- **Fuente de evidencia:** evaluación de mantenimiento y recuperación con usuarios; aún no realizada.
- **Criterion status:** approved (D-005)

## Límites y exclusiones

- **Plataforma:** el producto se orienta exclusivamente a Codex. Las superficies concretas y los entornos compatibles de Codex se definirán posteriormente, sin afirmar compatibilidad aún no verificada.
- **Alcance funcional:** la meta incluye todas las capacidades de la referencia: captura y consulta, ingestión, organización, investigación, reflexión y decisiones, síntesis y reconciliación, visualización y exportación, integraciones, extensibilidad y mantenimiento automatizado. Estas familias son orientativas y no sustituyen el inventario completo ni excluyen capacidades no enumeradas aquí.
- **Referencia:** R-001 fija una base inicial identificable. Las novedades futuras del proyecto original requerirán evaluación; no se asume sincronización perpetua automática.
- **Fuera de alcance:** mantener versiones para Claude, Gemini u otros agentes. Se conservan formatos interoperables. D-007 descarta integraciones propias con proveedores externos.
- **Relación con Obsidian y proveedores:** D-007 prioriza resolver las tareas desde Codex sin integrar otras herramientas ni pedir claves de proveedores. Se conserva la meta de resultados funcionales, pero no la obligación de reproducir cada integración externa. Las equivalencias y brechas se deben demostrar. Obsidian permanece opcional para los archivos; el adaptador propuesto se reevaluará bajo esta restricción.
- **Alcance de este documento:** define dirección de producto. No decide lenguaje de implementación, organización de módulos, contratos, calendario de entregas ni mecanismos técnicos. Su aprobación no autoriza por sí sola implementación, instalación o publicación.
- **Estado y fuente:** approved; D-005 aprueba estos límites con las definiciones futuras indicadas expresamente como pendientes.

## Gobierno y revisión

- **Quién propone cambios:** el promotor del proyecto y colaboradores, incluidos asistentes que actúan por encargo.
- **Quién los aprueba:** el usuario promotor del proyecto, identificado en los metadatos; identidad nominal pendiente de confirmación.
- **Cuándo revisar:** cambios en las necesidades de los usuarios, restricciones de Codex o evidencia que cuestione objetivos, principios o límites; también propuestas de reducción del alcance funcional.
- **Regla de aprobación:** los cambios de misión, visión, principios, objetivos, criterios de éxito o límites requieren decisión explícita sobre la versión presentada. Las correcciones editoriales no amplían compromisos.
- **Regla de mantenimiento:** una funcionalidad nueva no exige por sí sola modificar esta constitución. Los conflictos deben exponerse antes de ejecutar trabajo afectado y conservar su resolución.
- **Regla de evidencia:** una capacidad propuesta, implementada y validada representa estados distintos. La aprobación documental no certifica resultados.

### Historial de decisiones

| Identificador | Fecha | Versión | Declaración o cambio y fundamento | Responsable y alcance |
|---|---|---|---|---|
| D-001 | 2026-10-08 | v001 | «Dale. Realicemos la […] del proyecto. Realmente me gustaría conservar todas las funcionalidades pero para Codex.» La referencia abreviada corresponde a la skill writing-constitution invocada en este chat. | Usuario promotor: encarga la constitución y confirma la intención de conservar funcionalidades para Codex; no aprueba el borrador aún no presentado. |
| D-002 | 2026-10-08 | v001 | «Usuarios de Codex en general», respuesta a la pregunta sobre usuario prioritario. | Usuario promotor: confirma el público objetivo, no la constitución completa. |
| D-003 | 2026-10-08 | v001 | Creación del borrador con fuentes identificadas y propuestas explícitas. No se encontraron documentos de producto en las carpetas actuales docs/, references/ y skills/, ni una constitución equivalente en la búsqueda acotada del repositorio. | Asistente por encargo; aprobación pendiente. No reemplaza ni atribuye vigencia a propuestas de sesiones anteriores ausentes del checkout. |
| D-004 | 2026-10-08 | v001 | «Si, aprobado.», respuesta a «¿la misión refleja lo que buscás, o ajustarías algo?» y a la misión citada íntegramente en este chat. Base presentada: `7fa8b76bfb36a26339c08ea4028f05c0b75b2df2:docs/superpowers/CONSTITUTION.md`, sección Misión. | Usuario promotor: aprobación de la formulación de la misión únicamente. Las restantes secciones siguen pendientes; no autoriza implementación ni publicación. |
| D-005 | 2026-10-08 | v001 | «Está todo aprobado.», declaración explícita en este chat sobre la constitución completa presentada y su registro de aprobación parcial. Base exacta: `995f242a891336a45c249d4273575e9611799d3d:docs/superpowers/CONSTITUTION.md`. | Usuario promotor del proyecto: aprueba misión, visión, usuarios, PRI-001–PRI-005, OBJ-001–OBJ-003, EXI-001–EXI-003, límites y gobierno. Los asuntos declarados pendientes se aceptan como abiertos, sin atribuirles una resolución ni mediciones inexistentes. No autoriza implementación, instalación o publicación. |

### Fuentes de contexto

- **R-001:** repositorio de referencia `eugeniughelbur/obsidian-second-brain`, copia local `.repos/obsidian-second-brain`, commit `b0089f7666d0c1e3db41c9d587b6f3c997665235`, versión declarada `0.17.0`. Fuentes inspeccionadas: `README.md`, `SKILL.md`, `.claude-plugin/plugin.json`, listado de `commands/` e `integrations/obsidian-mcp-server/README.md`. Describen capacidades de referencia, no prueban funcionamiento ni aprueban prioridades del nuevo producto. La copia permanece intacta.
- **R-002:** declaraciones del usuario en este chat del 2026-10-08, preservadas en D-001, D-002, D-004 y D-005. No se dispone de enlace recuperable a esas declaraciones; se conserva aquí su contenido y alcance.

## Asuntos pendientes

D-005 aprueba la constitución con los siguientes asuntos abiertos. Su resolución
requiere decisiones o evidencia posteriores; no modifica la dirección aprobada.

| Información o decisión pendiente | Cómo resolverla | Responsable |
|---|---|---|
| Equivalencia funcional solo con Codex | D-007 resuelve la dirección; evaluar capacidades disponibles, brechas y adaptación de las funciones originalmente ligadas a servicios externos | Usuario promotor y colaboradores |
| Medidas, umbrales y horizonte temporal de EXI-002 y EXI-003 | Obtener evidencia con usuarios y acordar metas; no se inventan valores | Usuario promotor y colaboradores |
| Condiciones de operación | Desktop y macOS inicial acordados en ARCHITECTURE.md; versiones/arquitecturas concretas y comprobaciones pendientes | Usuario promotor y colaboradores |

### D-006 — Confirmaciones y reconciliación de pendientes

El 2026-10-08, el usuario promotor anotó «Aprobado el nombre `llm-wiki`»
y «Si, aprobado.» sobre las recomendaciones 1 y 2 del mensaje
`msg_0a9075ab2a4808b2016ac7e44b7700819183e30367bef0e6ba`. Confirma
el nombre y R-001 como base de equivalencia; los cambios futuros del original
se evaluarán por separado. No confirma expresamente la identidad nominal.
La propuesta se presentó en el chat; no se inventa un commit de esa presentación.

Base aprobada anterior preservada:
`eaa8e3075c482dc3fd9cb9864b50b34e668d97a0:docs/superpowers/CONSTITUTION.md`.
Esta revisión registra las confirmaciones y reconcilia pendientes con las
decisiones de arquitectura; conserva la autoridad de D-005 sobre la dirección
sin cambios. No atribuye aprobación nueva a mediciones, código o publicación.

### D-007 — Resolver las tareas con Codex

El 2026-10-08 el usuario indicó: «Yo no integraría otras herramientas.
Quiero que todo lo resuelva Codex.» Se registra como aprobación de esa
dirección: sin adaptadores propios a proveedores externos ni claves de esos
servicios. Sustituye la expectativa de conservar cada integración como medio;
la meta de resultados completos se mantiene, con brechas explícitas hasta
verificar equivalencia. No aprueba una reducción funcional no identificada.
ARCHITECTURE.md, AC-019, registra las consecuencias técnicas y pendientes.

Base anterior preservada: `5d42830ea1b8ff5f844935e0277807a3db39e825:docs/superpowers/CONSTITUTION.md`.

## Actualización editorial de entrega — 2026-10-10

Referencia vigente v0.2.1; propósito y decisiones aprobadas sin cambios.
La dirección anterior queda preservada en
`5c2829f92b08d6756745361e49572e7530122300:docs/superpowers/CONSTITUTION.md`; esta actualización no amplía
la aprobación de D-005/D-006/D-007 a resultados de usuario todavía pendientes.

## Actualización editorial de entrega — 2026-10-12

Referencia vigente v0.2.2. El usuario solicitó el nombre visible **LLM Wiki**;
el identificador técnico sigue siendo `llm-wiki`. El registro D-006 conserva el
nombre aprobado originalmente como antecedente histórico. Propósito, principios
y dirección aprobada sin cambios; la corrección de siete problemas de plantillas
y la actualización del nombre no acreditan resultados de usuario pendientes.
