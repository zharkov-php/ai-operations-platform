import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { PageIntro } from "@/components/public-shell";
import { pageMetadata } from "@/lib/site";

const comparison = {
  slug: "illustrative-frontier-vs-illustrative-small",
  title: "Illustrative Frontier vs Illustrative Small",
  description: "A qualitative, fictional comparison showing the evidence required for a routing decision.",
};

export function generateStaticParams() { return [{ comparison: comparison.slug }]; }
export async function generateMetadata({ params }: { params: Promise<{ comparison: string }> }): Promise<Metadata> { const value = await params; return value.comparison === comparison.slug ? pageMetadata(comparison.title, comparison.description, `/compare/${comparison.slug}`) : {}; }

export default async function ComparisonPage({ params }: { params: Promise<{ comparison: string }> }) {
  if ((await params).comparison !== comparison.slug) notFound();
  return <div className="section"><PageIntro eyebrow="Fictional catalog comparison" title={comparison.title} description={comparison.description} /><div className="comparison-table" role="region" aria-label="Model comparison" tabIndex={0}><table><thead><tr><th scope="col">Decision input</th><th scope="col">Illustrative Frontier</th><th scope="col">Illustrative Small</th></tr></thead><tbody><tr><th scope="row">Intended evaluation</th><td>Complex reasoning and critical-quality cases</td><td>Repetitive classification and extraction cases</td></tr><tr><th scope="row">Cost status</th><td>Illustrative catalog estimate</td><td>Illustrative catalog estimate</td></tr><tr><th scope="row">Quality status</th><td>Not established for your workload</td><td>Requires evaluation before approval</td></tr><tr><th scope="row">Routing decision</th><td colSpan={2}>Use sanitized ground-truth cases, capability checks, and experiment guardrails</td></tr></tbody></table></div><aside className="notice"><strong>Not a benchmark</strong><p>Both names and all associated demonstration pricing are fictional. This page explains comparison methodology only.</p></aside><Link className="text-link" href="/compare">← Comparison framework</Link></div>;
}
