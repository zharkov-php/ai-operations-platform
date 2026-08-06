import Link from "next/link";
import { PageIntro } from "@/components/public-shell";
import { docs } from "@/lib/content";
import { pageMetadata } from "@/lib/site";

export const metadata = pageMetadata("Documentation", "Technical documentation for ingestion, costing, recommendations, and security.", "/docs");

export default function DocsPage() {
  return <div className="section"><PageIntro eyebrow="Documentation" title="Understand the system before operating it" description="These guides document current behavior and explicitly identify capabilities that remain planned." /><div className="card-grid three-up">{docs.map((item) => <article className="feature-card" key={item.slug}><h2>{item.title}</h2><p>{item.summary}</p><Link className="text-link" href={`/docs/${item.slug}`}>Read guide →</Link></article>)}</div></div>;
}
