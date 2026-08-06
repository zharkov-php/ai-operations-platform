"use client";

import type { components } from "@ai-operations/api-client";
import type { ColumnDef } from "@tanstack/react-table";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { api, apiMessage } from "@/lib/api";
import { DataTable } from "./data-table";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

type LLMCall = components["schemas"]["LLMCall"];
const columns: ColumnDef<LLMCall>[] = [
  { accessorKey: "created_at", header: "Timestamp", cell: ({ row }) => new Date(row.original.created_at).toLocaleString("en-US", { timeZone: "UTC" }) + " UTC" },
  { accessorKey: "provider", header: "Provider" }, { accessorKey: "model", header: "Model" }, { accessorKey: "estimated_cost", header: "Observed cost", cell: ({ row }) => `${row.original.currency} ${row.original.estimated_cost}` }, { accessorKey: "external_call_id", header: "External ID" },
];

export function Calls() {
  const [projectID, setProjectID] = useState("");
  const validProjectID = !projectID || /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(projectID);
  const query = useQuery({ queryKey: ["calls", projectID], enabled: validProjectID, queryFn: async () => { const response = await api.GET("/api/v1/llm-calls", { params: { query: { limit: 100, project_id: projectID || undefined } } }); if (!response.data) throw new Error("Calls are unavailable"); return response.data; } });
  return <><DashboardHeading eyebrow="Execution explorer" title="LLM calls" description="Redacted, organization-scoped execution records. Full prompts and responses are not displayed." /><label className="dashboard-filter">Project ID<input value={projectID} onChange={(event) => setProjectID(event.target.value.trim())} placeholder="Optional UUID filter" />{!validProjectID && <span className="form-error" role="alert">Enter a valid project UUID.</span>}</label>{!validProjectID ? null : query.isPending ? <LoadingState /> : query.isError ? <ErrorState message={apiMessage(query.error)} /> : query.data.items.length ? <DataTable columns={columns} data={query.data.items} label="LLM calls" /> : <EmptyState title="No calls found" message="Ingest execution metadata or change the project filter." />}</>;
}
