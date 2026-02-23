import type { ConnectionStatus as Status } from "../../types";

interface ConnectionStatusProps {
  status: Status;
  onlineCount: number;
}

const config: Record<Status, { color: string; label: string }> = {
  connected: { color: "bg-green-500", label: "Connected" },
  connecting: { color: "bg-yellow-500", label: "Connecting..." },
  disconnected: { color: "bg-red-500", label: "Disconnected" },
};

export function ConnectionStatus({ status, onlineCount }: ConnectionStatusProps) {
  const { color, label } = config[status];

  return (
    <div className="flex items-center gap-2 text-xs text-gray-500">
      <span className={`inline-block h-2 w-2 rounded-full ${color}`} />
      <span>{label}</span>
      {status === "connected" && onlineCount > 0 && (
        <span>&middot; {onlineCount} online</span>
      )}
    </div>
  );
}
