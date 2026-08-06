"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { api, apiMessage } from "@/lib/api";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

export function RecommendationsList() {
  const [priority, setPriority] = useState("");
  const [status, setStatus] = useState("");
  const query = useQuery({ queryKey: ["recommendations", priority, status], queryFn: async () => { const response = await api.GET("/api/v1/recommendations", { params: { query: { priority: priority || undefined, status: status || undefined } } }); if (!response.data) throw new Error("Recommendations are unavailable"); return response.data; } });
  return <><DashboardHeading eyebrow="Evidence review" title="Recommendations" description="Review evidence and authorize evaluation without switching production execution." /><div className="filter-row"><label>Priority<select value={priority} onChange={(event) => setPriority(event.target.value)}><option value="">All priorities</option><option value="high">High</option><option value="medium">Medium</option><option value="low">Low</option></select></label><label>Status<select value={status} onChange={(event) => setStatus(event.target.value)}><option value="">All statuses</option><option value="new">New</option><option value="accepted">Accepted</option><option value="rejected">Rejected</option><option value="testing">Testing</option><option value="verified">Verified</option></select></label></div>{query.isPending ? <LoadingState /> : query.isError ? <ErrorState message={apiMessage(query.error)} /> : query.data.items.length ? <ul className="recommendation-list">{query.data.items.map((item) => <li key={item.id}><div><span className={`priority priority-${item.priority}`}>{item.priority}</span><h2>{item.recommendation_type.replaceAll("_", " ")}</h2><p>{item.reason_codes.join(" · ")}</p></div><div><strong>{item.currency} {item.estimated_monthly_savings}</strong><span>estimated monthly savings</span><span>{item.status}</span><Link className="text-link" href={`/dashboard/recommendations/${item.id}`}>Inspect evidence →</Link></div></li>)}</ul> : <EmptyState title="No matching recommendations" message="Change filters or analyze eligible workload evidence." />}</>;
}
