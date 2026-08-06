import { Text } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { Body, Card, Screen, State, s } from "../../components/ui";
import { useAPI } from "../../lib/api";
export function NotificationList({
  items,
}: {
  items: { id: string; title: string; body: string; status: string }[];
}) {
  return (
    <>
      {items.map((item) => (
        <Card key={item.id}>
          <Text style={s.title}>{item.title}</Text>
          <Body>{item.body}</Body>
          <Body>{item.status.replaceAll("_", " ")}</Body>
        </Card>
      ))}
    </>
  );
}
export default function Notifications() {
  const api = useAPI();
  const query = useQuery({
    queryKey: ["notifications"],
    queryFn: async () => {
      const response = await api.GET("/api/v1/notifications");
      if (!response.data) throw new Error("Unable to load notifications");
      return response.data;
    },
  });
  const items = query.data?.items ?? [];
  return (
    <Screen eyebrow="Activity" title="Notifications">
      <State
        loading={query.isLoading}
        error={query.error}
        empty={
          !query.isLoading && !query.error && !items.length
            ? "No notifications yet."
            : undefined
        }
      />
      <NotificationList items={items} />
    </Screen>
  );
}
