export type Article = {
  slug: string;
  title: string;
  summary: string;
  published: string;
  readingTime: string;
  sections: Array<{ heading: string; paragraphs: string[] }>;
};

export const articles: Article[] = [
  {
    slug: "deterministic-code-versus-llms",
    title: "When deterministic code should replace an LLM",
    summary: "A practical test for separating language problems from exact transformations and fixed business rules.",
    published: "2026-08-01",
    readingTime: "6 min read",
    sections: [
      { heading: "Start with the shape of the answer", paragraphs: ["An exact answer produced from fixed inputs is a strong signal that ordinary code may be the safer execution path. Arithmetic, schema validation, unit conversion, sorting, and fixed lookups do not gain reliability from probabilistic generation.", "Natural-language input can still require parsing. Separate that concern from the deterministic calculation so each part can be evaluated independently."] },
      { heading: "Use evidence, not intuition", paragraphs: ["Inspect repeated templates, expected output constraints, failure impact, and the presence of an authoritative rule set. A recommendation should state which observations support deterministic behavior and which unknowns remain."] },
      { heading: "Validate before changing production", paragraphs: ["Replacement code has implementation and maintenance cost. Run sanitized historical cases against both paths, compare outputs, and release through a guarded experiment. Estimated savings become verified only after measured production results satisfy quality guardrails."] },
    ],
  },
  {
    slug: "hosted-versus-local-models",
    title: "Hosted versus local models: an operational decision",
    summary: "Why privacy is only one input in a local-inference decision, alongside quality, throughput, and ownership cost.",
    published: "2026-07-24",
    readingTime: "7 min read",
    sections: [
      { heading: "Privacy is necessary but not sufficient", paragraphs: ["Local inference can change the data boundary, but sensitive data alone does not prove that a local model is suitable. Context limits, model capability, deployment controls, and staff access still matter."] },
      { heading: "Model the full operating cost", paragraphs: ["Include hardware amortization, electricity, maintenance, unused capacity, monitoring, upgrades, and incident response. Compare the resulting cost per request at realistic utilization rather than peak advertised throughput."] },
      { heading: "Evaluate the workload", paragraphs: ["Use a representative sanitized dataset to compare quality, latency, and failure modes. Treat the result as workload-specific; it is not a general benchmark of either deployment approach."] },
    ],
  },
  {
    slug: "measuring-llm-cost",
    title: "Measuring LLM cost without losing the audit trail",
    summary: "A cost model that preserves pricing dates, token categories, and the difference between observations and projections.",
    published: "2026-07-16",
    readingTime: "5 min read",
    sections: [
      { heading: "Price each call with the effective catalog", paragraphs: ["Store input, output, and cached-input token counts separately. Resolve the pricing period that was effective when the request occurred and retain enough provenance to explain the calculation later."] },
      { heading: "Keep monetary arithmetic exact", paragraphs: ["Use decimal-safe arithmetic from ingestion through aggregation. Binary floating-point errors become visible when millions of small call costs are summed or converted into budget decisions."] },
      { heading: "Label forecasts clearly", paragraphs: ["Observed cost, projected monthly cost, estimated candidate cost, and verified savings answer different questions. A dashboard should never collapse them into a single savings number."] },
    ],
  },
  {
    slug: "estimated-savings-require-evaluation",
    title: "Why estimated AI savings require evaluation",
    summary: "Cost reduction is a hypothesis until candidate quality and production guardrails are measured.",
    published: "2026-07-08",
    readingTime: "6 min read",
    sections: [
      { heading: "A recommendation is not a result", paragraphs: ["Historical volume and catalog pricing can estimate a financial opportunity. They cannot establish that a cheaper execution path preserves quality for the actual workload."] },
      { heading: "Build a representative evaluation", paragraphs: ["Sanitize real cases, preserve important edge conditions, define validation rules before running candidates, and retain failed cases for review. Avoid invented quality percentages when ground truth is absent."] },
      { heading: "Verify through controlled exposure", paragraphs: ["An experiment needs quality, latency, error-rate, and cost guardrails with an explicit rollback path. Savings are verified only after measured results pass those controls."] },
    ],
  },
  {
    slug: "reduce-unnecessary-context",
    title: "Reducing unnecessary LLM context safely",
    summary: "Deterministic signals for oversized prompts and a cautious path to testing smaller context windows.",
    published: "2026-06-30",
    readingTime: "5 min read",
    sections: [
      { heading: "Measure what can be measured", paragraphs: ["Track input-token distributions, repeated system text, stable reference blocks, retained conversation depth, and retrieved-document counts. These signals identify possible waste without pretending to measure semantic relevance."] },
      { heading: "Change one source at a time", paragraphs: ["Separate history trimming, reference deduplication, retrieval limits, and prompt-template changes. Independent candidates make regressions easier to diagnose and roll back."] },
      { heading: "Protect quality with cases and guardrails", paragraphs: ["Evaluate cases that depend on long-range context before experimentation. Monitor validation failures and task-specific quality signals, not token reduction alone."] },
    ],
  },
];

export const publicRoutes = [
  "/", "/features", "/pricing", "/docs", "/docs/getting-started", "/docs/ingestion-api",
  "/docs/cost-model", "/docs/recommendations", "/docs/security", "/blog", "/compare",
  "/compare/illustrative-frontier-vs-illustrative-small", "/security", "/privacy", "/terms",
  "/tools/llm-cost-calculator", "/tools/prompt-cost-estimator", "/tools/model-routing-advisor",
  ...articles.map((article) => `/blog/${article.slug}`),
] as const;

export const docs = [
  { slug: "getting-started", title: "Getting started", summary: "Understand the planned local workflow and the boundary between implemented and upcoming capabilities." },
  { slug: "ingestion-api", title: "Ingestion API", summary: "Send redacted execution metadata with scoped API keys and idempotent request identifiers." },
  { slug: "cost-model", title: "Cost model", summary: "Calculate observed hosted-model cost with effective-dated, decimal-safe pricing." },
  { slug: "recommendations", title: "Recommendations", summary: "Interpret deterministic reason codes, evidence, confidence inputs, and required next actions." },
  { slug: "security", title: "Security", summary: "Review tenant isolation, credential handling, redaction, retention, and safe logging controls." },
] as const;
