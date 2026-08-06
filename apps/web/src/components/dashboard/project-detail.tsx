"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { api, apiMessage } from "@/lib/api";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

export function ProjectDetail({ id }: { id: string }) {
  const query = useQuery({ queryKey: ["project", id], queryFn: async () => { const [project, workloads, analytics] = await Promise.all([api.GET("/api/v1/projects/{projectID}", { params: { path: { projectID: id } } }), api.GET("/api/v1/workloads", { params: { query: { project_id: id, limit: 100 } } }), api.GET("/api/v1/analytics/overview", { params: { query: { project_id: id } } })]); const error = project.error ?? workloads.error ?? analytics.error; if (error) throw error; return { project: project.data!, workloads: workloads.data!, analytics: analytics.data! }; } });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState message={apiMessage(query.error)} />;
  const { project, workloads, analytics } = query.data;
  return <><DashboardHeading eyebrow={`${project.environment} project`} title={project.name} description={`Budget ${project.currency} ${project.monthly_budget} · ${project.status}`} /><section className="metric-grid"><article><span>Observed cost</span><strong>{analytics.currency} {Number(analytics.observed_cost).toFixed(2)}</strong></article><article><span>Calls</span><strong>{analytics.call_count.toLocaleString()}</strong></article><article><span>Average latency</span><strong>{Number(analytics.average_latency_ms).toFixed(0)} ms</strong></article><article><span>Error rate</span><strong>{(Number(analytics.error_rate) * 100).toFixed(2)}%</strong></article></section><section className="dashboard-panel"><h2>Workloads</h2>{workloads.items.length ? <ul className="compact-list">{workloads.items.map((item) => <li key={item.id}><Link href={`/dashboard/workloads/${item.id}`}>{item.name}</Link><span>{item.type} · {item.criticality}</span></li>)}</ul> : <EmptyState title="No workloads" message="Register the first workload for this project." />}</section></>;
}
