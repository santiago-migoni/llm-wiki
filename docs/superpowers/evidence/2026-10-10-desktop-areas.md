# Intento de aceptación Desktop de llm-wiki 0.2.0

Fecha: 2026-10-10. Resultado: **bloqueado antes de las operaciones de vault**.
Este registro complementa [la evidencia técnica anterior](2026-10-10-areas.md);
no cambia sus resultados históricos ni acredita ejecución de workflows.

## Comprobaciones realizadas

- `codex plugin list --marketplace llm-wiki`: instalado y habilitado, versión
  `0.2.0`. La CLI consultada informa `0.135.0`; no se verificó la versión de
  la aplicación Desktop.
- Manifiestos del repositorio, marketplace instalado y cache coinciden en
  versión. Sus binarios tienen SHA-256
  `556600690bda3d338184d621a60f38875545692ad196b0a2f886899f6f1e53ce`.
- El manifiesto declara `mcpServers: ./.mcp.json`; el servidor declara
  `${PLUGIN_ROOT}/bin/llm-wiki serve`. La configuración de usuario habilita
  el plugin. La ausencia de un servidor standalone en esa configuración no
  demuestra un error: los plugins pueden suministrar servidores propios.
- Chat principal: búsqueda en `ALL_TOOLS`, cero herramientas del runtime
  llm-wiki expuestas.
- Segundo chat local independiente: creado con autorización explícita del
  usuario en el mismo proyecto, sin historial heredado. Confirmó cero
  herramientas `wiki_*` en dos búsquedas; leyó las skills instaladas del router,
  áreas y OKRs. Su turno final y salidas fueron recuperados desde el principal.
- El control de interfaz rechazó el acceso a la aplicación Codex por una
  restricción de seguridad. No se inspeccionó ni modificó su interfaz.

No se creó un vault sintético ni se ejecutaron mutaciones de vault. Ninguna
prueba de runtime por CLI, simulación o lectura de una skill se cuenta como
aceptación Desktop. La causa de la falta de herramientas sigue sin determinar;
no se atribuye al binario ni a un defecto específico del host.

## Cobertura de este intento

| Grupo | Resultado | Evidencia pendiente |
| --- | --- | --- |
| Instalación local 0.2.0 | Confirmada por inventario CLI y archivos | Conexión MCP operativa en Desktop |
| Áreas agnósticas y pertenencia múltiple | No ejecutado | Inventario completo, creación, enlaces en cuerpo, ausencia de duplicación |
| OKRs | No ejecutado | Creación/consulta/actualización, campos pendientes e historial preservado |
| Doce workflows modificados, con/sin áreas | No ejecutado | Herramientas invocadas, hashes actuales y diffs autorizados |
| Casos adicionales de la matriz | No ejecutado | Resultados por caso sobre vault sintético |
| Continuidad de notas entre dos chats | No ejecutado | Segundo chat recupera contenido persistido por el primero |

Los 38 casos de [la matriz](../../../tests/acceptance/areas-cases.json) conservan
`host_result: pending`. La creación y lectura de otro chat comprueba coordinación
del host, pero no persistencia ni recuperación del conocimiento del vault.

## Continuación reproducible

1. Restablecer la exposición de herramientas MCP en ambos chats y registrar
   el diagnóstico. La documentación oficial describe servidores suministrados
   por plugins y recomienda reiniciar la aplicación después de actualizar
   archivos de un plugin local; esto es una acción de diagnóstico pendiente,
   no una solución verificada.
2. Seleccionar explícitamente un vault sintético aislado mediante el runtime;
   cargar manual y Folder Map. No usar bóvedas personales.
3. En el primer chat, ejecutar los casos de áreas y OKRs y los doce workflows
   de la matriz en sus modos autorizados y sin áreas. Registrar llamadas,
   resultados, hashes y diffs por caso, incluyendo fallos y cobertura parcial.
4. En el segundo chat, seleccionar ese mismo vault y recuperar desde archivos
   las áreas, propósito, OKRs y la misma nota vinculada a dos áreas. Comprobar
   enlaces del cuerpo, identidad de la nota y ausencia de duplicación sin
   suministrar su contenido en el prompt.
