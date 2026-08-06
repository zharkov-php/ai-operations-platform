import { Text } from "react-native";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Action, Body, Card, Screen, State, s } from "../../components/ui";
import { useAPI } from "../../lib/api";
export function AlertCard({
  alert,
  onAcknowledge,
}: {
  alert: any;
  onAcknowledge(): void;
}) {
  return (
    <Card>
      <Text style={s.title}>{alert.severity} alert</Text>
      <Body>
        {alert.threshold_type}: {alert.threshold_value}
      </Body>
      {alert.status === "open" ? (
        <Action label="Acknowledge" onPress={onAcknowledge} />
      ) : (
        <Body>Acknowledged</Body>
      )}
    </Card>
  );
}
export default function Alerts() {
  const api = useAPI(),
    qc = useQueryClient();
  const q = useQuery({
    queryKey: ["alerts"],
    queryFn: async () => {
      const r = await api.GET("/api/v1/budget-alerts");
      if (!r.data) throw new Error("Unable to load alerts");
      return r.data;
    },
  });
  const m = useMutation({
    mutationFn: async (id: string) => {
      const r = await api.POST("/api/v1/budget-alerts/{alertID}/acknowledge", {
        params: { path: { alertID: id } },
      });
      if (r.error) throw r.error;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["alerts"] }),
  });
  const items = (q.data as any)?.items ?? [];
  return (
    <Screen eyebrow="Governance" title="Budget alerts">
      <State
        error={q.error}
        loading={q.isLoading}
        empty={
          !q.isLoading && !q.error && !items.length
            ? "No active budget alerts."
            : undefined
        }
      />
      {items.map((a: any) => (
        <AlertCard alert={a} key={a.id} onAcknowledge={() => m.mutate(a.id)} />
      ))}
    </Screen>
  );
}
