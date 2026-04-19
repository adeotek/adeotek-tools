"""Build a LlamaIndex agent and stream events out as AgentEvents."""

import inspect
import logging
from typing import AsyncIterator

from llama_index.core.agent.workflow import AgentStream, ReActAgent, ToolCallResult
from llama_index.core.llms import LLM, ChatMessage
from llama_index.core.tools import FunctionTool

from app.llm.sse import AgentEvent

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = (
    "You are a helpful assistant for a home lab knowledge base. "
    "You have two tools available:\n"
    "- search_homelab_docs: for prose/conceptual questions about the docs\n"
    "- query_homelab_inventory: for structured questions about devices, services, and inventory\n\n"
    "Call tools when they help answer the user's question. Answer concisely. "
    "If the tools return no useful information, say so."
)


def build_messages(history: list[tuple[str, str]], user_message: str) -> list[ChatMessage]:
    messages: list[ChatMessage] = [ChatMessage(role="system", content=SYSTEM_PROMPT)]
    for role, content in history:
        messages.append(ChatMessage(role=role, content=content))
    messages.append(ChatMessage(role="user", content=user_message))
    return messages


async def run_agent(
    llm: LLM,
    tools: list[FunctionTool],
    history: list[tuple[str, str]],
    user_message: str,
    tools_enabled: bool,
) -> AsyncIterator[AgentEvent]:
    """Stream AgentEvents while the agent produces a response."""
    try:
        if not tools_enabled or not tools:
            async for delta in _stream_chat(llm, history, user_message):
                yield AgentEvent("text-delta", {"text": delta})
            yield AgentEvent("done", {})
            return

        chat_history = [ChatMessage(role="system", content=SYSTEM_PROMPT)]
        for role, content in history:
            chat_history.append(ChatMessage(role=role, content=content))
        agent = ReActAgent(tools=tools, llm=llm, verbose=False)
        handler = agent.run(user_msg=user_message, chat_history=chat_history)
        async for event in handler.stream_events():
            if isinstance(event, AgentStream) and event.delta:
                yield AgentEvent("text-delta", {"text": event.delta})
            elif isinstance(event, ToolCallResult):
                yield AgentEvent(
                    "tool-result",
                    {
                        "name": event.tool_name,
                        "summary": str(event.tool_output.content)[:500]
                        if event.tool_output.content
                        else "",
                    },
                )
        await handler  # ensure workflow completes
        yield AgentEvent("done", {})
    except Exception as e:  # noqa: BLE001
        logger.exception("agent error")
        yield AgentEvent("error", {"message": str(e)})


async def _stream_chat(
    llm: LLM, history: list[tuple[str, str]], user_message: str
) -> AsyncIterator[str]:
    messages = build_messages(history, user_message)
    result = llm.astream_chat(messages)
    # LlamaIndex LLMs return a coroutine from astream_chat; test fakes return async generators directly
    if inspect.isawaitable(result):
        result = await result
    async for chunk in result:
        delta = chunk.delta if hasattr(chunk, "delta") else ""
        if delta:
            yield delta
