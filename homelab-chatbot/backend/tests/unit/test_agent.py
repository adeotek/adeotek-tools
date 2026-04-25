from collections.abc import AsyncIterator

from llama_index.core.base.llms.types import ChatMessage, ChatResponse

from app.llm.agent import run_agent
from app.llm.sse import AgentEvent


class FakeLLM:
    """Minimal stand-in for LlamaIndex LLM's astream_chat."""

    metadata = None

    async def astream_chat(self, messages) -> AsyncIterator[ChatResponse]:
        text = "Hello world"
        acc = ""
        for ch in text:
            acc += ch
            yield ChatResponse(
                message=ChatMessage(role="assistant", content=acc),
                delta=ch,
                raw={},
            )


async def test_run_agent_without_tools_yields_text_deltas():
    events: list[AgentEvent] = []
    async for event in run_agent(
        llm=FakeLLM(),
        tools=[],
        history=[],
        user_message="hi",
        tools_enabled=False,
    ):
        events.append(event)
    kinds = [e.kind for e in events]
    assert "text-delta" in kinds
    assert kinds[-1] == "done"
    joined = "".join(e.data.get("text", "") for e in events if e.kind == "text-delta")
    assert joined == "Hello world"
