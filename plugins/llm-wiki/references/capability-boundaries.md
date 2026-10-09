# Límites de acceso externo

Usa capacidades que la sesión realmente expone. La investigación y síntesis web de Codex puede aportar páginas leídas con fecha, URL, cobertura, truncamiento y errores; un snippet no equivale a página completa. No uses conectores o proveedores externos ni pidas claves.

No hay ruta propia confirmada aquí para leer/escribir calendario externo, recibir una cola Telegram, buscar tendencias/respuestas completas de X, extraer o transcribir audio/video, descargar escenas ni leer comentarios de YouTube. Mantén esas modalidades separadas. Si el usuario aporta texto, eventos, transcript o frames, analízalos como material aportado e identifica su origen; eso no demuestra que se leyó el servicio o archivo original. Si se pide una fase que requiere el acceso ausente, indica `capability_unavailable` y la fase no hecha, pero completa con los datos que sí están disponibles.

La invocación es explícita. No hay hooks ni scheduling configurados por estos archivos. La síntesis web del host sí puede operar cuando la herramienta web de la sesión está disponible; no implica acceso a X.
