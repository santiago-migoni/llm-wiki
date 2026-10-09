---
title: Empaquetado Codex Desktop para macOS
status: draft
version: v001
responsible: null
created_at: "2026-10-08"
updated_at: "2026-10-08"
reviewed_at: null
approval_reference: null
---

# Empaquetado Codex Desktop para macOS

## Recomendación

Empaquetar una distribución local Codex con un marketplace Git en el propio
repositorio, un plugin de producto llamado `llm-wiki` en
`plugins/llm-wiki/`, skills nativas en
`skills/<nombre>/SKILL.md` y un servidor MCP Go con transporte STDIO. Mantener
el manifiesto Agent Plugins portable (`plugin.json`, `mcp.json`) en la raíz del
plugin. Construir binarios Go macOS arm64 y amd64 y ensamblarlos durante el build
con `lipo` en `bin/llm-wiki`, que Codex invocará directamente. No usar launcher
de shell en tiempo de ejecución. Esta propuesta cubre el marketplace local/Git;
no afirma compatibilidad con el directorio universal público, cuyas condiciones
para MCP local deben revisarse por separado.

La correspondencia con la estructura actual es una dirección de empaquetado,
no evidencia de paridad funcional. El build `agent-skills` ya crea skills pero
deja fuera MCP e integraciones; tampoco copia `integrations/`, que necesita el
script `semantic_search.py`. Antes de incluir cada skill, hay que cerrar esas
dependencias y confirmar que la instrucción describe una capacidad ejecutable.
El inventario local identifica 12 herramientas MCP de referencia, aunque el
README de esa integración está desactualizado.

## Archivos propuestos

```text
.agents/plugins/marketplace.json
plugins/llm-wiki/
  plugin.json
  mcp.json
  skills/
    <skill-portada>/SKILL.md
    <skill-consulta>/SKILL.md
    ...
  bin/llm-wiki
  README.md
```

`plugin.json` contiene identidad estable, versión semántica, descripción,
autoría y la presentación `extensions.com.openai.interface`. `skills/` es una
ruta fija del paquete portable; no hace falta declararla en el manifiesto.
`mcp.json` es la configuración portable del servidor, y no el formato heredado
`.mcp.json`.

Marketplace mínimo para probar desde el propio checkout:

```json
{
  "name": "llm-wiki-local",
  "interface": { "displayName": "LLM Wiki local" },
  "plugins": [
    {
      "name": "llm-wiki",
      "source": { "source": "local", "path": "./plugins/llm-wiki" },
      "policy": { "installation": "AVAILABLE", "authentication": "ON_INSTALL" },
      "category": "Productivity"
    }
  ]
}
```

El marketplace local usa una ruta `source.path` como `./plugins/llm-wiki`,
resuelta desde la raíz del catálogo. Para obtener el catálogo por Git,
`codex plugin marketplace add <URL-GIT>` registra esa fuente. La entrada manual
para un plugin remoto en subdirectorio usa `source: "git-subdir"`, URL del
repositorio, `path: "./plugins/llm-wiki"` y un `ref` o `sha`; `source: "url"`
es para plugins en la raíz. Son campos distintos llamados `source`: el
`source.path` local apunta dentro del catálogo y la URL identifica el repositorio
remoto. Preferir SHA o tag de release y usar `main` solo en pruebas.

Configuración MCP ilustrativa, sujeta a validar en Codex Desktop la expansión
de `${PLUGIN_ROOT}` y su ejecución desde la caché del plugin:

```json
{
  "$schema": "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json",
  "mcpServers": {
    "llm-wiki": {
      "type": "stdio",
      "command": "${PLUGIN_ROOT}/bin/llm-wiki",
      "args": ["serve"],
      "cwd": "${PLUGIN_ROOT}"
    }
  }
}
```

La especificación MCP exige que stdout quede reservado para mensajes del
protocolo y que los logs vayan a stderr. `lipo` ensambla las arquitecturas
durante el build, no al iniciar el plugin. Comprobar `lipo -archs bin/llm-wiki`
y validar después con ambos tipos de Mac que Go, dependencias y llamadas del
runtime funcionen en el ejecutable universal. No introducir `go run`, Python,
npm ni descargas durante el arranque.

