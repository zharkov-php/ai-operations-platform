"use client";

import { useQuery } from "@tanstack/react-query";
import { api, apiMessage } from "@/lib/api";
import { DashboardHeading, ErrorState, LoadingState } from "./states";

function JsonEvidence({ value }: { value: Record<string, unknown> }) { return <dl className="evidence-json">{Object.entries(value).map(([key, item]) => <div key={key}><dt>{key.replaceAll("_", " ")}</dt><dd>{typeof item === "object" ? JSON.stringify(item) : String(item)}</dd></div>)}</dl>; }

export function RecommendationDetail({ id }: { id: string }) {
  const query = useQuery({ queryKey: ["recommendation", id], queryFn: async () => { const response = await api.GET("/api/v1/recommendations/{recommendationID}", { params: { path: { recommendationID: id } } }); if (response.error) throw response.error; return response.data; } });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState message={apiMessage(query.error)} />;
  const item = query.data;
  return <><DashboardHeading eyebrow={`${item.priority} priority · ${item.status}`} title={item.recommendation_type.replaceAll("_", " ")} description={`Rule ${item.rule_version} · ${item.confidence_level} confidence · savings remain estimated`} /><section className="metric-grid"><article><span>Estimated savings</span><strong>{item.currency} {item.estimated_monthly_savings}</strong><small>Not verified</small></article><article><span>Implementation</span><strong>{item.currency} {item.estimated_implementation_cost}</strong></article><article><span>Break-even</span><strong>{item.estimated_break_even_months ?? "Not available"}</strong><small>months</small></article><article><span>Next action</span><strong>{item.required_next_action.replaceAll("_", " ")}</strong></article></section><div className="dashboard-grid"><section className="dashboard-panel"><h2>Current execution</h2><JsonEvidence value={item.current_execution} /></section><section className="dashboard-panel"><h2>Proposed execution</h2><JsonEvidence value={item.proposed_execution} /></section><section className="dashboard-panel"><h2>Evidence</h2><JsonEvidence value={item.evidence_summary} /></section><section className="dashboard-panel"><h2>Confidence inputs</h2><JsonEvidence value={item.confidence_inputs} /></section><section className="dashboard-panel wide"><h2>Risk boundary</h2><p><strong>Quality:</strong> {item.quality_risk}</p><p><strong>Operational:</strong> {item.operational_risk}</p><p>Accept and reject actions are intentionally unavailable until the audited state-transition workflow is implemented in Phase 12.</p></section></div></>;
}
