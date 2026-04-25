import json

from app.llm.sse import AgentEvent, format_sse


def test_format_sse_contains_event_and_data():
    out = format_sse(AgentEvent("text-delta", {"text": "hi"}))
    decoded = out.decode("utf-8")
    assert "event: text-delta" in decoded
    assert "data:" in decoded
    payload_line = [line for line in decoded.splitlines() if line.startswith("data:")][0]
    assert json.loads(payload_line[len("data: "):]) == {"text": "hi"}


def test_done_event_serialization():
    out = format_sse(AgentEvent("done", {}))
    assert b"event: done" in out
    assert b"data: {}" in out
