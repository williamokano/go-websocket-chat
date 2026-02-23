import { useEffect, useRef } from "react";
import type { ServerMessage } from "../../types";
import { MessageItem } from "../MessageItem";

interface MessageListProps {
  messages: ServerMessage[];
}

export function MessageList({ messages }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  return (
    <div className="flex-1 overflow-y-auto">
      {messages.length === 0 && (
        <div className="flex items-center justify-center h-full text-gray-400 text-sm">
          No messages yet. Say hello!
        </div>
      )}
      {messages.map((msg, i) => (
        <MessageItem key={msg.type === "chat_message" ? msg.id : i} message={msg} />
      ))}
      <div ref={bottomRef} />
    </div>
  );
}
