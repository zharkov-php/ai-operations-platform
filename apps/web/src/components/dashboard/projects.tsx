"use client";

import type { components } from "@ai-operations/api-client";
import type { ColumnDef } from "@tanstack/react-table";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { api, apiMessage } from "@/lib/api";
import { DataTable } from "./data-table";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

type Project = components["schemas"]["Project"];
const columns: ColumnDef<Project>[] = [
  { accessorKey: "name", header: "Project", cell: ({ row }) => <Link href={`/dashboard/projects/${row.original.id}`}>{row.original.name}</Link> },
  { accessorKey: "environment", header: "Environment" }, { accessorKey: "monthly_budget", header: "Monthly budget", cell: ({ row }) => `${row.original.currency} ${row.original.monthly_budget}` }, { accessorKey: "status", header: "Status" },
];

export function Projects() {
  const [environment, setEnvironment] = useState("");
  const query = useQuery({ queryKey: ["projects", environment], queryFn: async () => { const response = await api.GET("/api/v1/projects", { params: { query: { limit: 100, environment: environment || undefined } } }); if (!response.data) throw new Error("Projects are unavailable"); return response.data; } });
  return <><DashboardHeading eyebrow="Portfolio" title="Projects" description="Filter the organization-scoped project inventory and inspect its workloads." /><label className="dashboard-filter">Environment<select value={environment} onChange={(event) => setEnvironment(event.target.value)}><option value="">All environments</option><option value="development">Development</option><option value="staging">Staging</option><option value="production">Production</option></select></label>{query.isPending ? <LoadingState /> : query.isError ? <ErrorState message={apiMessage(query.error)} /> : query.data.items.length ? <DataTable columns={columns} data={query.data.items} label="Projects" /> : <EmptyState title="No matching projects" message="Change the filter or create a project through the API." />}</>;
}
