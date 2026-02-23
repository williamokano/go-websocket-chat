import { useState } from "react";
import { useAuth } from "./hooks/useAuth";
import { ChatProvider } from "./contexts/ChatContext";
import { ChatRoom } from "./components/ChatRoom";
import { LoginForm } from "./components/LoginForm";
import { RegisterForm } from "./components/RegisterForm";

type AuthView = "login" | "register";

function AuthGate() {
  const { isAuthenticated, isLoading } = useAuth();
  const [view, setView] = useState<AuthView>("login");

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-sm text-gray-400">Loading...</div>
      </div>
    );
  }

  if (isAuthenticated) {
    return (
      <ChatProvider>
        <ChatRoom />
      </ChatProvider>
    );
  }

  if (view === "register") {
    return <RegisterForm onSwitchToLogin={() => setView("login")} />;
  }

  return <LoginForm onSwitchToRegister={() => setView("register")} />;
}

export default function App() {
  return <AuthGate />;
}
