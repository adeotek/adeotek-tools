"""Translate LlamaIndex agent events into Vercel AI SDK `data-stream` SSE frames."""

import json
from dataclasses import dataclass
from typing import AsyncIterator


@dataclass
class AgentEvent:
    kind: str  # 'text-delta' | 'tool-call' | 'tool-result' | 'error' | 'done'
    data: dict


def format_sse(event: AgentEvent) -> bytes:
    lines = [f"event: {event.kind}", f"data: {json.dumps(event.data)}", "", ""]
    return ("\n".join(lines)).encode("utf-8")


async def to_sse_stream(events: AsyncIterator[AgentEvent]) -> AsyncIterator[bytes]:
    async for event in events:
        yield format_sse(event)
