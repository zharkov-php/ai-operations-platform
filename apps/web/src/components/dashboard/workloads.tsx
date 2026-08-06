"use client";

import type { components } from "@ai-operations/api-client";
import type { ColumnDef } from "@tanstack/react-table";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { api, apiMessage } from "@/lib/api";
import { DataTable } from "./data-table";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

type Workload = components["schemas"]["Workload"];
const columns: ColumnDef<Workload>[] = [
  { accessorKey: "name", header: "Workload", cell: ({ row }) => <Link href={`/dashboard/workloads/${row.original.id}`}>{row.original.name}</Link> }, { accessorKey: "type", header: "Type" }, { accessorKey: "owner", header: "Owner" }, { accessorKey: "criticality", header: "Criticality" }, { accessorKey: "privacy_classification", header: "Privacy" },
];

export function Workloads() {
  const [criticality, setCriticality] = useState("");
  const query = useQuery({ queryKey: ["workloads"], queryFn: async () => { const response = await api.GET("/api/v1/workloads", { params: { query: { limit: 100 } } }); if (!response.data) throw new Error("Workloads are unavailable"); return response.data; } });
  const items = query.data?.items.filter((item) => !criticality || item.criticality === criticality) ?? [];
  return <><DashboardHeading eyebrow="Workload inventory" title="AI workloads" description="Operational features, agents, workflows, jobs, and internal tools registered in this organization." /><label className="dashboard-filter">Criticality<select value={criticality} onChange={(event) => setCriticality(event.target.value)}><option value="">All levels</option><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option><option value="critical">Critical</option></select></label>{query.isPending ? <LoadingState /> : query.isError ? <ErrorState message={apiMessage(query.error)} /> : items.length ? <DataTable columns={columns} data={items} label="Workloads" /> : <EmptyState title="No matching workloads" message="Change the filter or register a workload." />}</>;
}
