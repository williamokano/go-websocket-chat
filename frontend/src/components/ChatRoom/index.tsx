import { useAuth } from "../../hooks/useAuth";
import { useChat } from "../../hooks/useChat";
import { ConnectionStatus } from "../ConnectionStatus";
import { MessageInput } from "../MessageInput";
import { MessageList } from "../MessageList";
import { UserInfo } from "../UserInfo";

export function ChatRoom() {
  const { user, logout } = useAuth();
  const { messages, sendMessage, connectionStatus, onlineCount } = useChat();

  return (
    <div className="flex flex-col h-screen bg-white">
      <header className="flex items-center justify-between border-b border-gray-200 px-4 py-3">
        <ConnectionStatus status={connectionStatus} onlineCount={onlineCount} />
        <UserInfo user={user!} onLogout={logout} />
      </header>
      <MessageList messages={messages} />
      <MessageInput onSend={sendMessage} connectionStatus={connectionStatus} />
    </div>
  );
}