`OBSIDIAN_VAULT_PATH` pertenece al servidor de referencia legado; no es el
contrato nuevo. `llm-wiki` debe permitir seleccionar y persistir el vault por
proyecto mediante el selector aprobado, con la asociación en `os.UserConfigDir()/llm-wiki/config.json`
del proyecto. El marketplace y los archivos versionados no deben incluir rutas
personales. Verificar en Codex
Desktop cómo la selección llega al proceso MCP, sin asumir que hereda variables
del shell cuando la app se abrió desde Finder o el Dock.

## Cómo obtener y distribuir el binario

La decisión de implementación Go ya está aprobada en AC-005. El checkout
observado no tiene Go disponible en PATH (`go version` devolvió
`command not found`) ni contiene `go.mod`, `go.sum` o fuente Go propia en la
raíz. El servidor Python de `.repos/obsidian-second-brain` es solo referencia
funcional, no el runtime que se empaquetará. La implementación Go debe cubrir
las operaciones y propiedades de `llm-wiki`, incluidas seguridad de rutas,
bookkeeping, paginación, errores y salida de protocolo.

Producir builds reproducibles por arquitectura en CI o en una máquina con Go
instalado, con versión y flags fijados; combinarlos después con `lipo`. Esquema
orientativo de build (no ejecutado):

```sh
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -o /tmp/llm-wiki-arm64 ./cmd/llm-wiki
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -o /tmp/llm-wiki-amd64 ./cmd/llm-wiki
lipo -create /tmp/llm-wiki-arm64 /tmp/llm-wiki-amd64 -output plugins/llm-wiki/bin/llm-wiki
lipo -archs plugins/llm-wiki/bin/llm-wiki
```

La guía del marketplace describe cómo obtener archivos Git y cargar el plugin,
pero no un paso para descargar automáticamente activos de GitHub Releases ni
para ejecutar una instalación de dependencias. Por ello, si el plugin necesita
STDIO inmediatamente después de instalarse, ambos binarios deben estar dentro
del árbol del plugin que recibe Codex. Puede ser en el repositorio o dentro de
un artefacto local de distribución. No apuntar `mcp.json` a un release asset
que el host no descargó. Registrar checksums y SBOM del binario. Evaluar firma y
notarización de macOS para evitar fricción de Gatekeeper cuando el artefacto se
descargue desde Git.

## Instalar y probar sin publicar

La fuente Git/local se mantiene separada del directorio universal. Para probar
el catálogo local, el archivo de repo se ubica en
`.agents/plugins/marketplace.json`; Codex Desktop puede descubrirlo al abrir el
repositorio como proyecto y mostrar su marketplace en el navegador de plugins.
La CLI también documenta `codex plugin marketplace add` para fuente local o Git,
pero ese comando registra el marketplace en la configuración del usuario. No
se ejecutó porque esta investigación debía ser de solo lectura.

Flujo de validación manual recomendado para un futuro piloto autorizado:

1. Validar los JSON contra los esquemas Agent Plugins 1.0 y comprobar paths,
   permisos ejecutables y que no haya secretos ni rutas personales versionadas.
2. Desde Codex Desktop abierto en el checkout, reiniciar la app si el catálogo
   o el plugin cambió, instalar el plugin local y abrir una conversación nueva.
   Los cambios al plugin se cargan desde la copia instalada en caché, que para
   fuentes locales usa la versión `local`.
3. Comprobar que las skills aparecen y se activan con sus disparadores
   previstos; verificar descubrimiento de herramientas, arranque STDIO, una
   llamada de lectura inocua y stderr/stdout separados.
4. Con un vault temporal, ejercitar una escritura segura y revisar resultado,
   validación, registro e índice; detener y volver a iniciar el proceso.
5. Repetir en Apple Silicon e Intel y revisar logs del host ante fallos.
6. Ejecutar además T10 del plan: actualización y retorno a versión previa compatible,
   continuidad entre chats y recuperación de lote tras reinicio. Este smoke test
   de arranque no sustituye esas pruebas de aceptación.

