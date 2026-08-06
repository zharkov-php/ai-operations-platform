import { useState } from "react";
import { Stack } from "expo-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider, useAuth } from "../lib/auth";
import { State } from "../components/ui";
function Routes() {
  const { session, loading } = useAuth();
  if (loading) return <State loading />;
  return (
    <Stack
      screenOptions={{
        headerStyle: { backgroundColor: "#07111f" },
        headerTintColor: "#f5f8fc",
      }}
    >
      <Stack.Protected guard={!session}>
        <Stack.Screen name="sign-in" options={{ headerShown: false }} />
      </Stack.Protected>
      <Stack.Protected guard={!!session}>
        <Stack.Screen name="(app)" options={{ headerShown: false }} />
      </Stack.Protected>
    </Stack>
  );
}
export default function RootLayout() {
  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: { queries: { retry: 1, staleTime: 30_000 } },
      }),
  );
  return (
    <QueryClientProvider client={client}>
      <AuthProvider>
        <Routes />
      </AuthProvider>
    </QueryClientProvider>
  );
}
