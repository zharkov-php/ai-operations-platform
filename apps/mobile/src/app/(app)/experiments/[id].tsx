import { Text } from "react-native";
import { useLocalSearchParams } from "expo-router";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Action, Body, Card, Screen, State, s } from "../../../components/ui";
import { useAPI } from "../../../lib/api";
export function ExperimentActions({
  allowed,
  status,
  onPause,
  onRollback,
}: {
  allowed: boolean;
  status: string;
  onPause(): void;
  onRollback(): void;
}) {
  if (!allowed) return <Body>Your role cannot control experiments.</Body>;
  return (
    <Card>
      {status === "running" ? (
        <Action label="Pause experiment" onPress={onPause} />
      ) : null}
      {["running", "paused"].includes(status) ? (
        <Action danger label="Rollback" onPress={onRollback} />
      ) : null}
    </Card>
  );
}
export default function Detail() {
  const { id } = useLocalSearchParams<{ id: string }>(),
    api = useAPI();
  const me = useQuery({
      queryKey: ["me"],
      queryFn: async () => {
        const r = await api.GET("/api/v1/me");
        if (r.error) throw r.error;
        return r.data;
      },
    }),
    q = useQuery({
      queryKey: ["experiment", id],
      queryFn: async () => {
        const r = await api.GET("/api/v1/experiments/{experimentID}", {
          params: { path: { experimentID: id } },
        });
        if (r.error) throw r.error;
        return r.data;
      },
    });
  const action = useMutation({
    mutationFn: async (kind: "pause" | "rollback") => {
      const path =
        kind === "pause"
          ? "/api/v1/experiments/{experimentID}/pause"
          : "/api/v1/experiments/{experimentID}/rollback";
      const r = await api.POST(
        path as any,
        {
          params: { path: { experimentID: id } },
          body:
            kind === "rollback"
              ? { reason: "Manual mobile rollback" }
              : undefined,
        } as any,
      );
      if (r.error) throw r.error;
    },
  });
  const x: any = q.data,
    allowed = ["owner", "admin", "engineer"].includes((me.data as any)?.role);
  return (
    <Screen eyebrow="Guardrails" title="Experiment">
      <State
        error={q.error || me.error}
        loading={q.isLoading || me.isLoading}
      />
      {x ? (
        <>
          <Card>
            <Text style={s.title}>{x.status}</Text>
            <Body>Candidate traffic: {x.traffic_percentage}%</Body>
            <Body>
              Verified savings: {x.verified_savings ?? "Not verified"}
            </Body>
            {x.rollback_reason ? (
              <Body>Rollback: {x.rollback_reason}</Body>
            ) : null}
          </Card>
          <ExperimentActions
            allowed={allowed}
            status={x.status}
            onPause={() => action.mutate("pause")}
            onRollback={() => action.mutate("rollback")}
          />
        </>
      ) : null}
    </Screen>
  );
}
