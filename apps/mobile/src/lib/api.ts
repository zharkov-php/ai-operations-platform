import { createAPIClient } from "@ai-operations/api-client";
import { apiBase, useAuth } from "./auth";
export function useAPI() {
  const { session } = useAuth();
  return createAPIClient(
    apiBase,
    session ? { Authorization: `Bearer ${session.accessToken}` } : {},
  );
}
