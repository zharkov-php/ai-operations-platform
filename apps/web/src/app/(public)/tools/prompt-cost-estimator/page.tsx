import { PromptEstimator } from "@/components/tools/prompt-estimator";
import { PageIntro } from "@/components/public-shell";
import { pageMetadata } from "@/lib/site";

export const metadata = pageMetadata("Prompt token estimator", "Estimate prompt tokens locally in the browser without transmitting pasted text.", "/tools/prompt-cost-estimator");
export default function PromptEstimatorPage() { return <div className="section"><PageIntro eyebrow="Public tool" title="Prompt token estimator" description="Paste text for a documented browser-local approximation. Nothing is submitted to the platform API." /><PromptEstimator /><aside className="notice"><strong>Approximation boundary</strong><p>The initial method divides UTF-8 bytes by four and rounds upward. It is reproducible but not model-specific; use the intended provider tokenizer for production pricing.</p></aside></div>; }
