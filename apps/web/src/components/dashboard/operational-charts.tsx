"use client";

import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import type { components } from "@ai-operations/api-client";

type Tokens = components["schemas"]["TokenSummary"];
type Latency = components["schemas"]["LatencySummary"];
type Errors = components["schemas"]["ErrorSummary"];

function MetricBars({ data, unit, label }: { data: Array<{ name: string; value: number }>; unit: string; label: string }) {
  return <div className="chart small-chart" role="img" aria-label={label}><ResponsiveContainer width="100%" height="100%"><BarChart data={data}><CartesianGrid stroke="#29425f" vertical={false} /><XAxis dataKey="name" stroke="#abc0d8" /><YAxis stroke="#abc0d8" /><Tooltip formatter={(value) => [`${Number(value).toLocaleString()}${unit}`, "Value"]} contentStyle={{ background: "#0d1b2e", border: "1px solid #29425f" }} /><Bar dataKey="value" fill="#65e6bd" radius={[5, 5, 0, 0]} /></BarChart></ResponsiveContainer></div>;
}

export function TokenChart({ value }: { value: Tokens }) { return <MetricBars label="Token distribution chart" unit="" data={[{ name: "Input", value: value.input }, { name: "Output", value: value.output }, { name: "Cached", value: value.cached_input }]} />; }
export function ReliabilityChart({ latency, errors }: { latency: Latency; errors: Errors }) { return <MetricBars label="Latency and error chart" unit="" data={[{ name: "P50 ms", value: Number(latency.p50_ms) }, { name: "P95 ms", value: Number(latency.p95_ms) }, { name: "Errors", value: errors.errors }, { name: "Retries", value: errors.retried_calls }]} />; }
