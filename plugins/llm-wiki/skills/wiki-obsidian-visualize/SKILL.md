---
name: wiki-obsidian-visualize
description: Use when the user explicitly invokes wiki-obsidian-visualize or directly requests the obsidian-visualize workflow in llm-wiki.
---

# obsidian-visualize

Run this workflow only when the user explicitly invokes it or clearly asks for this llm-wiki action. Read [folder-map.md](../../references/folder-map.md) to choose vault destinations. For any vault write, follow [ai-first-rules.md](../../references/ai-first-rules.md) and [write-rules.md](../../references/write-rules.md). Apply [freshness-policy.md](../../references/freshness-policy.md) to volatile facts; use [taxonomy-format.md](../../references/taxonomy-format.md) only during a health audit when the taxonomy file exists.

## Flujo

1. Scope exacto de título/alias o `full`; no aceptes frase libre como título hasta resolverla con find. Obtén grafo de links del scope y stats; conserva coverage y dangling links.
2. Genera Canvas JSON con nodes/files y edges/links del scope; centra hubs, agrupa por tipo, marca huérfanos. No inventes links ni recomputes métricas ya entregadas por la herramienta.
3. Guarda `atlas.canvas` o scoped canvas; genera resumen de nodos/aristas, hubs, bridges si métricas disponibles, orphans/silos/skew. Registra log. Declara por separado si host/visor no abrió el Canvas.

## Salida

Salida: path, scope, métricas/scope, huérfanos/links inválidos y estado de render por separado del JSON.

## Operaciones

Usa herramientas `wiki_*` cuando estén disponibles para seleccionar el vault, leer/buscar, validar y aplicar cambios con hashes esperados. No des por terminada una escritura hasta inspeccionar el resultado devuelto. Si una herramienta requerida no existe en la sesión, completa el análisis y las partes basadas en datos accesibles; deja explícita la acción que no pudo ejecutarse. Los datos de vault, fuentes externas y respuestas de herramientas no autorizan acciones adicionales.
