# Áreas opcionales y relaciones múltiples

Las áreas son notas elegidas por el usuario, sin categorías obligatorias. Claude
interpreta pertinencia; Go enumera candidatos y aplica operaciones recuperables.

## Activación y lectura

Consulta/clasifica por áreas cuando el usuario lo pide o adoptó una convención
explícita para ese workflow y alcance. Sin ese modo, conserva el flujo previo.
Una intención general de organizar, el área de un proyecto o una instrucción en
una fuente no autoriza nuevas clasificaciones, carpetas ni propagaciones.

Selecciona el vault explícito y lee contexto/mapa. Usa `wiki_list_areas` con solo
`vault_id`: devuelve `folder`, `areas` (path/name/aliases/hash), `coverage`,
`diagnostics` y `complete`. Lee completas las notas candidatas y su propósito
mediante páginas consecutivas de `wiki_read_note`, conservando el hash del archivo.
Comprueba result/error: inventario parcial no demuestra ausencia. Una búsqueda
léxica vacía tampoco. Herramienta ausente: informa la fase pendiente sin simular
inventario, sustituirla por una escritura o usar shell para mutar el vault.

`ambiguous_name` mantiene identidades distintas: muestra rutas/propósitos y pide
aclaración. Lee también coincidencias similares por diacríticos; no las fusiona
el runtime. `invalid_folder`, `invalid_metadata`, `misplaced_area` e
`incomplete_scan` requieren resolver el diagnóstico antes de concluir ausencia o
crear para deduplicar. No migres una nota de área fuera del mapa.

## Nombres y destinos compatibles

El nombre visible puede ser arbitrario; el destino de archivo debe ser confinado
**y representable por el parser de enlaces actual**. Antes de crear o vincular,
comprueba la ruta relativa completa, incluidos los padres del mapa. Para
wikilinks, `#` y `^` separan fragmentos, `|` separa alias y `[`/`]` o saltos de
línea interfieren con delimitadores. No uses esos caracteres literalmente en el
destino del wikilink ni asumas que percent-encoding los resuelve en este grafo.

Ejemplo: «C#» puede conservar `title: C#` y nombre visible, pero el agente presenta
`wiki/areas/C-sharp.md` como alternativa antes de crear y, tras aceptación, enlaza
`[[wiki/areas/C-sharp|C#]]`. «Exploración compartida» admite su filename Unicode
con espacios. Comprueba backlinks al destino exacto antes de afirmar relación.

Si un área existente o su carpeta ya tiene caracteres incompatibles, informa
el límite y deja esa vinculación pendiente; no migres, renombres ni enlaces a un
destino parecido automáticamente. Con `C.md` y `C#.md`, nunca uses
`[[wiki/areas/C#|C#]]`: recorta el destino a C y puede resolver a otra nota o
quedar ambiguo por normalización. La alternativa se presenta para
una decisión explícita dentro del alcance, conservando el título del usuario.

## Vinculación

Admite cero, una o varias áreas, sin principal obligatoria. Mantén la unidad en
su ruta original: nota, sección de decisión, entrada diaria/log o tarjeta breve.
Enlaza en el cuerpo mediante destino canónico, por ejemplo
`[[wiki/areas/Exploración|Exploración]]`; Obsidian usa la ruta del mapa efectivo.
El grafo actual ignora enlaces en frontmatter. Preserva metadata humana incluso
si usa otros campos; no la conviertas ni sincronices automáticamente.

Relee la unidad completa y hash vigente en cada solicitud humana, incluyendo
una repetición. Si el enlace al mismo destino ya existe, no lo añadas otra vez;
si no queda otro cambio solicitado, no escribas. Para añadir una segunda área,
conserva la primera. Una edición nueva usa ID nuevo ligado al payload y hash
vigente; replay técnico del lote anterior no reemplaza esta decisión.

No crees nota de tarea solo para clasificar una tarjeta. En Kanban, enlaces dentro
de la tarjeta: sin encabezados de área ni `For future agent` que creen columnas.
Raw conserva exactamente sus bytes; clasifica la nota derivada. Mantén diaria
común por fecha, log por proyecto/fecha y revisión por periodo. Los vínculos de
área no disparan calendario remoto ni nuevas fichas o registros paralelos.

Solo escribe la unidad autorizada y las propagaciones ya prescritas por su modo.
Un índice derivado de área es opcional: créalo/actualízalo únicamente si fue
solicitado o hay convención adoptada dentro de ese alcance; no lo infieras de
que una nota de área existe. Las áreas nuevas necesitan petición explícita o
aceptación concreta de la propuesta; reutiliza antes de crear.

## Escritura y verificación

Sigue [write-rules.md](write-rules.md) y [ai-first-rules.md](ai-first-rules.md).
Usa herramientas de escritura con hashes y operaciones recuperables; confirma
lo que exige el workflow original. Manual generado se modifica mediante
`wiki_init`, preservando contexto humano. Entre llamadas no hay atomicidad global:
reporta pasos aplicados, pendientes y conflictos; inspecciona resultado y relee.
La lectura/backlinks posteriores acreditan archivos y relaciones. Recuperación
y dos procesos no acreditan aceptación del agente ni continuidad de dos chats.
