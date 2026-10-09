---
date: 2026-10-08
type: research
ai-first: true
tags: [ensayo]
---

## For future agent
Ensayo sintético de llm-wiki, sin datos de usuarios. El alcance y las fuentes se declaran abajo.

# Transporte STDIO de MCP

## Summary
La especificación MCP 2025-06-18 describe STDIO mediante un servidor lanzado como subproceso por el cliente. La comunicación usa JSON-RPC por stdin/stdout; los diagnósticos pueden ir por stderr y stdout se reserva al protocolo. Fuente consultada el 2026-10-08, sección stdio.

## Facts
- Versión de la fuente: 2025-06-18, no se afirma que sea la más reciente.
- Autor individual: no identificado; proyecto Model Context Protocol.

## Timeline
Edición de referencia 2025-06-18. Lectura del ensayo 2026-10-08.

## Actors
Cliente y servidor MCP.

## Dissent
No evaluado: solo una fuente primaria.

## Further reading / Sources
[Especificación de transportes](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports#stdio).

## Open questions
Esta fuente no verifica la expansión de PLUGIN_ROOT en Desktop ni la aceptación del paquete; hacen falta pruebas del host.

## Coverage
Página primaria accesible mediante web nativa de Codex; se usó la sección stdio. Sin proveedores propios ni claves. No se consultaron notas privadas ni servicios adicionales.
