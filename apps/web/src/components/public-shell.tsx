import Link from "next/link";

const navigation = [
  ["Features", "/features"],
  ["Pricing", "/pricing"],
  ["Docs", "/docs"],
  ["Blog", "/blog"],
  ["Compare", "/compare"],
  ["Tools", "/tools/llm-cost-calculator"],
] as const;

export function PublicHeader() {
  return (
    <header className="site-header">
      <Link className="brand" href="/" aria-label="AI Execution Advisor home">
        <span className="brand-mark" aria-hidden="true">EA</span>
        <span>AI Execution Advisor</span>
      </Link>
      <nav aria-label="Primary navigation">
        <ul className="nav-list">
          {navigation.map(([label, href]) => <li key={href}><Link href={href}>{label}</Link></li>)}
        </ul>
      </nav>
      <Link className="button button-small" href="/dashboard">Open dashboard</Link>
    </header>
  );
}

export function PublicFooter() {
  return (
    <footer className="site-footer">
      <div>
        <strong>AI Execution Advisor</strong>
        <p>Traceable operations and controlled optimization for AI workloads.</p>
      </div>
      <nav aria-label="Footer navigation">
        <Link href="/docs/security">Documentation</Link>
        <Link href="/tools/llm-cost-calculator">Cost calculator</Link>
        <Link href="/tools/prompt-cost-estimator">Prompt estimator</Link>
        <Link href="/tools/model-routing-advisor">Routing advisor</Link>
        <Link href="/security">Security</Link>
        <Link href="/privacy">Privacy</Link>
        <Link href="/terms">Terms</Link>
      </nav>
      <p className="footer-note">Open-source portfolio software. No production claims or provider benchmarks.</p>
    </footer>
  );
}

export function PageIntro({ eyebrow, title, description }: { eyebrow: string; title: string; description: string }) {
  return (
    <section className="page-intro">
      <p className="eyebrow">{eyebrow}</p>
      <h1>{title}</h1>
      <p className="lede">{description}</p>
    </section>
  );
}
