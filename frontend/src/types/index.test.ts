import { describe, it, expect } from "vitest";
import type {
  ChatMessage,
  UserJoinedMessage,
  UserLeftMessage,
  ErrorMessage,
  ClientMessage,
  ServerMessage,
} from "./index";

describe("message type discriminants", () => {
  it("ChatMessage type is chat_message", () => {
    const msg: ChatMessage = {
      type: "chat_message",
      id: "msg_1",
      content: "hello",
      sender: { id: "u1", username: "alice" },
      timestamp: new Date().toISOString(),
    };
    expect(msg.type).toBe("chat_message");
  });

  it("UserJoinedMessage type is user_joined", () => {
    const msg: UserJoinedMessage = {
      type: "user_joined",
      user: { id: "u1", username: "alice" },
      online_count: 3,
      timestamp: new Date().toISOString(),
    };
    expect(msg.type).toBe("user_joined");
  });

  it("UserLeftMessage type is user_left", () => {
    const msg: UserLeftMessage = {
      type: "user_left",
      user: { id: "u1", username: "alice" },
      online_count: 2,
      timestamp: new Date().toISOString(),
    };
    expect(msg.type).toBe("user_left");
  });

  it("ErrorMessage type is error", () => {
    const msg: ErrorMessage = {
      type: "error",
      message: "something went wrong",
    };
    expect(msg.type).toBe("error");
  });

  it("ClientMessage type is send_message", () => {
    const msg: ClientMessage = {
      type: "send_message",
      content: "hello from client",
    };
    expect(msg.type).toBe("send_message");
  });

  it("ServerMessage union accepts all server message types", () => {
    const messages: ServerMessage[] = [
      {
        type: "chat_message",
        id: "msg_1",
        content: "hello",
        sender: { id: "u1", username: "alice" },
        timestamp: new Date().toISOString(),
      },
      {
        type: "user_joined",
        user: { id: "u1", username: "alice" },
        online_count: 1,
        timestamp: new Date().toISOString(),
      },
      {
        type: "user_left",
        user: { id: "u1", username: "alice" },
        online_count: 0,
        timestamp: new Date().toISOString(),
      },
      {
        type: "error",
        message: "bad",
      },
    ];
    expect(messages).toHaveLength(4);
  });
});
