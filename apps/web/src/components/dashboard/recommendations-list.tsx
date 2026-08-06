"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { api, apiMessage } from "@/lib/api";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

export function RecommendationsList() {
  const query = useQuery({ queryKey: ["recommendations"], queryFn: async () => { const response = await api.GET("/api/v1/recommendations"); if (!response.data) throw new Error("Recommendations are unavailable"); return response.data; } });
  return <><DashboardHeading eyebrow="Evidence review" title="Recommendations" description="Read-only queue for Phase 11. Acceptance and rejection workflow is implemented in Phase 12." />{query.isPending ? <LoadingState /> : query.isError ? <ErrorState message={apiMessage(query.error)} /> : query.data.items.length ? <ul className="recommendation-list">{query.data.items.map((item) => <li key={item.id}><div><span className={`priority priority-${item.priority}`}>{item.priority}</span><h2>{item.recommendation_type.replaceAll("_", " ")}</h2><p>{item.reason_codes.join(" · ")}</p></div><div><strong>{item.currency} {item.estimated_monthly_savings}</strong><span>estimated monthly savings</span><Link className="text-link" href={`/dashboard/recommendations/${item.id}`}>Inspect evidence →</Link></div></li>)}</ul> : <EmptyState title="No recommendations" message="Analyze eligible workload evidence to build the review queue." />}</>;
}
