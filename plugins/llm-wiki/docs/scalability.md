# Escalabilidad del vault

## Decisión de la Fase 6

La arquitectura inicial sigue siendo file-first: Markdown y Git son la fuente de verdad. La
Fase 6 agrega medición y criterios de decisión, pero no agrega por anticipación una base de datos,
un vector store, un servicio externo ni un índice obligatorio.

El objetivo es detectar cuándo el enfoque index-first deja de ser práctico y elegir la mínima
capa derivada que resuelva el problema observado. Una capa derivada debe poder reconstruirse desde
el estado actual del vault y no debe convertirse en una segunda fuente de verdad.

## Qué se mide

El evaluador de escala (tests/assess-vault-scale.sh) es de solo lectura. Recibe la raíz de un
vault y reporta:

- cantidad de source records actuales bajo raw/sources/ (recorriendo namespaces y contando solo carpetas hoja);
- cantidad de páginas canónicas bajo wiki/pages/;
- cantidad de síntesis;
- archivos pendientes en raw/inbox/;
- bytes de originales, extracciones y assets;
- archivo más grande del layer de fuentes;
- bytes totales de Markdown bajo wiki/ y archivo Markdown más grande;
- bytes totales de síntesis y síntesis Markdown más grande;
- source records sin source.<ext> o sin extracted.md;
- source records con más de un source.<ext>, slugs inválidos o entradas inesperadas;
- rutas obligatorias ausentes, con tipo incorrecto o inaccesibles;
- disponibilidad, estado de cambios y profundidad básica de Git.

El reporte mide tamaño y señales estructurales. Una anomalía estructural produce `Scale status:
INVALID` y un diagnóstico estable; por ejemplo, `Invalid path type: raw/sources must be a
directory` o `Inaccessible directory: wiki/pages`. El evaluador no continúa contando el corpus
como si fuera válido cuando falta una ruta obligatoria, hay un tipo incorrecto o una estructura de
source record contradictoria.

No inventa una latencia de consulta ni considera que un corpus grande sea un problema por sí
mismo. La experiencia observada —tiempo de respuesta, cantidad de lecturas necesarias, cobertura
y errores de recuperación— puede elevar la prioridad aunque el contador de archivos todavía esté
por debajo de los umbrales.

Ejecutar:

~~~bash
bash tests/assess-vault-scale.sh <vault-root>
~~~

El comando no crea índices, no modifica archivos y no genera un reporte persistente. Si se
necesita conservar una medición, debe registrarse explícitamente en la documentación operativa o
en una síntesis, con fecha, commit y vault evaluado.

## Umbrales operativos iniciales

Los umbrales son un punto de partida revisable, no un límite técnico del modelo.

| Estado | Señal cuantitativa | Decisión |
|---|---|---|
| GREEN | Menos de 80 source records, menos de 200 páginas canónicas, ningún archivo de fuentes o Markdown de wiki de 5 MiB o más y menos de 50 MiB de Markdown de wiki total | Mantener recuperación targeted/index-first con Markdown y Git. |
| WATCH | Entre 80 y 100 source records, entre 200 y 300 páginas, algún archivo de fuentes o Markdown de wiki de 5 MiB o más, o entre 50 y 249 MiB de Markdown de wiki total, sin alcanzar DERIVED-SEARCH-CANDIDATE | Medir consultas reales, revisar cobertura y volver a evaluar antes de agregar infraestructura. |
| DERIVED-SEARCH-CANDIDATE | Más de 100 source records, más de 300 páginas canónicas o 250 MiB o más de Markdown de wiki total | Evaluar un índice full-text local regenerable; no activarlo automáticamente. |

El estado numérico se determina por la señal más exigente. Un vault con archivos faltantes,
extracciones incompletas o inbox pendiente no queda “listo” solo porque su tamaño sea GREEN;
esas condiciones se reportan por separado y afectan la preparación de contexto.

También se puede adelantar la evaluación de una capa derivada si se observan durante dos
mediciones comparables:

- búsquedas targeted que requieren escanear repetidamente gran parte del corpus;
- consultas current-corpus que no pueden producir cobertura en un tiempo operativo razonable;
- degradación reproducible de tiempos, consumo o precisión;
- fallas de navegación, límites de archivos del host o imposibilidad de cargar extracciones
  relevantes.

Estos son criterios operativos que deben acompañarse con evidencia del comando, del host y del
commit; no se deducen solo del número de documentos.

## Orden de adopción

Cuando exista evidencia suficiente, adoptar una sola extensión por vez y volver a ejecutar lint y
la evaluación de escala.

1. **Búsqueda full-text local derivada.** Primera opción para acelerar descubrimiento lexical.
   Debe reconstruirse desde raw/sources/<slug>/extracted.md, páginas y metadatos actuales;
   `<slug>` puede ser plano o jerárquico.
   Debe conservar path, slug, hash y commit de origen. No reemplaza las citas al vault.
2. **Caché de metadatos.** Útil si el costo dominante es recorrer el árbol, contar archivos o
   leer frontmatter. Su clave de invalidación debe incluir como mínimo el slug y los hashes de
   los archivos que indexa.
3. **Git LFS para binarios grandes.** Considerarlo para originales o assets grandes cuando el
   tamaño o los límites del host perjudiquen clone, fetch o revisión. No usarlo para esconder
   extracciones Markdown ni para crear copias por revisión.
4. **Embeddings opcionales.** Considerarlos solo si la búsqueda lexical no encuentra relaciones
   semánticas relevantes. Son una ayuda de descubrimiento, no autoridad: cada resultado debe
   conservar su path, slug, hash de contenido y versión del modelo, y poder regenerarse.
5. **Herramientas de migración.** Introducirlas únicamente para un cambio de esquema concreto.
   Deben ofrecer dry-run, reporte de cambios, preservación de Git y una aplicación explícita.

No se recomienda saltar directamente a embeddings, un grafo o un servicio remoto. Esas opciones
agregan invalidación, privacidad, costos y una superficie de fallo mayor antes de probar si un
índice local resuelve el cuello de botella.

## Invariantes de cualquier capa derivada

- raw/sources/<slug>/source.<ext>, extracted.md, las páginas canónicas y Git siguen siendo
  la autoridad.
- El índice o caché nunca recibe ediciones manuales como parte de la operación normal.
- Una reconstrucción completa desde el vault actual debe ser posible.
- Un resultado debe poder rastrearse hasta un slug, un path y un hash o commit.
- Los documentos pendientes, inaccesibles, parciales o contradictorios no se convierten en
  conocimiento confiable por el solo hecho de estar indexados.
- La capa derivada debe tener una política clara de invalidación y una prueba de consistencia con
  el árbol canónico.
- En Codex y ChatGPT Work se debe reportar si el host no expone el índice, el Git o el corpus
  completo; la instalación del plugin no resuelve esa limitación.

## Criterio de salida

La Fase 6 queda satisfecha cuando:

1. existe una medición reproducible y de solo lectura;
2. sus umbrales y señales operativas están documentados;
3. el contrato de pruebas ejecuta la medición sobre un vault válido;
4. no se agrega infraestructura derivada sin evidencia;
5. una futura extensión tiene una fuente de verdad, una estrategia de reconstrucción y una
   política de trazabilidad definidas.
