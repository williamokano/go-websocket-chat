export interface User {
  id: string;
  username: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ClientMessage {
  type: "send_message";
  content: string;
}

export type ServerMessage =
  | ChatMessage
  | UserJoinedMessage
  | UserLeftMessage
  | ErrorMessage;

export interface ChatMessage {
  type: "chat_message";
  id: string;
  content: string;
  sender: User;
  timestamp: string;
}

export interface UserJoinedMessage {
  type: "user_joined";
  user: User;
  online_count: number;
  timestamp: string;
}

export interface UserLeftMessage {
  type: "user_left";
  user: User;
  online_count: number;
  timestamp: string;
}

export interface ErrorMessage {
  type: "error";
  message: string;
}

export type ConnectionStatus = "connected" | "connecting" | "disconnected";
