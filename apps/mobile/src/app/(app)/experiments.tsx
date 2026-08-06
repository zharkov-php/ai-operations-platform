import { Text } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { Body, Card, NavLink, Screen, State, s } from "../../components/ui";
import { useAPI } from "../../lib/api";
export default function Experiments() {
  const api = useAPI();
  const q = useQuery({
    queryKey: ["experiments"],
    queryFn: async () => {
      const r = await api.GET("/api/v1/experiments");
      if (!r.data) throw new Error("Unable to load experiments");
      return r.data;
    },
  });
  const items = (q.data as any)?.items ?? [];
  return (
    <Screen eyebrow="Controlled change" title="Experiments">
      <State
        error={q.error}
        loading={q.isLoading}
        empty={
          !q.isLoading && !q.error && !items.length
            ? "No experiments are active."
            : undefined
        }
      />
      {items.map((x: any) => (
        <Card key={x.id}>
          <Text style={s.title}>{x.status}</Text>
          <Body>{x.traffic_percentage}% candidate traffic</Body>
          <NavLink
            href={{ pathname: "/(app)/experiments/[id]", params: { id: x.id } }}
          >
            Monitor experiment
          </NavLink>
        </Card>
      ))}
    </Screen>
  );
}
