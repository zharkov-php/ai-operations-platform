import { Text } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { Body, Card, NavLink, Screen, State, s } from "../../components/ui";
import { useAPI } from "../../lib/api";
export function OverviewView({ data }: { data: any }) {
  return (
    <>
      <Card>
        <Text style={s.title}>
          {data?.current_month_cost ?? "0.00"} {data?.currency ?? "USD"}
        </Text>
        <Body>Current observed cost</Body>
        <Body>
          Projected: {data?.projected_month_cost ?? "0.00"} · Verified savings:{" "}
          {data?.verified_savings ?? "0.00"}
        </Body>
      </Card>
      <NavLink href="/(app)/alerts">Budget alerts</NavLink>
      <NavLink href="/(app)/recommendations">Recommendations</NavLink>
      <NavLink href="/(app)/experiments">Active experiments</NavLink>
      <NavLink href="/(app)/projects">Projects</NavLink>
      <NavLink href="/(app)/settings">Settings</NavLink>
    </>
  );
}
export default function Overview() {
  const api = useAPI();
  const query = useQuery({
    queryKey: ["overview"],
    queryFn: async () => {
      const r = await api.GET("/api/v1/analytics/overview");
      if (!r.data) throw new Error("Unable to load overview");
      return r.data;
    },
  });
  return (
    <Screen eyebrow="Operations" title="Overview">
      <State error={query.error} loading={query.isLoading} />
      {query.data ? <OverviewView data={query.data} /> : null}
    </Screen>
  );
}
