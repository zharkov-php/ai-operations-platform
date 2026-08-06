"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { api, apiMessage } from "@/lib/api";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

export function WorkloadDetail({ id }: { id: string }) {
  const query = useQuery({ queryKey: ["workload", id], queryFn: async () => {
    const [workload, overview, providers, models, recommendations] = await Promise.all([
      api.GET("/api/v1/workloads/{workloadID}", { params: { path: { workloadID: id } } }),
      api.GET("/api/v1/analytics/overview", { params: { query: { workload_id: id } } }),
      api.GET("/api/v1/analytics/costs", { params: { query: { dimension: "provider", workload_id: id } } }),
      api.GET("/api/v1/analytics/costs", { params: { query: { dimension: "model", workload_id: id } } }), api.GET("/api/v1/recommendations"),
    ]);
    const error = workload.error ?? overview.error ?? providers.error ?? models.error ?? recommendations.error;
    if (error) throw error;
    return { workload: workload.data!, overview: overview.data!, providers: providers.data!, models: models.data!, recommendations: recommendations.data!.items.filter((item) => item.workload_id === id) };
  } });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState message={apiMessage(query.error)} />;
  const { workload, overview, providers, models, recommendations } = query.data;
  return <><DashboardHeading eyebrow={`${workload.type.replaceAll("_", " ")} workload`} title={workload.name} description={`${workload.owner} · ${workload.privacy_classification} data · ${workload.quality_requirement} quality`} /><section className="metric-grid"><article><span>Observed cost</span><strong>{overview.currency} {Number(overview.observed_cost).toFixed(2)}</strong></article><article><span>Call volume</span><strong>{overview.call_count.toLocaleString()}</strong></article><article><span>Latency</span><strong>{Number(overview.average_latency_ms).toFixed(0)} ms</strong></article><article><span>Error rate</span><strong>{(Number(overview.error_rate) * 100).toFixed(2)}%</strong></article></section><div className="dashboard-grid"><section className="dashboard-panel"><h2>Token distribution</h2><dl className="result-list"><div><dt>Input</dt><dd>{overview.input_tokens.toLocaleString()}</dd></div><div><dt>Output</dt><dd>{overview.output_tokens.toLocaleString()}</dd></div><div><dt>Cached input</dt><dd>{overview.cached_input_tokens.toLocaleString()}</dd></div></dl></section><section className="dashboard-panel"><h2>Provider and model usage</h2><ul className="compact-list">{providers.breakdown.map((item) => <li key={`provider-${item.key}`}><span>{item.key}</span><strong>{item.currency} {Number(item.cost).toFixed(2)}</strong></li>)}{models.breakdown.map((item) => <li key={`model-${item.key}`}><span>{item.key}</span><strong>{item.calls} calls</strong></li>)}</ul></section><section className="dashboard-panel wide"><h2>Recommendation history</h2>{recommendations.length ? <ul className="compact-list">{recommendations.map((item) => <li key={item.id}><Link href={`/dashboard/recommendations/${item.id}`}>{item.recommendation_type.replaceAll("_", " ")}</Link><span>{item.priority} · {item.status} · {item.confidence_level} confidence</span></li>)}</ul> : <EmptyState title="No recommendation history" message="Evidence analysis has not produced a recommendation for this workload." />}</section><section className="dashboard-panel wide"><h2>Evaluation and experiment history</h2><EmptyState title="No evaluated candidates" message="Evaluation datasets and controlled experiments are introduced in later phases." /></section></div></>;
}
