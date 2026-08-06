import { Text } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { Body, Card, Screen, State, s } from "../../components/ui";
import { useAPI } from "../../lib/api";
export default function Projects() {
  const api = useAPI();
  const q = useQuery({
    queryKey: ["projects"],
    queryFn: async () => {
      const r = await api.GET("/api/v1/projects");
      if (!r.data) throw new Error("Unable to load projects");
      return r.data;
    },
  });
  const items = (q.data as any)?.items ?? [];
  return (
    <Screen eyebrow="Portfolio" title="Projects">
      <State
        error={q.error}
        loading={q.isLoading}
        empty={
          !q.isLoading && !q.error && !items.length
            ? "No projects are available."
            : undefined
        }
      />
      {items.map((x: any) => (
        <Card key={x.id}>
          <Text style={s.title}>{x.name}</Text>
          <Body>
            {x.environment} · budget {x.monthly_budget} {x.currency}
          </Body>
        </Card>
      ))}
    </Screen>
  );
}
