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
