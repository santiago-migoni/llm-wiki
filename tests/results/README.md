# Evidencia de pruebas

Este directorio contiene evidencia verificable de la auditoría y de las
pruebas del plugin. La matriz de escenarios y el formato operativo de cada
registro están documentados en
[`plugins/llm-wiki/docs/testing.md`](../../plugins/llm-wiki/docs/testing.md).

## Línea base de la remediación

La línea base se clasifica por tipo de evidencia y no por la existencia de una
respuesta del host:

| Área | Estado inicial | Evidencia esperada |
| --- | --- | --- |
| Empaquetado estático | Validado | `C-001`, `C-002` y el contrato del plugin |
| Comportamiento en Codex | Pendiente de evidencia de host | Resultados bajo `tests/results/codex/`, incluidos `H-001` y los escenarios aplicables |
| Comportamiento en ChatGPT Work | Pendiente de evidencia de host | Resultados bajo `tests/results/chatgpt-work/`, incluidos `H-002` y los escenarios aplicables |
| Estabilidad semántica | Pendiente de evidencia funcional | Resultados `G-*`, `Q-*` y `L-*`, con artefactos inspeccionables |

Un estado `validado` significa que existe una comprobación reproducible. Los
estados `pendiente` no deben presentarse como capacidades demostradas.

## Invariantes del vault

Cada resultado debe evaluar, cuando corresponda, estos invariantes:

1. Existe un único esquema operativo y una única raíz del vault.
2. Cada registro de fuente tiene una única copia actual `source.*` y un único
   `extracted.md` derivado.
3. Los hashes declarados coinciden con el original actual; las discrepancias
   quedan visibles y no se corrigen silenciosamente.
4. Toda página canónica está indexada y es alcanzable desde el índice del
   vault.
5. Toda afirmación que combine fuentes conserva la trazabilidad a cada fuente
   y a su marcador o ubicación verificable.
6. Toda contradicción material permanece visible en la página afectada y en el
   registro de auditoría.
7. Cada commit operativo contiene únicamente las rutas autorizadas y los
   artefactos que la operación declara haber modificado.

Si un escenario no puede evaluar un invariante, el registro debe marcarlo como
`blocked` y explicar qué evidencia falta.

## Registro obligatorio

Cada archivo de resultado debe incluir, como mínimo:

- commit exacto del plugin bajo prueba;
- host y versión o entorno relevante;
- fecha y hora con zona horaria;
- ubicación del vault, redactada si contiene información sensible;
- escenarios ejecutados y sus pasos relevantes;
- estado por escenario: `passed`, `failed` o `blocked`;
- rutas relativas y hashes de los artefactos generados, cuando aplique;
- limitaciones, observaciones y cualquier desviación del procedimiento.

Los artefactos deben permitir que otra persona inspeccione la afirmación sin
depender de una transcripción manual de la respuesta del modelo. No se deben
guardar secretos, tokens, credenciales, datos personales ni copias completas
de fuentes privadas. Usar redacciones explícitas y rutas relativas.

## Estados de evidencia

- `implemented`: existe código o documentación para la capacidad.
- `automated`: una prueba reproducible comprueba la capacidad.
- `host-validated`: la prueba se ejecutó en el host indicado y sus artefactos
  fueron inspeccionados.
- `blocked`: la prueba no pudo completarse; el registro debe indicar la causa.

Una capacidad solo puede describirse como demostrada en un host cuando tiene
evidencia `host-validated`. Un resultado aislado no reemplaza la matriz de
promesas públicas ni modifica el contrato por sí solo.

## Organización

- `codex/`: resultados de pruebas ejecutadas en Codex.
- `chatgpt-work/`: resultados de pruebas ejecutadas en ChatGPT Work.

Los nombres de archivo deben ser estables y descriptivos, por ejemplo
`YYYY-MM-DD_<host>_<scenario-set>.md`. Si un resultado se repite para otro
commit, se crea un archivo nuevo; no se sobrescribe evidencia histórica.
