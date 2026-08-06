"use client";

import { CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import type { components } from "@ai-operations/api-client";

type DailyPoint = components["schemas"]["DailyPoint"];

export function CostChart({ data, currency }: { data: DailyPoint[]; currency: string }) {
  const points = data.map((item) => ({ ...item, value: Number(item.cost) }));
  return <div className="chart" role="img" aria-label="Daily observed cost chart"><ResponsiveContainer width="100%" height="100%"><LineChart data={points} margin={{ top: 12, right: 12, left: 0, bottom: 0 }}><CartesianGrid stroke="#29425f" vertical={false} /><XAxis dataKey="date" stroke="#abc0d8" tickLine={false} /><YAxis stroke="#abc0d8" tickLine={false} /><Tooltip formatter={(value) => [`${currency} ${Number(value).toFixed(4)}`, "Observed cost"]} contentStyle={{ background: "#0d1b2e", border: "1px solid #29425f" }} /><Line type="monotone" dataKey="value" stroke="#65e6bd" strokeWidth={3} dot={false} /></LineChart></ResponsiveContainer></div>;
}