En el entorno inspeccionado, `codex plugin --help` expone `add`, `list`,
`remove` y `marketplace`; `codex plugin marketplace --help` expone `add`,
`list`, `upgrade` y `remove`. No expone un subcomando `validate`. `list` es una
consulta del catálogo/configuración actual; `marketplace add`, `upgrade`,
`remove`, `plugin add` e instalar desde UI sí cambian el estado del usuario o
de la caché. El ejecutable de esta sesión fue `/opt/homebrew/bin/codex`; se
leyó `--help`, sin instalar ni modificar configuración.

Para validación estática, usar validación JSON y los esquemas oficiales como
base, más comprobaciones propias de rutas, frontmatter, archivos referenciados,
permisos y compatibilidad de arquitectura. El validador de envío público sirve
para una futura carga al directorio universal; no sustituye la prueba local de
un binario STDIO. `scripts/conformance_report.py` valida aspectos del repo y
builds actuales, no el contrato del plugin Codex ni el arranque MCP Go. La
validación funcional decisiva sucede en Codex Desktop instalado localmente.

## Riesgos y límites

- **Compatibilidad por superficie:** los catálogos local y Git están documentados
  para Codex Desktop, pero las superficies varían. No inferir que un servidor
  que ejecuta localmente estará disponible en web o móvil.
- **Compatibilidad del manifiesto:** Codex documenta el overlay
  `.codex-plugin/plugin.json` además del formato portable `plugin.json` y
  `mcp.json`. Empezar por el formato portable; añadir el overlay Codex solo si
  la versión objetivo no descubre o resuelve ese formato. Sincronizar nombre,
  versión, presentación y rutas, y volver a probar.
- **Binarios:** arquitectura equivocada, ausencia del ejecutable, permisos,
  cuarentena/notarización o un error de build `lipo` impedirán el inicio.
  Git marketplace por sí solo no compila Go.
- **Selección del vault:** `OBSIDIAN_VAULT_PATH` solo pertenece al servidor de
  referencia legado. Validar que el selector nuevo se persista por proyecto y
  entregue la ruta al proceso MCP sin incorporarla al paquete Git.
- **Equivalencia de implementación:** el servidor Python es evidencia de
  referencia, no destino de port. La implementación Go debe demostrar ranking,
  escritura, indexado y límites de seguridad con pruebas de contrato.
- **Skills:** el inventario reporta 46 skills construidas actualmente y excluye
  calendario. No anunciar paridad con las 47 funciones ni incluir skills que
  dependan de MCP/herramientas aún ausentes.
- **Configuración local:** registrar un marketplace o instalar un plugin sí
  modifica estado de usuario/caché. Mantener esa acción fuera de una inspección
  read-only y documentar claramente el cambio cuando llegue el piloto.

## Fuentes

- Skill local leído: `${CODEX_HOME}/plugins/cache/openai-curated-remote/plugin-creator/0.1.22/skills/create-plugin/SKILL.md`; referencias locales `.../references/local-plugins.md` y `.../references/plugin-packages.md`. Describen integración MCP local, estructura, catálogo e instalación.
- Documentación web oficial [Package your plugin](https://developers.openai.com/plugins/build/plugins): estructura portable, catálogo Git/local, rutas, instalación y pruebas.
- [Esquema MCP Agent Plugins 1.0](https://github.com/agentplugins/agent-plugins-spec/blob/main/schemas/1.0.0/mcp.schema.json): estructura y transporte del manifiesto.
- [Parser de plugins de Codex](https://github.com/openai/codex/blob/main/codex-rs/codex-mcp/src/agent_plugin_config.rs): prueba de que el parser Codex actual reconoce `stdio`, `command`, `args`, `env` y `cwd`; la versión instalada debe seguir verificándose.
- [Inventario Codex local](../inventario-codex/README.md), [Integraciones](../inventario-codex/INTEGRACIONES.md) y [Ciclo de vida](../inventario-codex/CICLO-DE-VIDA.md): evidencia de cobertura, dependencias faltantes y herramientas actuales del repo.

## Evidencia de esta inspección

Se leyó `create-plugin/SKILL.md` y sus referencias locales `local-plugins.md` y
`plugin-packages.md`. Se consultaron documentación oficial OpenAI, el esquema
Agent Plugins y el parser upstream de Codex. Se leyó la ayuda del CLI local. No
se instalaron plugins ni dependencias, no se escribió configuración de usuario,
no se ejecutó el servidor y no se compiló Go. `go version` no está disponible.
