import { RoutingAdvisor } from "@/components/tools/routing-advisor";
import { PageIntro } from "@/components/public-shell";
import { pageMetadata } from "@/lib/site";

export const metadata = pageMetadata("Model routing advisor", "Use a deterministic questionnaire to identify an execution candidate for evaluation.", "/tools/model-routing-advisor");
export default async function RoutingAdvisorPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) { const query = await searchParams; const initial = Object.fromEntries(Object.entries(query).filter((entry): entry is [string, string] => typeof entry[1] === "string")); return <div className="section"><PageIntro eyebrow="Public tool" title="Model routing advisor" description="Describe the task and receive an educational execution candidate with a transparent deterministic reason." /><RoutingAdvisor initial={initial} /><aside className="notice"><strong>Decision boundary</strong><p>The result does not inspect workload data, run a quality evaluation, estimate local hardware, or change routing. Manual review remains appropriate when evidence is incomplete.</p></aside></div>; }
