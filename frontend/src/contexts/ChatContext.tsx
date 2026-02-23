import { createContext, useMemo } from "react";
import type { ReactNode } from "react";
import type { ConnectionStatus, ServerMessage } from "../types";
import { useWebSocket } from "../hooks/useWebSocket";
import { useAuth } from "../hooks/useAuth";

export interface ChatContextValue {
  messages: ServerMessage[];
  sendMessage: (content: string) => void;
  connectionStatus: ConnectionStatus;
  onlineCount: number;
}

export const ChatContext = createContext<ChatContextValue | null>(null);

export function ChatProvider({ children }: { children: ReactNode }) {
  const { token, logout } = useAuth();
  const { messages, sendMessage, connectionStatus } = useWebSocket({
    token: token!,
    onLogout: logout,
  });

  const onlineCount = useMemo(() => {
    for (let i = messages.length - 1; i >= 0; i--) {
      const msg = messages[i];
      if (msg.type === "user_joined" || msg.type === "user_left") {
        return msg.online_count;
      }
    }
    return 0;
  }, [messages]);

  return (
    <ChatContext.Provider
      value={{ messages, sendMessage, connectionStatus, onlineCount }}
    >
      {children}
    </ChatContext.Provider>
  );
}
