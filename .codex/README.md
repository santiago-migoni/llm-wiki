# Configuración de desarrollo con Codex

Este directorio configura el trabajo en el repositorio. La guía común es
[`AGENTS.md`](../AGENTS.md), con el nombre que Codex descubre automáticamente.

`config.toml` habilita subagentes y limita a tres los hilos delegados simultáneos,
además del agente principal. Los cuatro roles se declaran en `[agents.<rol>]`
mediante `description` y `config_file`, con rutas relativas a este directorio.
Sus capas TOML en `agents/` se seleccionan por tarea; no se ejecutan todos
automáticamente:

| Rol | Responsabilidad | Permisos configurados |
| --- | --- | --- |
| `implementer` | Implementación y pruebas de archivos asignados | Hereda del principal |
| `code_reviewer` | Corrección y regresiones | Solo lectura |
| `security_reviewer` | Fronteras de confianza y recursos | Solo lectura |
| `release_reviewer` | Paquete, procedencia y privacidad | Solo lectura |

Modelos, autenticación, confianza del proyecto y permisos del principal pertenecen
al entorno de cada colaborador. Los controles efectivos del host prevalecen sobre
estos valores; configurar un rol no concede una herramienta ausente.
No se activan MCPs, hooks, instalaciones ni tareas programadas desde estos archivos.

Ejemplo de trabajo:

> Implementa el cambio solicitado. Usa implementer para el módulo asignado y,
> al completar el parche, pide a code_reviewer una revisión independiente.
> Usa security_reviewer si cambian entradas, rutas, escrituras o presupuestos.
> Usa release_reviewer antes de presentar un paquete distribuible. Integra los
> resultados y comunica pruebas, hallazgos corregidos y límites.

## Comprobación y carga

Codex carga la configuración del proyecto solo cuando este es de confianza.
Abre una sesión nueva después de editar las instrucciones o los roles.
Se usa `max_threads`, alias compatible, y declaraciones explícitas de roles
para soportar el CLI 0.135.0 comprobado. Las capas de rol contienen instrucciones
y, para revisores, sandbox de solo lectura.

Para comprobar TOML sin realizar llamadas al modelo:

```sh
python3 -c 'import pathlib,tomllib; [tomllib.loads(p.read_text()) for p in pathlib.Path(".codex").rglob("*.toml")]'
```

Este comando de desarrollo requiere Python 3.11 o posterior; no añade una
dependencia al plugin. Para comprobar la carga de instrucciones sin iniciar
una tarea del modelo, usa `codex debug prompt-input "Comprueba instrucciones"`.
Su salida puede contener contexto local: no la versiones ni la publiques.
`--strict-config` no está disponible para ese subcomando en el CLI comprobado.
No uses `codex exec` como simple validador: inicia una tarea del agente.

Fuentes oficiales consultadas:

- [AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
- [Subagentes y agentes personalizados](https://learn.chatgpt.com/docs/agent-configuration/subagents)
- [Referencia de configuración](https://learn.chatgpt.com/docs/config-file/config-reference)
