import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { MessageItem } from "./index";
import type { ChatMessage, UserJoinedMessage, UserLeftMessage, ErrorMessage } from "../../types";

describe("MessageItem", () => {
  it("renders a chat message with sender name and content", () => {
    const msg: ChatMessage = {
      type: "chat_message",
      id: "1",
      content: "Hello world",
      sender: { id: "u1", username: "alice" },
      timestamp: "2026-01-01T12:00:00Z",
    };

    render(<MessageItem message={msg} />);
    expect(screen.getByText("alice")).toBeInTheDocument();
    expect(screen.getByText("Hello world")).toBeInTheDocument();
  });

  it("renders a user_joined message with join notification", () => {
    const msg: UserJoinedMessage = {
      type: "user_joined",
      user: { id: "u2", username: "bob" },
      online_count: 2,
      timestamp: "2026-01-01T12:00:00Z",
    };

    render(<MessageItem message={msg} />);
    expect(screen.getByText(/bob joined/)).toBeInTheDocument();
    expect(screen.getByText(/2 online/)).toBeInTheDocument();
  });

  it("renders a user_left message with leave notification", () => {
    const msg: UserLeftMessage = {
      type: "user_left",
      user: { id: "u2", username: "bob" },
      online_count: 1,
      timestamp: "2026-01-01T12:00:00Z",
    };

    render(<MessageItem message={msg} />);
    expect(screen.getByText(/bob left/)).toBeInTheDocument();
    expect(screen.getByText(/1 online/)).toBeInTheDocument();
  });

  it("renders an error message", () => {
    const msg: ErrorMessage = {
      type: "error",
      message: "Something went wrong",
    };

    render(<MessageItem message={msg} />);
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });
});
