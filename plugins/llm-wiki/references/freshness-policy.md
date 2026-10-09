# Vigencia de hechos

Cada hecho guardado debe ser atemporal, fechado o un puntero a dónde vive la verdad. Un conteo, estado, balance o fecha próxima no puede quedar como presente eterno: guarda el dato como `(as of YYYY-MM-DD, fuente)` o enlaza el sistema dueño; también puedes conservarlo dentro de un snapshot fechado.

Al revisar una nota, trata estas señales así:

- **FRESH-1:** afirmación cuantitativa/volátil en presente sin fecha o puntero: corregir con fuente/fecha o convertirla en puntero.
- **FRESH-2:** sello anterior a la ventana de frescura (7 días por defecto o la definida por carpeta): re-observar, convertir a puntero o mover a historia fechada.
- **FRESH-3:** puntero sin destino resoluble: conservar el hallazgo e indicar qué target falta.
- **FRESH-4:** snapshot fechado es historia; no reescribir su contenido como si fuera estado actual.
- **FRESH-5:** directiva `freshness: example` sin hallazgo que suprimir o duplicada: reportar como limpieza menor.

No actualices una cifra vieja sin volver a consultar su fuente. Si no tienes acceso, conserva el valor fechado como último dato observado y di que no fue refrescado. La vigencia cambia rankings y acciones ofrecidas; nunca justifica borrar historia.
