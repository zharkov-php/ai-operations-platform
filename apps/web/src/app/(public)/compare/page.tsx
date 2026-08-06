import Link from "next/link";
import { PageIntro } from "@/components/public-shell";
import { pageMetadata } from "@/lib/site";

export const metadata = pageMetadata("Model comparisons", "A workload-first framework for comparing model execution candidates.", "/compare");

export default function ComparePage() {
  return <div className="section"><PageIntro eyebrow="Compare execution paths" title="Compare models in the context of a workload" description="Useful comparisons include quality requirements, failure impact, privacy, context, latency, capabilities, operating cost, and evaluation evidence." /><section className="split-section comparison-intro"><div><h2>Illustrative example</h2><p>The demonstration catalog uses fictional model names and prices. This keeps the decision workflow testable without presenting stale provider claims.</p></div><div><Link className="button" href="/compare/illustrative-frontier-vs-illustrative-small">View qualitative comparison</Link></div></section><aside className="notice"><strong>No universal winner</strong><p>A less expensive model is a candidate only until a representative evaluation establishes acceptable workload quality.</p></aside></div>;
}
