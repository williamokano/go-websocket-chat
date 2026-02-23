import type { ServerMessage } from "../../types";

interface MessageItemProps {
  message: ServerMessage;
}

function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function MessageItem({ message }: MessageItemProps) {
  if (message.type === "user_joined") {
    return (
      <div className="py-1 px-4 text-center">
        <span className="text-xs text-gray-400">
          {message.user.username} joined &middot; {formatTime(message.timestamp)}{" "}
          &middot; {message.online_count} online
        </span>
      </div>
    );
  }

  if (message.type === "user_left") {
    return (
      <div className="py-1 px-4 text-center">
        <span className="text-xs text-gray-400">
          {message.user.username} left &middot; {formatTime(message.timestamp)}{" "}
          &middot; {message.online_count} online
        </span>
      </div>
    );
  }

  if (message.type === "error") {
    return (
      <div className="py-1 px-4 text-center">
        <span className="text-xs text-red-400">{message.message}</span>
      </div>
    );
  }

  return (
    <div className="py-2 px-4 hover:bg-gray-50">
      <div className="flex items-baseline gap-2">
        <span className="text-sm font-semibold text-gray-900">
          {message.sender.username}
        </span>
        <span className="text-xs text-gray-400">
          {formatTime(message.timestamp)}
        </span>
      </div>
      <p className="text-sm text-gray-700 whitespace-pre-wrap break-words">
        {message.content}
      </p>
    </div>
  );
}
