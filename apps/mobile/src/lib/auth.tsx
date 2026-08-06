import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { createAPIClient, type components } from "@ai-operations/api-client";
import { secureTokenStorage, type TokenStorage } from "./storage";
type Session = { accessToken: string; refreshToken: string };
type AuthValue = {
  session: Session | null;
  loading: boolean;
  signIn(email: string, password: string): Promise<void>;
  signOut(): Promise<void>;
};
const Context = createContext<AuthValue | null>(null);
export const apiBase =
  process.env.EXPO_PUBLIC_API_URL ?? "http://localhost:8080";
export function AuthProvider({
  children,
  storage = secureTokenStorage,
}: {
  children: React.ReactNode;
  storage?: TokenStorage;
}) {
  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    storage
      .get()
      .then(async (raw) => {
        if (!raw) return;
        try {
          const saved = JSON.parse(raw) as Session;
          const response = await createAPIClient(apiBase).POST(
            "/api/v1/auth/refresh",
            { body: { refresh_token: saved.refreshToken } },
          );
          if (response.data?.access_token && response.data.refresh_token) {
            const next = {
              accessToken: response.data.access_token,
              refreshToken: response.data.refresh_token,
            };
            await storage.set(JSON.stringify(next));
            setSession(next);
          } else await storage.clear();
        } catch {
          await storage.clear();
        }
      })
      .finally(() => setLoading(false));
  }, [storage]);
  const value = useMemo<AuthValue>(
    () => ({
      session,
      loading,
      signIn: async (email, password) => {
        const response = await createAPIClient(apiBase).POST(
          "/api/v1/auth/login",
          { body: { email, password, client: "mobile" } },
        );
        if (!response.data?.access_token || !response.data.refresh_token)
          throw new Error("Sign-in failed");
        const next = {
          accessToken: response.data.access_token,
          refreshToken: response.data.refresh_token,
        };
        await storage.set(JSON.stringify(next));
        setSession(next);
      },
      signOut: async () => {
        if (session)
          await createAPIClient(apiBase, {
            Authorization: `Bearer ${session.accessToken}`,
          }).POST("/api/v1/auth/logout", {
            body: { refresh_token: session.refreshToken },
          });
        await storage.clear();
        setSession(null);
      },
    }),
    [loading, session, storage],
  );
  return <Context.Provider value={value}>{children}</Context.Provider>;
}
export function useAuth() {
  const value = useContext(Context);
  if (!value) throw new Error("useAuth requires AuthProvider");
  return value;
}
export type Role = components["schemas"]["CurrentUser"]["role"];
