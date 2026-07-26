# Racore local AI

Racore uses `MadeAgents/Hammer2.0-0.5b` as its bundled, local-only agent
planner. The shipped artifact is a Q8_0 GGUF conversion, executed by a pinned
CPU build of `llama.cpp`.

## Why this model

- 0.5 billion parameters and a 531 MB quantized artifact.
- Fine-tuned for function calling and on-device agentic applications.
- Built on Qwen 2.5, with a compact footprint suitable for background use.
- CC-BY-4.0 licensed, unlike Hammer 2.1's non-commercial license.

At this size the model is intentionally a planner and tool router, not an
autonomous browser runtime. Racore validates its structured output and owns all
navigation, permission, safety, and side-effect execution.

## Reproducible download

Run:

```sh
npm run local-ai:prepare
```

The script downloads the pinned GGUF from Hugging Face and the pinned
platform-specific `llama.cpp` runtime, then verifies both with SHA-256. Large
runtime artifacts stay under the ignored `desktop/runtime/` directory and are
copied into the desktop bundle during packaging.

Model and runtime versions, source revisions, licenses, sizes, and checksums are
locked in `config/local-ai.json`.

## Attribution

`MadeAgents/Hammer2.0-0.5b` is provided by MadeAgents under CC-BY-4.0 and is
derived from the Qwen 2.5 model family. The GGUF artifact was converted by the
Hugging Face user Nekuromento. `llama.cpp` is provided by ggml-org under MIT.
