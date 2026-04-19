import type { Message } from "@/lib/api";

export type StreamEvent =
  | { kind: "conversation"; id: string }
  | { kind: "text-delta"; text: string }
  | { kind: "tool-call"; name: string; args: unknown }
  | { kind: "tool-result"; name: string; summary: string }
  | { kind: "error"; message: string }
  | { kind: "done" };

export async function* parseSseStream(
  response: Response,
): AsyncGenerator<StreamEvent> {
  if (!response.body) return;
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const frames = buffer.split("\n\n");
    buffer = frames.pop() ?? "";
    for (const frame of frames) {
      const ev = parseFrame(frame);
      if (ev) yield ev;
    }
  }
}

function parseFrame(frame: string): StreamEvent | null {
  const lines = frame.split("\n");
  let event: string | null = null;
  let data: string | null = null;
  for (const line of lines) {
    if (line.startsWith("event:")) event = line.slice(6).trim();
    else if (line.startsWith("data:")) data = line.slice(5).trim();
  }
  if (!event || data === null) return null;
  try {
    const parsed = data === "" ? {} : JSON.parse(data);
    return { kind: event as StreamEvent["kind"], ...parsed } as StreamEvent;
  } catch {
    return null;
  }
}

export interface DisplayMessage {
  id: string;
  role: Message["role"];
  content: string;
  toolEvents?: Array<{ kind: string; name: string; summary?: string }>;
  partial?: boolean;
}
