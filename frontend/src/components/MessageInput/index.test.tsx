import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import { MessageInput } from "./index";

describe("MessageInput", () => {
  it("renders a textarea and send button", () => {
    render(<MessageInput onSend={vi.fn()} connectionStatus="connected" />);
    expect(screen.getByPlaceholderText("Type a message...")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Send" })).toBeInTheDocument();
  });

  it("calls onSend with trimmed text on form submit", async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();

    render(<MessageInput onSend={onSend} connectionStatus="connected" />);

    const textarea = screen.getByPlaceholderText("Type a message...");
    await user.type(textarea, "hello there");
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(onSend).toHaveBeenCalledWith("hello there");
  });

  it("clears input after sending", async () => {
    const user = userEvent.setup();

    render(<MessageInput onSend={vi.fn()} connectionStatus="connected" />);

    const textarea = screen.getByPlaceholderText("Type a message...");
    await user.type(textarea, "hello");
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(textarea).toHaveValue("");
  });

  it("disables input and button when disconnected", () => {
    render(<MessageInput onSend={vi.fn()} connectionStatus="disconnected" />);

    expect(screen.getByPlaceholderText("Reconnecting...")).toBeDisabled();
    expect(screen.getByRole("button", { name: "Send" })).toBeDisabled();
  });

  it("disables input and button when connecting", () => {
    render(<MessageInput onSend={vi.fn()} connectionStatus="connecting" />);

    expect(screen.getByPlaceholderText("Reconnecting...")).toBeDisabled();
    expect(screen.getByRole("button", { name: "Send" })).toBeDisabled();
  });

  it("does not send empty messages", async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();

    render(<MessageInput onSend={onSend} connectionStatus="connected" />);

    await user.click(screen.getByRole("button", { name: "Send" }));
    expect(onSend).not.toHaveBeenCalled();
  });

  it("sends on Enter key press", async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();

    render(<MessageInput onSend={onSend} connectionStatus="connected" />);

    const textarea = screen.getByPlaceholderText("Type a message...");
    await user.type(textarea, "hello{Enter}");

    expect(onSend).toHaveBeenCalledWith("hello");
  });
});
