import { PageIntro } from "@/components/public-shell";
import { pageMetadata } from "@/lib/site";

export const metadata = pageMetadata("Features", "Trace AI usage, evaluate optimization candidates, and verify controlled outcomes.", "/features");

const features = [
  ["Cost observability", "Attribute decimal-safe cost and token usage across organizations, projects, workloads, providers, and models."],
  ["Transparent recommendations", "Inspect rule versions, reason codes, measurable evidence, confidence inputs, risks, and required next actions."],
  ["Quality evaluation", "Use sanitized datasets and explicit validators to compare current and candidate execution paths."],
  ["Guarded experiments", "Control traffic, enforce quality and operational guardrails, and retain a rollback trail."],
  ["Budget governance", "Track projections, thresholds, alerts, acknowledgement, and ownership without exposing prompt content."],
  ["Local-model economics", "Compare amortization, utilization, throughput, maintenance, and workload constraints against hosted execution."],
];

export default function FeaturesPage() {
  return <div className="section"><PageIntro eyebrow="Platform capabilities" title="Operational evidence for every optimization decision" description="The platform connects usage records to recommendations, evaluations, experiments, and audit history without automatically switching production execution." /><section className="card-grid three-up" aria-label="Platform features">{features.map(([title, body]) => <article className="feature-card" key={title}><h2>{title}</h2><p>{body}</p></article>)}</section></div>;
}
