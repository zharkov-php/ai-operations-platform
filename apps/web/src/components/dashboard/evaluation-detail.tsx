"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { api, apiMessage } from "@/lib/api";
import { DashboardHeading, ErrorState, LoadingState } from "./states";

export function EvaluationDetail({ id }: { id: string }) {
  const client = useQueryClient();
  const [mode, setMode] = useState<"expected" | "echo">("expected");
  const [actionError, setActionError] = useState("");
  const query = useQuery({ queryKey: ["evaluation-dataset", id], queryFn: async () => { const response = await api.GET("/api/v1/evaluation-datasets/{datasetID}", { params: { path: { datasetID: id } } }); if (!response.data) throw response.error; return response.data; } });
  const run = useMutation({ mutationFn: async () => { const response = await api.POST("/api/v1/evaluation-runs", { body: { dataset_id: id, candidate_execution: { adapter: "deterministic_mock", mode, currency: "USD", latency_ms: 12, cost_per_case: "0.001", baseline_latency_ms: 80, baseline_cost_per_case: "0.02" } } }); if (response.error) throw response.error; return response.data; }, onSuccess: async () => { setActionError(""); await client.invalidateQueries({ queryKey: ["evaluation-dataset", id] }); }, onError: (error) => setActionError(apiMessage(error)) });
  if (query.isPending) return <LoadingState />;
  if (query.isError) return <ErrorState message={apiMessage(query.error)} />;
  const item = query.data;
  return <><DashboardHeading eyebrow={`${item.privacy_classification} · ${item.case_count} cases`} title={item.name} description={`${item.source} · sanitized evaluation inputs only`} /><div className="dashboard-grid"><section className="dashboard-panel"><h2>Candidate setup</h2><label className="evaluation-label">Deterministic candidate<select value={mode} onChange={(event) => setMode(event.target.value as "expected" | "echo")}><option value="expected">Expected-output fixture</option><option value="echo">Echo sanitized input</option></select></label><p>The expected fixture proves the validation harness; echo demonstrates partial or complete failure without provider credentials.</p><button type="button" className="button" disabled={run.isPending} onClick={() => run.mutate()}>Run evaluation</button>{actionError && <p className="form-error" role="alert">{actionError}</p>}</section><section className="dashboard-panel"><h2>Validation cases</h2><ul className="compact-list">{item.cases?.map((testCase, index) => <li key={testCase.id ?? index}><strong>Case {index + 1}</strong><span>{testCase.validation_rules.map((rule) => rule.type).join(", ")}</span></li>)}</ul></section><section className="dashboard-panel wide"><h2>Run progress and comparison</h2>{item.runs?.length ? <div className="table-scroll"><table className="data-table"><thead><tr><th>Status</th><th>Score</th><th>Passed</th><th>Failed</th><th>Latency</th><th>Candidate cost</th><th>Cost delta</th></tr></thead><tbody>{item.runs.map((result) => { const comparison = result.results.comparison as Record<string, string> | undefined; return <tr key={result.id}><td>{result.status}</td><td>{String(result.results.score ?? "—")}</td><td>{result.passed_cases}</td><td>{result.failed_cases}</td><td>{result.average_latency_ms} ms</td><td>{result.currency} {result.estimated_cost}</td><td>{comparison?.cost_delta ?? "—"}</td></tr>; })}</tbody></table></div> : <p>No runs yet. Configure and execute a deterministic candidate above.</p>}</section></div></>;
}
