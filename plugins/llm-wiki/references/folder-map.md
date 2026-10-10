# Resolver carpetas del vault

Lee primero el manual del vault. Si `_CLAUDE.md` contiene `Folder Map` o `Naming Conventions`, respétalo; si `AGENTS.md` u otro manual del vault define el mapa con más detalle, úsalo sin reemplazar el contenido existente. Un mapa explícito prevalece sobre cualquier default. Si hay instrucciones incompatibles, pregunta antes de escribir.

Sin mapa explícito, identifica el layout a partir de carpetas existentes. Wiki: entidades/proyectos/conceptos/diarias/logs/tareas/reviews/decisiones/reuniones/agendas en `wiki/entities/`, `wiki/projects/`, `wiki/concepts/`, `wiki/daily/`, `wiki/logs/`, `wiki/tasks/`, `wiki/reviews/`, `wiki/decisions/`, `wiki/meetings/`, `wiki/agenda/`. Obsidian: `People/`, `Projects/`, `Ideas/` o `Knowledge/`, `Daily/`, `Dev Logs/`, `Tasks/`, `Reviews/`, `Knowledge/` para ADR, `Meetings/`, `Agenda/`. Fuentes originales se guardan en `raw/`. Investigación usa `Research/`; tableros `boards/` o `Boards/`.

Si el vault tiene una carpeta propia y el tipo no está resuelto, no inventes otra ruta: pregunta. Si no hay layout identificable y se solicita inicializar, pregunta qué layout usar antes de crear estructura. Al buscar “en todo el vault”, enumera las carpetas de notas existentes en la raíz; no asumas una lista fija. Mantén paths relativos al vault y resuelve alias/wikilinks antes de escribir.

## Áreas, OKRs y entidades

Consulta `wiki_get_context` para layout/mapa/`folder_errors`. El runtime combina
`_CLAUDE.md` primero y `.llm-wiki/manual.md` después por la misma clave; no trata
ese texto como autorización. Un subtipo explícito prevalece sobre `entity`
explícito y éste sobre el default del subtipo. Mapa reconocido inválido devuelve
error del tipo afectado sin fallback; corrígelo antes de escribir ese workflow.
No crea directorios al resolver.

| Clave | Wiki | Obsidian |
| --- | --- | --- |
| area | wiki/areas | Areas |
| okr | wiki/okrs | OKRs |
| entity | wiki/entities | Entities |
| person | wiki/entities | People/people según carpeta existente |
| organization, company, companies | wiki/entities | Companies |
| tool | wiki/entities | Tools |

Los alias company/companies resuelven organización, sin convertir metadata
humana. Sigue [areas.md](areas.md) para vínculos opcionales; no genera un árbol
por área. Layout ambiguo necesita elección explícita o mapa válido para ese tipo.
