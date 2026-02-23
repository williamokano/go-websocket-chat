import { createContext, useCallback, useEffect, useState } from "react";
import type { ReactNode } from "react";
import type { User } from "../types";
import * as authApi from "../api/auth";
import * as webauthnApi from "../api/webauthn";

export interface AuthContextValue {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string) => Promise<void>;
  logout: () => void;
  loginWithPasskey: () => Promise<void>;
  registerWithPasskey: (username: string) => Promise<void>;
  supportsWebAuthn: boolean;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

const TOKEN_KEY = "chat_token";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const stored = localStorage.getItem(TOKEN_KEY);
    if (!stored) {
      setIsLoading(false);
      return;
    }
    authApi
      .getMe(stored)
      .then((u) => {
        setToken(stored);
        setUser(u);
      })
      .catch(() => {
        localStorage.removeItem(TOKEN_KEY);
      })
      .finally(() => setIsLoading(false));
  }, []);

  const login = useCallback(async (username: string, password: string) => {
    const res = await authApi.login(username, password);
    localStorage.setItem(TOKEN_KEY, res.token);
    setToken(res.token);
    setUser(res.user);
  }, []);

  const register = useCallback(async (username: string, password: string) => {
    const res = await authApi.register(username, password);
    localStorage.setItem(TOKEN_KEY, res.token);
    setToken(res.token);
    setUser(res.user);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setUser(null);
  }, []);

  const supportsWebAuthn =
    typeof window !== "undefined" &&
    window.PublicKeyCredential !== undefined;

  const loginWithPasskey = useCallback(async () => {
    const { sessionId, options } = await webauthnApi.beginLogin();
    const credential = (await navigator.credentials.get({
      publicKey: options,
    })) as PublicKeyCredential | null;
    if (!credential) {
      throw new Error("Passkey authentication was cancelled");
    }
    const res = await webauthnApi.finishLogin(sessionId, credential);
    localStorage.setItem(TOKEN_KEY, res.token);
    setToken(res.token);
    setUser(res.user);
  }, []);

  const registerWithPasskey = useCallback(async (username: string) => {
    const { sessionId, options } = await webauthnApi.beginRegistration(username);
    const credential = (await navigator.credentials.create({
      publicKey: options,
    })) as PublicKeyCredential | null;
    if (!credential) {
      throw new Error("Passkey registration was cancelled");
    }
    const res = await webauthnApi.finishRegistration(sessionId, credential);
    localStorage.setItem(TOKEN_KEY, res.token);
    setToken(res.token);
    setUser(res.user);
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!user && !!token,
        isLoading,
        login,
        register,
        logout,
        loginWithPasskey,
        registerWithPasskey,
        supportsWebAuthn,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
