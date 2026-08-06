import { CostCalculator } from "@/components/tools/cost-calculator";
import { PageIntro } from "@/components/public-shell";
import { pageMetadata } from "@/lib/site";

export const metadata = pageMetadata("LLM cost calculator", "Estimate per-request, daily, monthly, and annual LLM cost from an illustrative pricing catalog.", "/tools/llm-cost-calculator");

export default async function CostCalculatorPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const query = await searchParams;
  const initial = Object.fromEntries(Object.entries(query).filter((entry): entry is [string, string] => typeof entry[1] === "string"));
  return <div className="section"><PageIntro eyebrow="Public tool" title="LLM cost calculator" description="Estimate workload cost with effective-dated fictional pricing and exact integer monetary arithmetic." /><CostCalculator initial={initial} /><aside className="notice"><strong>Methodology</strong><p>Per-call cost equals input tokens × input rate, output tokens × output rate, and cached-input tokens × the catalog’s cached rate. Monthly requests scale that value; daily uses a 30-day month and annual uses 12 months.</p></aside></div>;
}
