"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { api, apiMessage } from "@/lib/api";
import { DashboardHeading, ErrorState, LoadingState } from "./states";

function JsonEvidence({ value }: { value: Record<string, unknown> }) { return <dl className="evidence-json">{Object.entries(value).map(([key, item]) => <div key={key}><dt>{key.replaceAll("_", " ")}</dt><dd>{typeof item === "object" ? JSON.stringify(item) : String(item)}</dd></div>)}</dl>; }
const rejectionSchema = z.object({ reason: z.string().trim().min(1).max(1000) });
type Rejection = z.infer<typeof rejectionSchema>;

export function RecommendationDetail({ id }: { id: string }) {
  const client = useQueryClient();
  const [actionError, setActionError] = useState("");
  const rejection = useForm<Rejection>({ defaultValues: { reason: "" } });
  const query = useQuery({ queryKey: ["recommendation", id], queryFn: async () => {
    const [recommendation, audit, me] = await Promise.all([api.GET("/api/v1/recommendations/{recommendationID}", { params: { path: { recommendationID: id } } }), api.GET("/api/v1/recommendations/{recommendationID}/audit-history", { params: { path: { recommendationID: id } } }), api.GET("/api/v1/me")]);
    const error = recommendation.error ?? audit.error ?? me.error;
    if (error) throw error;
    return { item: recommendation.data!, history: audit.data!.items, me: me.data! };
  } });
  const review = useMutation({ mutationFn: async ({ action, reason }: { action: "accept" | "reject"; reason: string }) => {
    const response = action === "accept" ? await api.POST("/api/v1/recommendations/{recommendationID}/accept", { params: { path: { recommendationID: id } }, body: { reason } }) : await api.POST("/api/v1/recommendations/{recommendationID}/reject", { params: { path: { recommendationID: id } }, body: { reason } });
    if (response.error) throw response.error;
    return response.data;
  }, onSuccess: async () => { setActionError(""); rejection.reset(); await Promise.all([client.invalidateQueries({ queryKey: ["recommendation", id] }), client.invalidateQueries({ queryKey: ["recommendations"] })]); }, onError: (error) => setActionError(apiMessage(error)) });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState message={apiMessage(query.error)} />;
  const { item, history, me } = query.data;
  const canManage = ["owner", "admin", "engineer"].includes(me.role);
  const reviewable = item.status === "new" || item.status === "under_review";
  const submitRejection = rejection.handleSubmit((value) => { const parsed = rejectionSchema.safeParse(value); if (!parsed.success) return; review.mutate({ action: "reject", reason: parsed.data.reason }); });
  return <><DashboardHeading eyebrow={`${item.priority} priority · ${item.status}`} title={item.recommendation_type.replaceAll("_", " ")} description={`Rule ${item.rule_version} · ${item.confidence_level} confidence · savings remain estimated`} /><section className="metric-grid"><article><span>Estimated savings</span><strong>{item.currency} {item.estimated_monthly_savings}</strong><small>Not verified</small></article><article><span>Implementation</span><strong>{item.currency} {item.estimated_implementation_cost}</strong></article><article><span>Break-even</span><strong>{item.estimated_break_even_months ?? "Not available"}</strong><small>months</small></article><article><span>Next action</span><strong>{item.required_next_action.replaceAll("_", " ")}</strong></article></section><div className="dashboard-grid"><section className="dashboard-panel"><h2>Current execution</h2><JsonEvidence value={item.current_execution} /></section><section className="dashboard-panel"><h2>Proposed execution</h2><JsonEvidence value={item.proposed_execution} /></section><section className="dashboard-panel"><h2>Evidence</h2><JsonEvidence value={item.evidence_summary} /><h3>Reason codes</h3><ul>{item.reason_codes.map((code) => <li key={code}>{code}</li>)}</ul></section><section className="dashboard-panel"><h2>Confidence inputs</h2><JsonEvidence value={item.confidence_inputs} /><p>Numeric confidence describes measurable evidence coverage; it is not a statistically validated quality percentage.</p></section><section className="dashboard-panel"><h2>Risk and evaluation requirements</h2><p><strong>Quality:</strong> {item.quality_risk}</p><p><strong>Operational:</strong> {item.operational_risk}</p><p><strong>Required:</strong> {item.required_next_action.replaceAll("_", " ")}</p><p>Acceptance authorizes evaluation or implementation planning only. It never changes production routing.</p></section><section className="dashboard-panel"><h2>Review action</h2>{!canManage ? <p className="permission-note">Your {me.role} role can inspect evidence but cannot review recommendations.</p> : !reviewable ? <p>This recommendation was already reviewed. Duplicate actions are disabled.</p> : <><button className="button" type="button" disabled={review.isPending} onClick={() => review.mutate({ action: "accept", reason: "accepted for evaluation" })}>Accept for evaluation</button><form className="review-form" onSubmit={submitRejection}><label>Rejection reason<textarea rows={4} maxLength={1000} {...rejection.register("reason", { required: true })} /></label>{rejection.formState.errors.reason && <span className="form-error">A rejection reason is required.</span>}<button className="danger-button" type="submit" disabled={review.isPending}>Reject recommendation</button></form></>}{actionError && <p className="form-error" role="alert">{actionError}</p>}</section><section className="dashboard-panel wide"><h2>Audit timeline</h2>{history.length ? <ol className="audit-timeline">{history.map((entry) => <li key={entry.id}><strong>{entry.action.replaceAll(".", " · ")}</strong><span>{new Date(entry.created_at).toLocaleString("en-US", { timeZone: "UTC" })} UTC</span><p>{String(entry.metadata.reason || "No reason recorded")}</p></li>)}</ol> : <p>No review actions have been recorded.</p>}</section></div></>;
}
