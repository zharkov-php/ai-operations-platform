import * as SecureStore from "expo-secure-store";
export interface TokenStorage {
  get(): Promise<string | null>;
  set(value: string): Promise<void>;
  clear(): Promise<void>;
}
export const secureTokenStorage: TokenStorage = {
  get: () => SecureStore.getItemAsync("ai_execution_session"),
  set: (value) => SecureStore.setItemAsync("ai_execution_session", value),
  clear: () => SecureStore.deleteItemAsync("ai_execution_session"),
};
