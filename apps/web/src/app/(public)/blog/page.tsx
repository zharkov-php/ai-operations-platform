import Link from "next/link";
import { PageIntro } from "@/components/public-shell";
import { articles } from "@/lib/content";
import { pageMetadata } from "@/lib/site";

export const metadata = pageMetadata("Engineering notes", "Practical writing about AI cost, model routing, evaluation, and operational safety.", "/blog");

export default function BlogPage() {
  return <div className="section"><PageIntro eyebrow="Engineering notes" title="Decisions behind reliable AI operations" description="Methodology-focused articles without invented benchmarks, customers, or savings claims." /><div className="article-list">{articles.map((article) => <article key={article.slug}><div><p className="card-label"><time dateTime={article.published}>{article.published}</time> · {article.readingTime}</p><h2><Link href={`/blog/${article.slug}`}>{article.title}</Link></h2><p>{article.summary}</p></div><Link className="arrow-link" aria-label={`Read ${article.title}`} href={`/blog/${article.slug}`}>→</Link></article>)}</div></div>;
}
