---
title: Corrección y regresiones de las plantillas generadas
status: verified
created_at: "2026-10-12"
updated_at: "2026-10-12"
scope: siete hallazgos de generación y compatibilidad con salud
---

# Correcciones de plantillas — 2026-10-12

El usuario autorizó corregir los siete hallazgos de la auditoría, añadir
regresiones, verificar todos los presets y reconstruir el paquete, sin instalar
ni publicar. Trabajo local en `codex/template-regressions`, sin commit, push,
tag ni cambio de versión. Los registros de aceptación anteriores se conservan.

## Cambios comprobados

| Hallazgo | Corrección | Regresión |
| --- | --- | --- |
| Person bloquea áreas | Los placeholders YAML están entrecomillados en los assets, incluido last_interaction; se elimina la sustitución parcial por nombre de campo | Bootstrap real, parseo de todo el Markdown e inventario completo en los ocho escenarios |
| Tags en lista rechazados | Salud extrae secuencias con el parser YAML compartido y diagnostica elementos no textuales | Listas inline, bloque indentado y no indentado; comentarios; vacíos, null, números y booleanos; strings legítimos |
| Plantillas tratadas como notas finales | Política específica para Templates: conserva controles estructurales y errores YAML, sin exigir fecha resuelta ni producir análisis semánticos de notas | Fecha placeholder, enlaces pendientes, fuentes, duplicados, vigencia, errores YAML y cobertura; enlaces humanos a plantillas siguen resolviendo |
| Literature enlaza consigo misma | `{{source}}` identifica una fuente independiente en metadata y cuerpo | Título y fuente distintos; backlinks reales hacia la nota fuente |
| Detector reconoce solo Templater | Reconoce placeholders conocidos actuales y conserva sintaxis anterior | Cinco placeholders, legacy, líneas correctas; código inline/fenced y mustache desconocido legítimos |
| Post no incluye su type en tags | Añade post y conserva content | Metadata de Post rellenado sin avisos de taxonomía |
| Queue nace incompleta | Dashboard con type, tags, ai-first y contexto futuro | Ambos layouts researcher pasan salud |

Los ocho escenarios son `wiki` y `obsidian`, cada uno con `executive`, `builder`,
`creator` y `researcher`. Cubren las 60 plantillas Markdown, los 24 archivos
Bases y las dos colas. Las Bases se parsean como YAML y se comprueban sus
filtros/vistas; no se afirma ejecución de Obsidian o Dataview.

Las plantillas humanas existentes se preservan. Una regresión reemplaza
Person por contenido humano inválido, repite bootstrap y exige conservar
exactamente sus bytes, mantener el diagnóstico de salud y el inventario de
áreas incompleto. No se migraron bóvedas existentes.

## Pruebas y revisión

Se observaron fallos de regresión antes de las correcciones: Person en los
ocho escenarios, fuente de Literature, metadata de Post/Queue, listas YAML,
política de plantillas y detector de placeholders. El nuevo test de aceptación
MCP también falló con el paquete instalado anterior en los ocho escenarios por
`invalid_metadata` de Person; pasó con el paquete reconstruido.

Comprobaciones aprobadas con Go 1.27.2:

```sh
go test ./internal/vault -count=1
go test ./internal/health -count=1
go test ./internal/service -run TestBootstrapTemplatesAllPresetsThroughRuntime -count=1
go test -race ./...
go vet ./...
LLM_WIKI_PACKAGE=<paquete-nuevo>/plugins/llm-wiki go test ./tests/acceptance -run TestPackaged -count=1
LLM_WIKI_PACKAGE=<paquete-nuevo>/plugins/llm-wiki go test -race ./...
git diff --check
```

La suite final incluye el paquete como variable de entorno: no se omitieron
las dos pruebas condicionales `TestPackagedBinaryWithoutDevelopmentToolchain`
y `TestPackagedBootstrapAllPresets`. El ejecutable se prueba mediante MCP
STDIO real con registro y bóvedas sintéticos y sin toolchain en su PATH.
La prueba existente comprueba además relocalización, identidad del servidor,
descriptores de arranque y continuidad entre procesos.

Revisión independiente de código y distribución completadas. Se reforzaron
dos casos señalados durante la revisión: enlaces humanos hacia Templates y
tags YAML booleanos/números. La lectura lineal preexistente de otros campos
de metadata con comentarios inline sigue fuera de este alcance; no se afirma
compatibilidad exhaustiva del validador con todo YAML válido.

## Paquete local

Reconstruido con `go run ./cmd/package --output <salida-nueva>`, Go 1.27.2,
binario macOS universal `arm64` y `x86_64`, firma ad-hoc verificada para ambas
arquitecturas y 70 checksums verificados. Se conserva versión `0.2.1`, las 50
skills, atribución/licencia y descriptores portables. No se detectaron rutas
personales, selección de vault ni estado técnico empaquetado en la revisión.

Los artefactos verificados se copiaron a `plugins/llm-wiki/bin/llm-wiki`,
`plugins/llm-wiki/build.json` y `plugins/llm-wiki/checksums.json`; se compararon
sus bytes con el paquete nuevo antes y después de copiar.

- Base Git: `c84a40eef45178f57ccca3ff5a9e194b8bd51070`.
- `source_dirty: true`: las correcciones están en el árbol local, no en esa base.
- Huella de fuentes: `538850d2091a3af01beeee14031350564d9f7cc48943a5bf16c880458e747b0c`.
- SHA-256 del binario: `bc86900c1a33990b142a780ae85696c5785d2a459e62d9f7c8b258c690014fa8`.
- Dominio de la huella: `cmd/**/*.go`, `internal/**/*.go`,
  `internal/vault/bootstrap_templates.json`, `go.mod`, `go.sum`; SHA-256 de
  entradas ordenadas `path + NUL + file sha256 hex + LF`.

El generador de paquetes existente no emite todos los campos de procedencia.
Para este artefacto se enriqueció `build.json` después de compilar, se comprobó
la huella contra las fuentes y se regeneraron/verificaron los checksums. No se
atribuyó el binario modificado a una base Git limpia. Cualquier reconstrucción
futura requiere repetir ese registro o mejorar el generador dentro de un alcance
autorizado; no debe reutilizar esta huella después de modificar fuentes.

Pendientes deliberados: instalación, publicación, ejecución Intel y renderizado
Obsidian/Dataview. Apple Silicon y protocolo MCP del paquete sí se ejecutaron.
