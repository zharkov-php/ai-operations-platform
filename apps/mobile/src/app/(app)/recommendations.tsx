import { Text } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { Body, Card, NavLink, Screen, State, s } from "../../components/ui";
import { useAPI } from "../../lib/api";
export default function Recommendations() {
  const api = useAPI();
  const q = useQuery({
    queryKey: ["recommendations"],
    queryFn: async () => {
      const r = await api.GET("/api/v1/recommendations");
      if (!r.data) throw new Error("Unable to load recommendations");
      return r.data;
    },
  });
  const items = (q.data as any)?.items ?? [];
  return (
    <Screen eyebrow="Evidence" title="Recommendations">
      <State
        error={q.error}
        loading={q.isLoading}
        empty={
          !q.isLoading && !q.error && !items.length
            ? "No recommendations yet."
            : undefined
        }
      />
      {items.map((x: any) => (
        <Card key={x.id}>
          <Text style={s.title}>
            {x.recommendation_type.replaceAll("_", " ")}
          </Text>
          <Body>
            {x.priority} priority · {x.confidence_level} confidence
          </Body>
          <NavLink
            href={{
              pathname: "/(app)/recommendations/[id]",
              params: { id: x.id },
            }}
          >
            Review evidence
          </NavLink>
        </Card>
      ))}
    </Screen>
  );
}
