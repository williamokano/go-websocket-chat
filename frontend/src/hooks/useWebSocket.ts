import { useCallback, useEffect, useRef, useState } from "react";
import type { ClientMessage, ConnectionStatus, ServerMessage } from "../types";

const MAX_BACKOFF = 30_000;
const INVALID_JWT_CODE = 4001;

interface UseWebSocketOptions {
  token: string;
  onLogout: () => void;
}

interface UseWebSocketReturn {
  messages: ServerMessage[];
  sendMessage: (content: string) => void;
  connectionStatus: ConnectionStatus;
}

export function useWebSocket({
  token,
  onLogout,
}: UseWebSocketOptions): UseWebSocketReturn {
  const [messages, setMessages] = useState<ServerMessage[]>([]);
  const [connectionStatus, setConnectionStatus] =
    useState<ConnectionStatus>("disconnected");
  const wsRef = useRef<WebSocket | null>(null);
  const backoffRef = useRef(1000);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const shouldReconnectRef = useRef(true);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = `${proto}//${window.location.host}/ws?token=${token}`;

    setConnectionStatus("connecting");
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnectionStatus("connected");
      backoffRef.current = 1000;
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data) as ServerMessage;
      setMessages((prev) => [...prev, msg]);
    };

    ws.onclose = (event) => {
      setConnectionStatus("disconnected");
      wsRef.current = null;

      if (event.code === INVALID_JWT_CODE) {
        shouldReconnectRef.current = false;
        onLogout();
        return;
      }

      if (shouldReconnectRef.current) {
        const delay = backoffRef.current;
        backoffRef.current = Math.min(delay * 2, MAX_BACKOFF);
        reconnectTimerRef.current = setTimeout(connect, delay);
      }
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [token, onLogout]);

  useEffect(() => {
    shouldReconnectRef.current = true;
    connect();

    return () => {
      shouldReconnectRef.current = false;
      clearTimeout(reconnectTimerRef.current);
      wsRef.current?.close();
    };
  }, [connect]);

  const sendMessage = useCallback((content: string) => {
    if (wsRef.current?.readyState !== WebSocket.OPEN) return;
    const msg: ClientMessage = { type: "send_message", content };
    wsRef.current.send(JSON.stringify(msg));
  }, []);

  return { messages, sendMessage, connectionStatus };
}
