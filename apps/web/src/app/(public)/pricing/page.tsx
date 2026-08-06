import Link from "next/link";
import { PageIntro } from "@/components/public-shell";
import { pageMetadata } from "@/lib/site";

export const metadata = pageMetadata("Pricing", "Open-source project status and illustrative model-pricing policy.", "/pricing");

export default function PricingPage() {
  return <div className="section narrow"><PageIntro eyebrow="Pricing" title="Open source, with pricing data you can audit" description="AI Execution Advisor is currently a portfolio implementation. There is no hosted commercial plan and no purchase flow." /><section className="pricing-card"><p className="card-label">Repository edition</p><h2>Self-hosted</h2><p className="price">$0 <span>software license fee</span></p><ul className="check-list"><li>MIT-licensed source</li><li>PostgreSQL and Redis deployment</li><li>Illustrative demo catalog</li><li>Your infrastructure and model-provider costs remain separate</li></ul><Link className="button" href="/docs/getting-started">Review setup status</Link></section><aside className="notice"><strong>Pricing catalog policy</strong><p>Demonstration model prices are fictional and explicitly marked illustrative with an effective date. They are not claims about current provider prices.</p></aside></div>;
}
