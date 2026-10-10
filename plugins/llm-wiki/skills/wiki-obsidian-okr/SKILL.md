---
name: wiki-obsidian-okr
description: Use when the user explicitly asks to create, consult, or update an OKR in llm-wiki, or invokes wiki-obsidian-okr.
---

# OKRs del vault

Ejecuta solo crear/consultar/actualizar según petición. Sigue
[folder-map.md](../../references/folder-map.md), [areas.md](../../references/areas.md),
[ai-first-rules.md](../../references/ai-first-rules.md), [write-rules.md](../../references/write-rules.md)
y [freshness-policy.md](../../references/freshness-policy.md).

1. Selecciona vault explícito, lee contexto y resuelve `okr`. Busca objetivo por
   título/aliases y lee completas las coincidencias; aclara homónimos. Consulta
   sola devuelve nota, fuentes/periodo y pendientes sin escribir.
2. Para crear o actualizar, conserva objetivo/resultados aportados con su fuente.
   Por resultado clave registra métrica, línea base, meta, responsable, periodo y
   evidencia solo cuando consten. Ausentes: `pendiente`, nunca cifras, porcentajes,
   fechas o responsables inferidos. Fecha de captura no es periodo del objetivo.
3. Una nota nueva lleva `type: okr`, fecha de captura, tags, `ai-first: true`, título,
   `## For future agent`, objetivo, resultados clave, periodo y procedencia. Si
   faltan resultados medibles, declara «borrador; todavía no permite medir progreso».
   Preserva resultados e historial existentes al actualizar.
4. Relaciona áreas solo por solicitud o convención adoptada; consulta el lector y
   propósito y añade enlaces canónicos en el cuerpo sin duplicar/mover la nota.
   No infieras áreas de proyectos. Nota OKR no implica tarea, tablero ni evento.
5. Tablero/índice y vínculos asociados se escriben solo si están solicitados o una
   convención adoptada los prescribe en ese alcance. No agendas calendario por
   inferencia. Usa creación exclusiva o update/lote con hash vigente, ID ligado al
   payload, y recuperabilidad existentes; manual generado usa `wiki_init`.
6. Inspecciona resultados y relee; reporta cambios aplicados, pendientes, fuentes
   y conflictos. Si falta herramienta, deja explícita la etapa no ejecutada.

Salida: ruta/link del OKR, objetivo/resultados sustentados, periodo y campos
pendientes, áreas elegidas y escrituras confirmadas. No presenta el borrador como
OKR cuantificado ni estima progreso sin valores y evidencia.
