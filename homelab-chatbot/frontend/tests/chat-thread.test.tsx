import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { ChatThread } from "@/components/chat/ChatThread";
import type { DisplayMessage } from "@/lib/stream";

const sample: DisplayMessage[] = [
  { id: "1", role: "user", content: "hello" },
  { id: "2", role: "assistant", content: "hi there" },
];

describe("ChatThread", () => {
  it("renders all messages", () => {
    render(<ChatThread messages={sample} />);
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(screen.getByText("hi there")).toBeInTheDocument();
  });

  it("shows placeholder when empty", () => {
    render(<ChatThread messages={[]} />);
    expect(screen.getByText(/start a conversation/i)).toBeInTheDocument();
  });
});
