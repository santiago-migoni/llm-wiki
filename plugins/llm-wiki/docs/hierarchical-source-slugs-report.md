# Informe: soporte para slugs jerárquicos en fuentes

**Estado:** implementada en el árbol de trabajo; pendiente de revisión y commit.

## Problema

El plugin asume que todos los registros fuente tienen una sola carpeta:

~~~text
raw/sources/<slug>/source.<ext>
~~~

Por eso no soporta correctamente una estructura jerárquica como:

~~~text
raw/sources/fundamentos/actividades/source.md
raw/sources/fundamentos/actividades/extracted.md
~~~

Puntos afectados:

- Antes de este cambio, `inventory.py` solo inspeccionaba directorios inmediatos de `raw/sources/`.
- `hashes.py` solo reconocía un segmento de slug.
- `provenance.py` rechazaba `/` porque el slug coincidía únicamente con kebab-case plano.
- `validate.py` interpretaba `fundamentos/` como un registro sin `source.*` y generaba falsos errores.
- El resolvedor de wikilinks usaba principalmente el nombre base del archivo, lo que podía producir
  ambigüedades entre sub-slugs iguales.

La implementación agrega una enumeración común de carpetas hoja, slugs seguros con namespaces,
hashes actuales e históricos recursivos, provenance jerárquica, páginas canónicas anidadas y
resolución exacta antes del fallback por basename. Las fuentes planas no se migran.

## Solución propuesta

Añadir soporte para slugs jerárquicos conservando compatibilidad con los slugs planos actuales.

Estructura objetivo:

~~~text
raw/sources/
  fundamentos/
    actividades/
      source.md
      extracted.md
      assets/
    contactos/
      source.md
      extracted.md
      assets/
~~~

La identidad completa de cada fuente sería:

~~~text
fundamentos/actividades
fundamentos/contactos
~~~

## Requisitos técnicos

1. Aceptar slugs compuestos por segmentos kebab-case separados por `/`.
2. Considerar como registro fuente únicamente una carpeta hoja que contenga exactamente un
   `source.*` y un `extracted.md`.
3. Permitir carpetas intermedias como namespaces, por ejemplo `fundamentos/`.
4. Hacer recursivos el inventario, cálculo de hashes, búsqueda de duplicados actuales e históricos
   y validación.
5. Usar el slug completo en provenance:

   ~~~yaml
   slug: fundamentos/actividades
   source: raw/sources/fundamentos/actividades/source.md
   extracted: raw/sources/fundamentos/actividades/extracted.md
   ~~~

6. Mantener válidos los registros existentes como `security-policy`.
7. Rechazar segmentos vacíos, `.`, `..`, barras duplicadas, rutas absolutas y nombres que no sean
   kebab-case.
8. Actualizar documentación, contratos y mensajes de error para describir la estructura
   jerárquica.
9. Añadir pruebas para:
   - inventario de fuentes anidadas;
   - hashes actuales e históricos;
   - detección de duplicados entre namespaces;
   - validación de provenance;
   - compatibilidad con fuentes planas;
   - rutas inseguras;
   - dos sub-slugs iguales bajo namespaces diferentes.

## Páginas canónicas anidadas

Si también se desea organizar las páginas canónicas físicamente, debería admitirse:

~~~text
wiki/pages/fundamentos/actividades.md
~~~

En ese caso, el validador debe comparar el slug con la ruta relativa completa, y el resolvedor de
wikilinks debe priorizar coincidencias exactas como `[[fundamentos/actividades]]` antes de buscar
por nombre base.

## Compatibilidad y migración

El cambio no debe considerarse una migración automática: las fuentes planas existentes deben
seguir funcionando, y los registros jerárquicos nuevos deben validarse como `namespace/sub-slug`.

La implementación futura debe preservar los paths actuales, los hashes, la procedencia y el
historial Git. Antes de modificar fuentes existentes deberá presentar un plan explícito y una
validación de compatibilidad.
