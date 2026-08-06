import { useState } from "react";
import { Text, TextInput } from "react-native";
import { useLocalSearchParams } from "expo-router";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Action, Body, Card, Screen, State, s } from "../../../components/ui";
import { useAPI } from "../../../lib/api";
export function RecommendationActions({
  allowed,
  onAccept,
  onReject,
}: {
  allowed: boolean;
  onAccept(): void;
  onReject(reason: string): void;
}) {
  const [reason, setReason] = useState("");
  if (!allowed)
    return (
      <Card>
        <Body>
          Your role can view evidence but cannot manage recommendations.
        </Body>
      </Card>
    );
  return (
    <Card>
      <Action label="Accept for evaluation" onPress={onAccept} />
      <TextInput
        accessibilityLabel="Rejection reason"
        onChangeText={setReason}
        placeholder="Reason for rejection"
        placeholderTextColor="#7890a8"
        style={s.input}
        value={reason}
      />
      <Action
        danger
        disabled={!reason.trim()}
        label="Reject"
        onPress={() => onReject(reason)}
      />
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
      queryKey: ["recommendation", id],
      queryFn: async () => {
        const r = await api.GET("/api/v1/recommendations/{recommendationID}", {
          params: { path: { recommendationID: id } },
        });
        if (r.error) throw r.error;
        return r.data;
      },
    });
  const action = useMutation({
    mutationFn: async (v: { kind: "accept" | "reject"; reason?: string }) => {
      const path =
        v.kind === "accept"
          ? "/api/v1/recommendations/{recommendationID}/accept"
          : "/api/v1/recommendations/{recommendationID}/reject";
      const r = await api.POST(
        path as any,
        {
          params: { path: { recommendationID: id } },
          body: { reason: v.reason },
        } as any,
      );
      if (r.error) throw r.error;
    },
  });
  const x: any = q.data,
    allowed = ["owner", "admin", "engineer"].includes((me.data as any)?.role);
  return (
    <Screen eyebrow="Decision support" title="Recommendation">
      <State
        error={q.error || me.error}
        loading={q.isLoading || me.isLoading}
      />
      {x ? (
        <>
          <Card>
            <Text style={s.title}>
              {x.recommendation_type.replaceAll("_", " ")}
            </Text>
            <Body>{x.reason_codes?.join(", ")}</Body>
            <Body>
              Estimated savings: {x.estimated_monthly_savings} {x.currency}.
              This is not verified savings.
            </Body>
            <Body>
              Quality risk: {x.quality_risk} · Operational risk:{" "}
              {x.operational_risk}
            </Body>
          </Card>
          <RecommendationActions
            allowed={allowed}
            onAccept={() => action.mutate({ kind: "accept" })}
            onReject={(reason) => action.mutate({ kind: "reject", reason })}
          />
        </>
      ) : null}
    </Screen>
  );
}
