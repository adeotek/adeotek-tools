import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SWRConfig } from "swr";

import { ProviderPicker } from "@/components/settings/ProviderPicker";

describe("ProviderPicker", () => {
  beforeEach(() => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        default_provider: "anthropic",
        default_model: "claude-sonnet-4-6",
        providers: [
          {
            id: "anthropic",
            available: true,
            models: ["claude-sonnet-4-6", "claude-haiku-4-5"],
            tool_capable: ["claude-sonnet-4-6"],
          },
          {
            id: "ollama",
            available: true,
            models: ["llama3.1:8b"],
            tool_capable: ["llama3.1:8b"],
          },
        ],
      }),
    });
  });

  it("changes model when provider changes", async () => {
    const changes: string[][] = [];
    function Wrapper() {
      return (
        <SWRConfig value={{ provider: () => new Map() }}>
          <ProviderPicker
            provider="anthropic"
            model="claude-sonnet-4-6"
            onChange={(p, m) => changes.push([p, m])}
          />
        </SWRConfig>
      );
    }
    render(<Wrapper />);
    await waitFor(() => screen.getByLabelText(/provider/i));
    await userEvent.selectOptions(screen.getByLabelText(/provider/i), "ollama");
    expect(changes[0]).toEqual(["ollama", "llama3.1:8b"]);
  });
});