5. Verificar invariantes de raw, metadata humana, Kanban, hashes obsoletos y
   calendario; actualizar únicamente los casos efectivamente ejecutados.

Fuentes de diagnóstico: [configuración MCP oficial](https://developers.openai.com/codex/mcp)
y [paquete e instalación local de plugins](https://developers.openai.com/plugins/build/plugins).
No se modificó configuración, instalación, runtime o paquete. No se repitieron
la suite Go, carreras ni vet: este intento sólo añade evidencia documental.

## Repetición después del reinicio

El usuario confirmó haber reiniciado Desktop y pidió continuar la prueba el
2026-10-10. Se reutilizaron los dos chats existentes; no se creó otro chat.

- Principal: búsqueda por nombre y descripción en `ALL_TOOLS`, cero herramientas
  llm-wiki. Segundo chat: nuevo turno autorizado, mismo resultado tras el reinicio.
- Inventario CLI: plugin instalado y habilitado `0.2.0`. Configuración de usuario
  habilitada; no se encontró una anulación del plugin en la configuración del
  proyecto. Esto no comprueba la configuración efectiva completa del host.
- Revisión dirigida de registros de arranque actuales: sin fallo de discovery
  identificado para llm-wiki. Los fallos encontrados para otro servidor no se
  atribuyen a este plugin ni se modifican dentro de esta prueba.
- Diagnóstico separado del binario instalado: proceso `serve --state-dir` con
  estado temporal aislado; `initialize` y `tools/list` respondieron sin error.
  Identidad `llm-wiki 0.2.0`, protocolo `2025-03-26`, **30 herramientas**,
  incluidas `wiki_list_areas`, `wiki_read_note`, `wiki_select_vault`,
  `wiki_save_note` y `wiki_update_note`. Proceso cerrado y estado temporal
  eliminado. No se invocó `tools/call` ni se seleccionó o modificó un vault.

El binario responde al MCP en el diagnóstico directo, mientras los dos chats
siguen sin recibir esas herramientas. Esto acota la investigación a la carga
o exposición del servidor en el host; no determina todavía su causa.
El reinicio no resolvió el síntoma observado. Se solicitó al usuario el estado
o error visible del servidor en Desktop porque el acceso automático a esa
interfaz está restringido. No se propone reinstalar, cambiar permisos ni
registrar un servidor alternativo sin conocer el diagnóstico del host.

Resultado conservado: **0/38 casos funcionales ejecutados** y continuidad de
vault sin verificar. El intercambio de mensajes entre chats sí fue comprobado,
pero no equivale a recuperar conocimiento persistido. El JSON de entrega
registra ambas comprobaciones de disponibilidad por separado.

## Causa identificada y corrección instalada

La investigación posterior utilizó el CLI incluido en Desktop, versión
`0.162.0-alpha.17.2`, en la aplicación `26.1007.21159` (build `20052`).
Es distinto del CLI de desarrollo `0.135.0` consultado inicialmente.
Se abrió un proceso diagnóstico del app-server oficial; no se enviaron mensajes
de prueba mediante él ni se lo contó como uno de los dos chats.

Antes de corregir, `plugin/installed` reconocía llm-wiki habilitado, pero
`mcpServerStatus/list` lo omitía. El stderr del cargador confirmó que rechazaba
`mcp.json` con el error:

> Agent Plugins stdio command must be a bare executable name or a contained `./` path

El manifiesto portable usaba `${PLUGIN_ROOT}/bin/llm-wiki` en `command`, una
forma incompatible con ese contrato. La presencia del manifiesto compatible
no evitaba el rechazo del portable. La fuente primaria del
[normalizador portable](https://github.com/openai/codex/blob/main/codex-rs/codex-mcp/src/agent_plugin_config.rs)
confirma el requisito. La prueba anterior arrancaba por ruta absoluta y no
ejercitaba este contrato; validar sólo `mcp.json` tampoco protegía `.mcp.json`.

Corrección: ambos descriptores usan `command: "./bin/llm-wiki"`,
`args: ["serve"]` y `cwd: "./"`. El validador revisa ambos. Las nuevas
regresiones rechazan el placeholder, escapes de cwd, transporte y argumentos
incompatibles, y aceptan la forma relativa legítima. La aceptación del paquete
lo copia a una raíz con espacios y consume sus descriptores literalmente.

La revisión independiente detectó que esa adaptación podía ejecutar siempre
el paquete actual durante downgrade. La comprobación de identidad del servidor
reprodujo el fallo (0.2.0 donde debía ser 0.1.0). Se corrigió con raíz explícita
por conexión; cada alternancia verifica nombre y versión reales. El paquete
legado se lanza por su ejecutable conocido sólo para comprobar compatibilidad
de runtime/estado; no se acredita su antiguo manifiesto como compatible.

Verificación: pruebas enfocadas, `go test -race ./...`, `go vet ./...` y
`git diff --check` aprobados. Tras corregir la prueba de alternancia también
pasaron carreras/vet de aceptación y la prueba empaquetada con ambos paquetes.
Firma ad hoc verificada y arquitecturas arm64/x86_64 presentes; Intel no se ejecutó.

El paquete final se reconstruyó desde fuentes comprometidas en
`d985fa177274aac764f187a623add06e0cc4a0d2`. Fingerprint del dominio Go declarado
en `build.json`: `8b745a1d3ad0dd090d281dbecec24f97824b20da8271a31c940b7d616f0c0811`.
Binario SHA-256: `f964722942b97154ba9e12441211e69c93296998226261e6cb47ca6e8f5ddd8c`.
Checksums, manifiestos y procedencia comprobados antes de copiarlo al repositorio
y al origen registrado. Reinstalado con el CLI oficial; no se editó el cache a
mano ni se modificaron vaults, permisos u otros plugins.

Después de reinstalar, el mismo cargador anuncia **llm-wiki y 30 herramientas**
y no emite el error de parseo. Este antes/después verifica la corrección de carga
del servidor. La instalación conserva versión 0.2.0 como corrección local:
no se movió el tag histórico v0.2.0 ni se publicó una nueva entrega.

Los dos chats existentes todavía devolvieron cero herramientas en su inventario.
La recarga por el socket de control del daemon administrado no estaba disponible
(socket inexistente). No se reinició la aplicación ni se rodeó la restricción
de control de interfaz. Se pidió reiniciar después de instalar la corrección
para volver a comprobar las sesiones. **La carga está corregida; las 38 pruebas
de workflows y la continuidad de vault entre chats siguen sin ejecutar.**

## Aceptación real después del último reinicio

Los estados anteriores son históricos. Ambos chats existentes exponen ahora
30 herramientas MCP. El escritor seleccionó una bóveda sintética temporal;
todas las mutaciones de contenido se hicieron mediante MCP recuperable.

Se ejecutaron los **24 casos pareados de los 12 workflows**, primero sin áreas
y después con Exploración/Comunidad. Lecturas completas, hashes, operaciones,
contenido anterior/posterior y decisiones sin escritura están en
[observaciones](../../../tests/acceptance/desktop-areas-observed.json).
La [matriz](../../../tests/acceptance/areas-cases.json) registra cada resultado.

- `projects` sólo escribió Last overview. `project` y `save` reutilizaron los
  enlaces ya presentes, leyendo el contenido y hash actuales sin otra escritura.
- Persona, captura, reunión local, daily, log, revisión, síntesis y destilación
  conservaron destinos canónicos. La tarea corta permaneció como tarjeta de
  Backlog, sin nota propia ni columnas nuevas.
- Frontmatter y párrafo humanos preservados; ambas fuentes raw conservaron sus
  hashes. Los nuevos enlaces quedaron en notas derivadas y cuerpos de notas.
- La revisión declara seis días ausentes. La síntesis usa dos contextos y no
  ejecuta concept-evolution sin tres actualizaciones. El claim destilado cita B2.
- El OKR es draft: periodo, owner, KR, métrica, baseline, target, actual y fecha
  pendientes. Se añadió evidencia e historial sin calcular progreso ni inferir
  periodo a partir de la fecha de captura.

También se registraron **14 escenarios de guardas**. Un frontmatter malformado
provocó inventario incomplete/error y no se dedujo ausencia; los homónimos se
presentaron sin selección; un escape fue rechazado. Se conservaron la primera
área al añadir otra y un enlace por área al repetir la petición. El caso de
herramientas ausentes usa la ausencia real observada antes del reinicio.
Para C# se propuso C-sharp.md y después se ejecutó el fixture con ese destino
explícito: el grafo resolvió C-sharp.md y no C.md. No se ejercitó un diálogo nuevo
de aprobación humana. El C#.md heredado se diagnosticó sin migración automática.

El segundo chat recibió sólo ubicaciones y recuperó propósitos, relaciones,
pendientes y hashes con MCP. La primera pasada coincidió con escrituras y no
representa un snapshot único. La segunda pasada, con el vault principal estable,
coincide con los hashes del escritor: dos áreas completas, cobertura de 20 archivos
sin errores/truncamiento y **12 backlinks por área**, sin ambigüedades. El lector
no escribió. Véase [continuidad](../../../tests/acceptance/desktop-continuity-observed.json).

### Defecto adicional del catálogo MCP

Los arrays changes/cases/scope se publicaban sin items. Desktop los presentaba
como listas de texto, mientras el runtime esperaba objetos para los primeros dos.
Strings fallaron antes de escribir. Objetos reales sí funcionaron en los lotes
de esta aceptación y en retrieval-eval: recall@10=1, MRR=1/3, informe no escrito.

La corrección declara objetos para changes/cases y strings para relevant/scope,
con binario base64 y destino nullable. La regresión de tools/list falló antes de
corregir y luego aprobó. Las pruebas STDIO cubren creación, actualización, bytes
binarios, rechazo atómico de hash obsoleto, métricas conocidas y rechazo de strings.
Pruebas enfocadas, suite con carreras y vet aprobados; revisión independiente sin
hallazgos concretos. La aceptación del paquete aprobó manifiestos y alternancia
con 0.1.0. La corrección no cambia la lógica de los workflows.

La aceptación funcional corresponde al binario de lanzamiento documentado arriba.
La reconstrucción e instalación del esquema adicional se registran por separado
para conservar su procedencia. Scope: workflows declarativos interpretados por el
agente, reunión calendar local y fixtures sintéticos. No acredita renderizado
Obsidian, eventos remotos, ejecución Intel, todas las ramas semánticas ni resultados
de usuario EXI-002/003. El tag histórico y publicación no cambian.

### Paquete del esquema reconstruido e instalado

Fuentes comprometidas: `3955c6c105929bf1a67e0a31cf5873d569a1525d`.
Fingerprint del dominio Go: `13832d760578c1b7e03ae3c469969f80966946d2227e30d62d1212036ef0fe85`.
Binario: `65ebe9b3b29476fd99d895e1ecccc51b4dac0cce5e867d117e40ca8028e91136`.
Los 70 checksums coinciden entre salida nueva, repositorio, origen registrado
y cache instalado. Firma ad hoc verificada, arm64/x86_64 presentes. La aceptación
del paquete final volvió a aprobar arranque por manifiesto y alternancia 0.1.0.
Reinstalación mediante remove/add del CLI oficial, sin edición manual del cache.

Un proceso nuevo del cargador oficial anuncia 30 herramientas y los items
correctos: changes/cases objeto, scope string. Esto es diagnóstico del cargador,
no una tercera conversación de aceptación. El segundo chat real aceptó un lote
con objeto en dry-run, confirmó que el destino no existe y evaluó casos con
objetos (recall@10=1, MRR=1/3; sin informe persistido).

**Límite de recarga pendiente:** las sesiones existentes aún describen changes
y cases como Array<string> en ALL_TOOLS. La ejecución con objetos funciona, pero
la declaración expuesta al modelo conserva el cache anterior. Un reinicio posterior
permitirá volver a verificar esa declaración; no se afirma que ya se haya refrescado.
Las pruebas funcionales de áreas, OKRs, workflows y continuidad sí están cerradas
en su alcance explícito. Versión local 0.2.0, sin mover tags ni publicar/push.
