import Link from "next/link";
import { pageMetadata, siteConfig } from "@/lib/site";

export const metadata = pageMetadata(
  "AI workload operations with traceable evidence",
  siteConfig.description,
  "/",
);

const decisions = [
  ["Observe", "Attribute tokens, cost, latency, errors, and retries to projects and workloads."],
  ["Recommend", "Explain deterministic optimization candidates with versioned rules and inspectable evidence."],
  ["Evaluate", "Compare candidates on sanitized cases before authorizing any operational change."],
  ["Verify", "Measure guarded experiments and keep verified savings separate from estimates."],
];

export default function HomePage() {
  const structuredData = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: siteConfig.name,
    applicationCategory: "DeveloperApplication",
    operatingSystem: "Web",
    description: siteConfig.description,
    url: siteConfig.url,
  };

  return (
    <>
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData).replace(/</g, "\\u003c") }} />
      <section className="hero section">
        <div>
          <p className="eyebrow">LLM operations · evidence first</p>
          <h1>Know where AI spend goes—and what can safely change.</h1>
          <p className="lede">Connect execution metadata to cost, quality risk, and controlled experiments. Every recommendation keeps its evidence and every savings claim keeps its status.</p>
          <div className="actions">
            <Link className="button" href="/docs/getting-started">Read the documentation</Link>
            <Link className="text-link" href="/features">Explore the decision workflow <span aria-hidden="true">→</span></Link>
          </div>
        </div>
        <div className="evidence-card" aria-label="Example recommendation evidence">
          <p className="card-label">Recommendation evidence</p>
          <h2>Replace fixed VAT calculation with code</h2>
          <dl className="evidence-list">
            <div><dt>Signal</dt><dd>Exact configured rules</dd></div>
            <div><dt>Confidence</dt><dd>High, evidence-based</dd></div>
            <div><dt>Next action</dt><dd>Run evaluation</dd></div>
            <div><dt>Savings</dt><dd>Estimated, not verified</dd></div>
          </dl>
        </div>
      </section>
      <section className="section section-muted" aria-labelledby="decision-heading">
        <p className="eyebrow">A controlled decision loop</p>
        <h2 id="decision-heading">From observation to verified outcome</h2>
        <div className="card-grid four-up">
          {decisions.map(([title, body], index) => <article className="number-card" key={title}><span>0{index + 1}</span><h3>{title}</h3><p>{body}</p></article>)}
        </div>
      </section>
      <section className="section split-section">
        <div><p className="eyebrow">Optimization without shortcuts</p><h2>Cost is one guardrail, not the objective.</h2></div>
        <div><p>AI Execution Advisor distinguishes observed cost, estimated opportunity, candidate evaluation, controlled experimentation, and verified savings.</p><p>It can also recommend keeping the current model when volume is low, evidence is weak, or failure impact is high.</p><Link className="text-link" href="/blog/estimated-savings-require-evaluation">Why estimates require evaluation →</Link></div>
      </section>
    </>
  );
}
