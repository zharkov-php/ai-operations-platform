"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { api, apiMessage } from "@/lib/api";
import { CostChart } from "./cost-chart";
import { ReliabilityChart, TokenChart } from "./operational-charts";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

const money = (value: string, currency: string) => `${currency} ${Number(value).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 4 })}`;
const percent = (value: string) => `${(Number(value) * 100).toFixed(2)}%`;

export function Overview() {
  const query = useQuery({ queryKey: ["dashboard", "overview"], queryFn: async () => {
    const [overview, costs, tokens, latency, errors, projects, workloads, recommendations] = await Promise.all([
      api.GET("/api/v1/analytics/overview"), api.GET("/api/v1/analytics/costs", { params: { query: { dimension: "project" } } }),
      api.GET("/api/v1/analytics/tokens"), api.GET("/api/v1/analytics/latency"), api.GET("/api/v1/analytics/errors"),
      api.GET("/api/v1/projects", { params: { query: { limit: 5 } } }), api.GET("/api/v1/workloads", { params: { query: { limit: 5 } } }), api.GET("/api/v1/recommendations"),
    ]);
    const error = overview.error ?? costs.error ?? tokens.error ?? latency.error ?? errors.error ?? projects.error ?? workloads.error ?? recommendations.error;
    if (error) throw error;
    return { overview: overview.data!, costs: costs.data!, tokens: tokens.data!, latency: latency.data!, errors: errors.data!, projects: projects.data!, workloads: workloads.data!, recommendations: recommendations.data! };
  } });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState message={apiMessage(query.error)} />;
  const { overview, costs, tokens, latency, errors, projects, workloads, recommendations } = query.data;
  const priorityCount = recommendations.items.reduce<Record<string, number>>((acc, item) => { acc[item.priority] = (acc[item.priority] ?? 0) + 1; return acc; }, {});
  return <><DashboardHeading eyebrow="Organization overview" title="AI operations at a glance" description="Observed usage remains separate from projections and estimated opportunities." /><section className="metric-grid" aria-label="Cost summary"><article><span>Observed cost</span><strong>{money(overview.observed_cost, overview.currency)}</strong><small>Current selected period</small></article><article><span>Monthly projection</span><strong>{money(overview.estimated_monthly_projection, overview.currency)}</strong><small>{overview.projection_method}</small></article><article><span>Verified savings</span><strong>Not available</strong><small>No verified experiment result yet</small></article><article><span>Calls</span><strong>{overview.call_count.toLocaleString()}</strong><small>{percent(overview.error_rate)} error rate</small></article></section><div className="dashboard-grid"><section className="dashboard-panel wide"><div className="panel-title"><div><p className="card-label">Observed cost</p><h2>Daily trend</h2></div></div>{costs.daily.length ? <CostChart data={costs.daily} currency={costs.currency} /> : <EmptyState title="No cost history" message="Ingest calls in the selected period to produce a trend." />}</section><section className="dashboard-panel"><p className="card-label">Tokens</p><h2>Token distribution</h2><TokenChart value={tokens} /></section><section className="dashboard-panel"><p className="card-label">Operations</p><h2>Latency and failures</h2><ReliabilityChart latency={latency} errors={errors} /></section><section className="dashboard-panel"><p className="card-label">Recommendation priority</p><h2>Review queue</h2>{recommendations.items.length ? <dl className="result-list">{Object.entries(priorityCount).map(([priority, count]) => <div key={priority}><dt>{priority}</dt><dd>{count}</dd></div>)}</dl> : <EmptyState title="No recommendations" message="Run evidence analysis after sufficient calls are available." />}<Link className="text-link" href="/dashboard/recommendations">Review recommendations →</Link></section><section className="dashboard-panel"><p className="card-label">Portfolio</p><h2>Projects</h2>{projects.items.length ? <ul className="compact-list">{projects.items.map((project) => <li key={project.id}><Link href={`/dashboard/projects/${project.id}`}>{project.name}</Link><span>{project.environment}</span></li>)}</ul> : <EmptyState title="No projects" message="Create a project through the API to begin." />}<Link className="text-link" href="/dashboard/projects">All projects →</Link></section><section className="dashboard-panel"><p className="card-label">Workload inventory</p><h2>Workloads</h2>{workloads.items.length ? <ul className="compact-list">{workloads.items.map((workload) => <li key={workload.id}><Link href={`/dashboard/workloads/${workload.id}`}>{workload.name}</Link><span>{workload.criticality}</span></li>)}</ul> : <EmptyState title="No workloads" message="Workloads appear after they are registered." />}<Link className="text-link" href="/dashboard/workloads">All workloads →</Link></section></div></>;
}
