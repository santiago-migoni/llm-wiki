# Políticas de escritura compartidas

Aplica la política indicada para el workflow: escritura prescrita automáticamente, oferta de guardar, petición explícita, aprobación por lote o dry-run. No agregues una confirmación universal a cada escritura pedida; tampoco conviertas una oferta, un análisis en chat o un dry-run en una mutación.

Antes de una escritura coordinada, presenta alcance y diferencias cuando el workflow lo requiere. Usa solo operaciones `wiki_*` propuestas por la arquitectura del paquete. `wiki_save_note` es create-only; para actualizar usa el hash esperado; reemplazos exigen ancla única; cambios múltiples usan `wiki_apply_changes` con precondiciones. Reporta archivos aplicados, pendientes y errores parciales. Si falta el backend, detén la mutación y marca la capacidad no disponible; nunca simules una llamada exitosa.

Las fuentes ya presentes en `raw/` son inmutables; nuevas fuentes se agregan sin sobrescribir. Preserva secciones humanas, historia, enlaces, Unicode y procedencia. Conflicto, coincidencia ambigua, destino existente o hash desactualizado detienen la operación afectada. No inventes entidades, fechas, credenciales ni datos personales.
