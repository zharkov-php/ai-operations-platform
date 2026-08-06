"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { api, apiMessage } from "@/lib/api";
import { DashboardHeading, EmptyState, ErrorState, LoadingState } from "./states";

const schema = z.object({ workloadId: z.string().min(1), name: z.string().trim().min(1).max(200), source: z.string().trim().min(1).max(100), privacy: z.enum(["public", "internal", "confidential", "restricted"]) });
type FormValue = z.infer<typeof schema>;

export function Evaluations() {
  const client = useQueryClient();
  const [formError, setFormError] = useState("");
  const form = useForm<FormValue>({ defaultValues: { workloadId: "", name: "", source: "sanitized manual sample", privacy: "internal" } });
  const query = useQuery({ queryKey: ["evaluation-datasets"], queryFn: async () => { const response = await api.GET("/api/v1/evaluation-datasets"); if (!response.data) throw response.error; return response.data; } });
  const workloads = useQuery({ queryKey: ["workloads", "evaluation-form"], queryFn: async () => { const response = await api.GET("/api/v1/workloads", { params: { query: { limit: 100, offset: 0 } } }); if (!response.data) throw response.error; return response.data.items; } });
  const create = useMutation({ mutationFn: async (value: FormValue) => {
    const response = await api.POST("/api/v1/evaluation-datasets", { body: { workload_id: value.workloadId, name: value.name, description: "Two sanitized classification cases for deterministic candidate comparison.", source: value.source, privacy_classification: value.privacy, cases: [
      { sanitized_input: { ticket: "invoice question" }, expected_output: { label: "billing" }, validation_rules: [{ type: "classification_match", field: "label" }], metadata: { sanitized: true } },
      { sanitized_input: { ticket: "password reset" }, expected_output: { label: "access" }, validation_rules: [{ type: "classification_match", field: "label" }], metadata: { sanitized: true } },
    ] } });
    if (response.error) throw response.error;
    return response.data;
  }, onSuccess: async () => { setFormError(""); form.reset(); await client.invalidateQueries({ queryKey: ["evaluation-datasets"] }); }, onError: (error) => setFormError(apiMessage(error)) });
  const submit = form.handleSubmit((value) => { const parsed = schema.safeParse(value); if (parsed.success) create.mutate(parsed.data); });
  return <><DashboardHeading eyebrow="Quality evidence" title="Evaluations" description="Build sanitized ground-truth datasets and compare deterministic candidates before any production experiment." /><section className="dashboard-panel wide"><h2>Create dataset</h2><form className="evaluation-form" onSubmit={submit}><label>Workload<select {...form.register("workloadId", { required: true })}><option value="">Select a workload</option>{workloads.data?.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</select></label><label>Dataset name<input {...form.register("name", { required: true })} /></label><label>Source<input {...form.register("source", { required: true })} /></label><label>Privacy<select {...form.register("privacy")}><option value="public">Public</option><option value="internal">Internal</option><option value="confidential">Confidential</option><option value="restricted">Restricted</option></select></label><button className="button" disabled={create.isPending || workloads.isPending} type="submit">Create sanitized dataset</button>{formError && <span className="form-error" role="alert">{formError}</span>}</form><p>Inputs are examples generated in this browser flow and contain no raw production prompt content.</p></section><section><h2>Datasets</h2>{query.isPending ? <LoadingState /> : query.isError ? <ErrorState message={apiMessage(query.error)} /> : query.data.items.length ? <ul className="recommendation-list">{query.data.items.map((item) => <li key={item.id}><div><span className="priority">{item.privacy_classification}</span><h2>{item.name}</h2><p>{item.source}</p></div><div><strong>{item.case_count} cases</strong><span>ground truth</span><Link className="text-link" href={`/dashboard/evaluations/${item.id}`}>Set up evaluation →</Link></div></li>)}</ul> : <EmptyState title="No evaluation datasets" message="Create a sanitized dataset to measure candidate quality." />}</section></>;
}
