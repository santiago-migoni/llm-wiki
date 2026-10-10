# Validación Desktop de llm-wiki

Estado vigente: aceptación sintética en dos chats reales sobre la corrección local
0.2.0, con 30 herramientas MCP por chat, 24 casos de workflows y 14 guardas.
[Informe y límites](../../docs/superpowers/evidence/2026-10-10-desktop-areas.md).
El paquete 0.2.1 incorpora las correcciones; su instalación Desktop aún no se repitió.
Permanece pendiente la recarga del esquema de arrays de las sesiones existentes.
No se acredita interfaz visual, todos los permisos/modos ni ejecución Intel.

## Registro histórico 0.1.0 (sin cambios)

# Validación Desktop pendiente — llm-wiki 0.1.0

Fuente del binario: `43ee802`. Paquete versionado en `619fecb`; macOS 27.0.1 arm64, CLI 0.135.0. Versión Desktop no capturada. No se declara soporte Intel por generar un binario universal.

## Evidencia ejecutada

- [x] Instalación y actualización mediante `codex plugin add llm-wiki@llm-wiki`, con marketplace local registrado.
- [x] `plugin/read` del host CLI: dev.2 instalado/habilitado, 48 skills y servidor MCP configurado.
- [x] Paquete aislado: handshake, selección, guardar/leer y continuidad tras reiniciar proceso sin toolchain en PATH.
- [x] Alternancia de binarios dev.1/dev.2 conserva todos los hashes del vault/estado/workflows.
- [x] Fallo parcial y segundo proceso: resume correcto y rollback rechaza edición humana posterior.

Estos checks no sustituyen las siguientes pruebas de interfaz, chats y permisos.

## Casos Desktop por ejecutar

Usar exclusivamente un vault sintético, por ejemplo una copia de `tests/fixtures/wiki` en una carpeta de ensayo. No usar notas personales para probar fallos. Mantener la zona horaria explícita.

- [ ] Registrar versión Desktop, macOS y arquitectura. Confirmar llm-wiki en Complementos → Instalados, versión dev.2.
- [ ] En una sesión Desktop comprobar 48 skills y herramientas `wiki_*` realmente expuestas. Llamar `wiki_select_vault` con proyecto/vault sintéticos; registrar ID retornado.
- [ ] Llamar `wiki_read_note` sobre una nota conocida. Registrar resultado y ausencia de descargas/toolchain; no deducir expansión de PLUGIN_ROOT solo del manifiesto.
- [ ] Guardar una decisión de ensayo; en otro chat seleccionar el mismo proyecto, cargar contexto y recuperar/citar esa nota sin historial anterior.
- [ ] Conservar una edición humana entre chats; update con hash anterior debe rechazarla, sin sobrescribir bytes.
- [ ] Actualizar y volver a un paquete compatible mediante flujo marketplace soportado; comparar hashes. El cambio directo de binario probado no cierra este caso.
- [ ] Ensayar proyecto sin trust y rechazo de permisos en entorno aislado. Rechazo debe dejar notas sin cambios y mostrar diagnóstico; no simular ejecución exitosa.
- [ ] Comprobar evento/hook únicamente si existe y está expresamente habilitado. Ausencia no acredita automatización; no crear horarios reales para cerrar un checkbox.
- [ ] Ejecutar en Intel antes de declararlo soportado; anotar plataforma realmente probada.

## Restricción observada

El control visual devolvió: "Computer Use is not allowed to use the app 'com.openai.codex' for safety reasons." No se evadió con otras APIs o capturas. Se solicitó al usuario confirmar si el plugin aparece instalado; respuesta pendiente. La sesión actual tampoco expone las herramientas wiki del paquete, por lo que no se ha ejecutado una llamada Desktop desde este chat.

## Modalidades y release

Resultados por criterio en `results.json`; pruebas nativas parciales en `capabilities.json`; reporte del paquete en `package-report.json`. T6/T8 conservan pruebas por modalidad pendientes. O-001–009 no se cierran por un paquete instalable. No se publicó remoto ni se afirma paridad completa.
