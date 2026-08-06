import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { PageIntro } from "@/components/public-shell";
import { docs } from "@/lib/content";
import { pageMetadata } from "@/lib/site";

const guideContent: Record<string, Array<{ heading: string; body: string }>> = {
  "getting-started": [
    {
      heading: "Current implementation",
      body: "The repository currently includes the Go service foundation, organization-scoped authentication, projects, workloads, effective-dated pricing, redacted LLM-call ingestion, analytics, and a deterministic recommendation engine. The public website is implemented in this phase.",
    },
    {
      heading: "Run the development stack",
      body: "Run make demo for the deterministic portfolio environment. Use the root Makefile for API, web, mobile, contract, and end-to-end validation.",
    },
    {
      heading: "Safe defaults",
      body: "Use only local development credentials from your untracked environment file. Raw prompts and responses are disabled by default; ingestion stores hashes and redacted previews.",
    },
  ],
  "ingestion-api": [
    {
      heading: "Authentication",
      body: "Create a hashed, organization-scoped API key with the ingest:calls scope. The plaintext value is shown only once and must never appear in application logs.",
    },
    {
      heading: "Idempotent ingestion",
      body: "POST execution metadata to /api/v1/llm-calls or the batch endpoint with an external call identifier. Repeating an accepted identifier returns the existing record instead of charging the call twice.",
    },
    {
      heading: "Data boundary",
      body: "Send token counts, timing, provider and model identifiers, redacted previews, and safe metadata. Full prompts and responses are not persisted by default.",
    },
  ],
  "cost-model": [
    {
      heading: "Hosted execution",
      body: "The observed call cost is the sum of input-token cost, output-token cost, and cached-input-token cost. Each category uses the pricing period effective at request time.",
    },
    {
      heading: "Exact arithmetic",
      body: "Monetary values use decimal-safe arithmetic. Pricing rows retain currency, unit size, effective dates, and a source note so calculations can be reviewed.",
    },
    {
      heading: "Statuses matter",
      body: "Observed cost, monthly projection, estimated candidate cost, estimated savings, and verified savings are distinct values. Only experiment results can produce verified savings.",
    },
  ],
  recommendations: [
    {
      heading: "Deterministic rules first",
      body: "The engine analyzes measurable workload signals for deterministic code, exact caching, smaller or local models, context and output reduction, structured output, batching, and keeping the current model.",
    },
    {
      heading: "Inspectable confidence",
      body: "Confidence uses call count, workload consistency, pricing completeness, classifications, and evaluation availability. The interface presents descriptive levels without inventing quality percentages.",
    },
    {
      heading: "No automatic production switch",
      body: "Accepting a recommendation authorizes evaluation or implementation work. It never directly changes production model routing.",
    },
  ],
  security: [
    {
      heading: "Tenant and credential boundaries",
      body: "Organization scoping is enforced in repositories and authorization middleware. API keys are hashed, access tokens are short lived, and refresh tokens rotate.",
    },
    {
      heading: "Redaction before persistence",
      body: "Known secret fields, authorization material, private-key markers, and optionally personal identifiers are redacted before previews are stored.",
    },
    {
      heading: "Operational logging",
      body: "Structured logs may carry request, organization, project, workload, and trace context. They must never contain credentials, prompt text, or model response content.",
    },
  ],
};

export function generateStaticParams() {
  return docs.map(({ slug }) => ({ slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const guide = docs.find((item) => item.slug === slug);
  return guide ? pageMetadata(guide.title, guide.summary, `/docs/${slug}`) : {};
}

export default async function GuidePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const guide = docs.find((item) => item.slug === slug);
  const sections = guideContent[slug];
  if (!guide || !sections) notFound();
  return (
    <article className="section article">
      <PageIntro
        eyebrow="Documentation"
        title={guide.title}
        description={guide.summary}
      />
      {sections.map((section) => (
        <section key={section.heading}>
          <h2>{section.heading}</h2>
          <p>{section.body}</p>
        </section>
      ))}
      <Link className="text-link" href="/docs">
        ← All documentation
      </Link>
    </article>
  );
}
