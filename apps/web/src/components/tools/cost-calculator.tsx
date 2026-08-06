"use client";

import { useForm, useWatch } from "react-hook-form";
import { z } from "zod";
import { callCostNanoUSD, formatNanoUSD, illustrativeModels, pricingEffectiveDate, scaleNanoUSD } from "@/lib/tools";

const schema = z.object({ model: z.string(), requests: z.coerce.number().int().min(1).max(1_000_000_000), inputTokens: z.coerce.number().int().min(0).max(10_000_000), outputTokens: z.coerce.number().int().min(0).max(10_000_000), cachedInputTokens: z.coerce.number().int().min(0).max(10_000_000) });
type Values = z.infer<typeof schema>;

export function CostCalculator({ initial }: { initial: Partial<Record<keyof Values, string>> }) {
  const { control, register, formState: { errors } } = useForm<Values>({ defaultValues: { model: initial.model ?? illustrativeModels[0].id, requests: Number(initial.requests ?? 100000), inputTokens: Number(initial.inputTokens ?? 800), outputTokens: Number(initial.outputTokens ?? 200), cachedInputTokens: Number(initial.cachedInputTokens ?? 0) }, mode: "onChange" });
  const raw = useWatch({ control });
  const parsed = schema.safeParse(raw);
  const result = (() => {
    if (!parsed.success) return null;
    const model = illustrativeModels.find((item) => item.id === parsed.data.model) ?? illustrativeModels[0];
    const perCall = callCostNanoUSD(model, parsed.data.inputTokens, parsed.data.outputTokens, parsed.data.cachedInputTokens);
    const monthly = scaleNanoUSD(perCall, parsed.data.requests);
    return { model, perCall, daily: monthly / BigInt(30), monthly, annual: monthly * BigInt(12) };
  })();

  function share() {
    if (!parsed.success) return;
    const query = new URLSearchParams(Object.entries(parsed.data).map(([key, value]) => [key, String(value)]));
    window.history.replaceState(null, "", `${window.location.pathname}?${query}`);
  }

  return <div className="tool-grid"><form className="tool-form" onSubmit={(event) => { event.preventDefault(); share(); }} noValidate><label>Model<select {...register("model")}>{illustrativeModels.map((model) => <option value={model.id} key={model.id}>{model.provider} — {model.name}</option>)}</select></label><label>Monthly requests<input type="number" inputMode="numeric" {...register("requests")} />{errors.requests && <span role="alert">Enter 1 to 1,000,000,000 requests.</span>}</label><label>Average input tokens<input type="number" inputMode="numeric" {...register("inputTokens")} />{errors.inputTokens && <span role="alert">Enter a valid non-negative token count.</span>}</label><label>Average output tokens<input type="number" inputMode="numeric" {...register("outputTokens")} />{errors.outputTokens && <span role="alert">Enter a valid non-negative token count.</span>}</label><label>Cached input tokens<input type="number" inputMode="numeric" {...register("cachedInputTokens")} />{errors.cachedInputTokens && <span role="alert">Enter a valid non-negative token count.</span>}</label><button className="button" type="submit">Update shareable URL</button></form><section className="result-card" aria-live="polite" aria-label="Cost estimate">{result ? <><p className="card-label">Illustrative estimate</p><h2>{formatNanoUSD(result.monthly)} <span>/ month</span></h2><dl className="result-list"><div><dt>Per request</dt><dd>{formatNanoUSD(result.perCall)}</dd></div><div><dt>Daily</dt><dd>{formatNanoUSD(result.daily)}</dd></div><div><dt>Annual</dt><dd>{formatNanoUSD(result.annual)}</dd></div></dl><p>Catalog effective {pricingEffectiveDate}. Fictional demonstration pricing; excludes provider discounts, taxes, retries, and infrastructure.</p></> : <p role="alert">Correct the inputs to calculate an estimate.</p>}</section></div>;
}
