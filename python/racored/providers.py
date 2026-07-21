from __future__ import annotations

import time
from dataclasses import dataclass, asdict
from typing import Any

import httpx

from .vault import CredentialVault


@dataclass(frozen=True)
class Provider:
    id: str
    name: str
    kind: str
    base_url: str
    default_model: str
    auth: str = "bearer"
    free: bool = False
    local: bool = False


PROVIDERS: dict[str, Provider] = {
    "openai": Provider("openai", "OpenAI", "responses", "https://api.openai.com/v1", "gpt-5.6-terra"),
    "anthropic": Provider("anthropic", "Anthropic", "anthropic", "https://api.anthropic.com/v1", "claude-sonnet-4-5", "x-api-key"),
    "gemini": Provider("gemini", "Google Gemini", "gemini", "https://generativelanguage.googleapis.com/v1beta", "gemini-2.5-flash", "query"),
    "openrouter": Provider("openrouter", "OpenRouter", "openai", "https://openrouter.ai/api/v1", "openrouter/auto"),
    "kimi": Provider("kimi", "Kimi / Moonshot", "openai", "https://api.moonshot.ai/v1", "kimi-k2.5"),
    "ollama": Provider("ollama", "Ollama", "ollama", "http://127.0.0.1:11434", "qwen3:8b", "none", True, True),
    "opencode": Provider("opencode", "OpenCode", "cli", "local://opencode", "configured", "local", True, True),
    "claude-code": Provider("claude-code", "Claude Code", "cli", "local://claude", "sonnet", "local", False, True),
    "kimi-code": Provider("kimi-code", "Kimi Code", "cli", "local://kimi", "default", "local", True, True),
}


