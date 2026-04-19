import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { SWRConfig } from "swr";

import { ConversationSidebar } from "@/components/sidebar/ConversationSidebar";

describe("ConversationSidebar", () => {
  beforeEach(() => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => [
        {
          id: "a",
          title: "First",
          provider: "anthropic",
          model: "x",
          created_at: "2026-04-16T00:00:00",
          updated_at: "2026-04-16T00:00:00",
        },
      ],
    });
  });

  it("renders conversation titles", async () => {
    render(
      <SWRConfig value={{ provider: () => new Map() }}>
        <ConversationSidebar activeId={null} onSelect={() => {}} onNew={() => {}} />
      </SWRConfig>,
    );
    expect(await screen.findByText("First")).toBeInTheDocument();
  });

  it("renders New chat button", () => {
    render(
      <SWRConfig value={{ provider: () => new Map() }}>
        <ConversationSidebar activeId={null} onSelect={() => {}} onNew={() => {}} />
      </SWRConfig>,
    );
    expect(screen.getByRole("button", { name: /new chat/i })).toBeInTheDocument();
  });
});
