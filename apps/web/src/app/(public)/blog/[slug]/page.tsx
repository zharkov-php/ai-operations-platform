import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { PageIntro } from "@/components/public-shell";
import { articles } from "@/lib/content";
import { pageMetadata, siteConfig } from "@/lib/site";

export function generateStaticParams() { return articles.map(({ slug }) => ({ slug })); }

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  const article = articles.find((item) => item.slug === slug);
  return article ? pageMetadata(article.title, article.summary, `/blog/${slug}`) : {};
}

export default async function ArticlePage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const article = articles.find((item) => item.slug === slug);
  if (!article) notFound();
  const structuredData = { "@context": "https://schema.org", "@type": "Article", headline: article.title, description: article.summary, datePublished: article.published, author: { "@type": "Organization", name: siteConfig.name }, mainEntityOfPage: `${siteConfig.url}/blog/${slug}` };
  return <article className="section article"><script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData).replace(/</g, "\\u003c") }} /><PageIntro eyebrow={`${article.published} · ${article.readingTime}`} title={article.title} description={article.summary} />{article.sections.map((section) => <section key={section.heading}><h2>{section.heading}</h2>{section.paragraphs.map((paragraph) => <p key={paragraph}>{paragraph}</p>)}</section>)}<aside className="notice"><strong>Methodology note</strong><p>Examples describe decision methods, not provider benchmarks or guarantees for a production workload.</p></aside><Link className="text-link" href="/blog">← All engineering notes</Link></article>;
}