class ProviderGateway:
    def __init__(self, vault: CredentialVault, timeout: float = 90):
        self.vault = vault
        self.client = httpx.AsyncClient(timeout=timeout, follow_redirects=True)

    def catalog(self) -> list[dict[str, Any]]:
        connected = set(self.vault.connected())
        return [{**asdict(provider), "connected": provider.local or provider.id in connected, "maskedKey": self.vault.masked(provider.id)} for provider in PROVIDERS.values()]

    async def health(self, provider_id: str) -> dict[str, Any]:
        provider = PROVIDERS[provider_id]
        started = time.perf_counter()
        try:
            if provider.kind == "ollama":
                response = await self.client.get(f"{provider.base_url}/api/tags", timeout=3)
                response.raise_for_status()
                models = [model.get("name") for model in response.json().get("models", [])]
                return {"ok": True, "latencyMs": round((time.perf_counter() - started) * 1000), "models": models}
            if provider.kind == "cli":
                import shutil

                executable = provider.base_url.removeprefix("local://")
                return {"ok": bool(shutil.which(executable)), "latencyMs": 0, "executable": executable}
            return {"ok": bool(self.vault.get(provider_id)), "latencyMs": 0, "configured": bool(self.vault.get(provider_id))}
        except Exception as exc:
            return {"ok": False, "latencyMs": round((time.perf_counter() - started) * 1000), "error": str(exc)}

    async def chat(self, provider_id: str, messages: list[dict[str, str]], model: str | None = None, system: str | None = None) -> dict[str, Any]:
        if provider_id not in PROVIDERS:
            raise ValueError(f"Unknown provider: {provider_id}")
        provider = PROVIDERS[provider_id]
        selected_model = model or provider.default_model
        started = time.perf_counter()
        if provider.kind == "cli":
            result = await self._cli(provider, messages)
        elif provider.kind == "ollama":
            result = await self._ollama(provider, selected_model, messages, system)
        elif provider.kind == "responses":
            result = await self._openai_responses(provider, selected_model, messages, system)
        elif provider.kind == "anthropic":
            result = await self._anthropic(provider, selected_model, messages, system)
        elif provider.kind == "gemini":
            result = await self._gemini(provider, selected_model, messages, system)
        else:
            result = await self._openai_compatible(provider, selected_model, messages, system)
        result.update({"provider": provider.id, "model": selected_model, "latencyMs": round((time.perf_counter() - started) * 1000)})
        return result

    def _key(self, provider: Provider) -> str:
        key = self.vault.get(provider.id)
        if not key:
            raise PermissionError(f"Connect {provider.name} in Racore before using it")
        return key

    async def _openai_responses(self, provider: Provider, model: str, messages: list[dict[str, str]], system: str | None) -> dict[str, Any]:
        payload: dict[str, Any] = {"model": model, "input": messages}
        if system:
            payload["instructions"] = system
        response = await self.client.post(f"{provider.base_url}/responses", headers={"Authorization": f"Bearer {self._key(provider)}"}, json=payload)
        response.raise_for_status()
        data = response.json()
        text = data.get("output_text") or "".join(part.get("text", "") for item in data.get("output", []) for part in item.get("content", []) if part.get("type") in {"output_text", "text"})
        return {"text": text, "usage": data.get("usage", {}), "rawId": data.get("id")}

    async def _openai_compatible(self, provider: Provider, model: str, messages: list[dict[str, str]], system: str | None) -> dict[str, Any]:
        payload_messages = ([{"role": "system", "content": system}] if system else []) + messages
        headers = {"Authorization": f"Bearer {self._key(provider)}", "Content-Type": "application/json", "X-Title": "Racore Browser", "HTTP-Referer": "https://racore.xyz"}
        response = await self.client.post(f"{provider.base_url}/chat/completions", headers=headers, json={"model": model, "messages": payload_messages})
        response.raise_for_status()
        data = response.json()
        return {"text": data["choices"][0]["message"].get("content", ""), "usage": data.get("usage", {}), "rawId": data.get("id")}

    async def _anthropic(self, provider: Provider, model: str, messages: list[dict[str, str]], system: str | None) -> dict[str, Any]:
        payload: dict[str, Any] = {"model": model, "max_tokens": 4096, "messages": messages}
        if system:
            payload["system"] = system
        response = await self.client.post(f"{provider.base_url}/messages", headers={"x-api-key": self._key(provider), "anthropic-version": "2023-06-01"}, json=payload)
        response.raise_for_status()
        data = response.json()
        return {"text": "".join(item.get("text", "") for item in data.get("content", []) if item.get("type") == "text"), "usage": data.get("usage", {}), "rawId": data.get("id")}

    async def _gemini(self, provider: Provider, model: str, messages: list[dict[str, str]], system: str | None) -> dict[str, Any]:
        contents = [{"role": "model" if m["role"] == "assistant" else "user", "parts": [{"text": m["content"]}]} for m in messages]
        payload: dict[str, Any] = {"contents": contents}
        if system:
            payload["systemInstruction"] = {"parts": [{"text": system}]}
        response = await self.client.post(f"{provider.base_url}/models/{model}:generateContent", params={"key": self._key(provider)}, json=payload)
        response.raise_for_status()
        data = response.json()
        text = "".join(part.get("text", "") for candidate in data.get("candidates", []) for part in candidate.get("content", {}).get("parts", []))
        return {"text": text, "usage": data.get("usageMetadata", {})}

    async def _ollama(self, provider: Provider, model: str, messages: list[dict[str, str]], system: str | None) -> dict[str, Any]:
        payload_messages = ([{"role": "system", "content": system}] if system else []) + messages
        response = await self.client.post(f"{provider.base_url}/api/chat", json={"model": model, "messages": payload_messages, "stream": False})
        response.raise_for_status()
        data = response.json()
        return {"text": data.get("message", {}).get("content", ""), "usage": {"prompt_tokens": data.get("prompt_eval_count"), "completion_tokens": data.get("eval_count")}}

    async def _cli(self, provider: Provider, messages: list[dict[str, str]]) -> dict[str, Any]:
        import asyncio

        prompt = messages[-1]["content"]
        executable = provider.base_url.removeprefix("local://")
        args = {"opencode": ["run", prompt], "claude": ["-p", prompt, "--output-format", "text"], "kimi": ["--print", prompt]}.get(executable, [prompt])
        process = await asyncio.create_subprocess_exec(executable, *args, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE)
        stdout, stderr = await asyncio.wait_for(process.communicate(), timeout=120)
        if process.returncode:
            raise RuntimeError(stderr.decode(errors="replace").strip() or f"{executable} failed")
        return {"text": stdout.decode(errors="replace").strip(), "usage": {}}
