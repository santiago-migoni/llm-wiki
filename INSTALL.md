# Install LLM Wiki in Codex

This repository is a Codex marketplace with one plugin: `llm-wiki`.

## From GitHub

Register the repository and install the plugin:

~~~bash
codex plugin marketplace add santiago-migoni/llm-wiki --ref main
codex plugin add llm-wiki@llm-wiki
~~~

## From a local checkout

Register the repository path instead:

~~~bash
codex plugin marketplace add /absolute/path/to/llm-wiki
codex plugin add llm-wiki@llm-wiki
~~~

After installing or updating the plugin, start a new Codex task so it loads the current skills.
The plugin does not upload or persist vault documents. In ChatGPT Work, expose the vault through
the workspace, a project, a connected source, or files supplied in the conversation.
