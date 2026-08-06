import { Stack } from "expo-router";
export default function AppLayout() {
  return (
    <Stack
      screenOptions={{
        headerStyle: { backgroundColor: "#07111f" },
        headerTintColor: "#f5f8fc",
        headerBackTitle: "Back",
      }}
    >
      <Stack.Screen name="index" options={{ title: "Overview" }} />
      <Stack.Screen name="alerts" options={{ title: "Budget alerts" }} />
      <Stack.Screen
        name="recommendations"
        options={{ title: "Recommendations" }}
      />
      <Stack.Screen name="experiments" options={{ title: "Experiments" }} />
      <Stack.Screen name="projects" options={{ title: "Projects" }} />
      <Stack.Screen name="settings" options={{ title: "Settings" }} />
      <Stack.Screen name="notifications" options={{ title: "Notifications" }} />
    </Stack>
  );
}
