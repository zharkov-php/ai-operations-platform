"use client";

import { useForm, useWatch } from "react-hook-form";
import { adviseRoute, type AdvisorInput } from "@/lib/tools";

const yesNo = (register: ReturnType<typeof useForm<AdvisorInput>>["register"], name: "understandsLanguage" | "exactAnswer" | "sensitive" | "externalKnowledge" | "latencyCritical") => <select {...register(name, { setValueAs: (value) => value === "true" })}><option value="true">Yes</option><option value="false">No</option></select>;

const defaults: AdvisorInput = { understandsLanguage: true, exactAnswer: false, sensitive: false, quality: "standard", failureImpact: "medium", volume: "medium", externalKnowledge: false, latencyCritical: false };

export function RoutingAdvisor({ initial = {} }: { initial?: Partial<Record<keyof AdvisorInput, string>> }) {
  const booleanValue = (key: keyof AdvisorInput, fallback: boolean) => initial[key] === "true" ? true : initial[key] === "false" ? false : fallback;
  const quality = ["standard", "high", "critical"].includes(initial.quality ?? "") ? initial.quality as AdvisorInput["quality"] : defaults.quality;
  const failureImpact = ["low", "medium", "high"].includes(initial.failureImpact ?? "") ? initial.failureImpact as AdvisorInput["failureImpact"] : defaults.failureImpact;
  const volume = ["low", "medium", "high"].includes(initial.volume ?? "") ? initial.volume as AdvisorInput["volume"] : defaults.volume;
  const form = useForm<AdvisorInput>({ defaultValues: { understandsLanguage: booleanValue("understandsLanguage", defaults.understandsLanguage), exactAnswer: booleanValue("exactAnswer", defaults.exactAnswer), sensitive: booleanValue("sensitive", defaults.sensitive), quality, failureImpact, volume, externalKnowledge: booleanValue("externalKnowledge", defaults.externalKnowledge), latencyCritical: booleanValue("latencyCritical", defaults.latencyCritical) } });
  const values = useWatch({ control: form.control }) as AdvisorInput;
  const advice = adviseRoute(values);
  function share() { const query = new URLSearchParams(Object.entries(values).map(([key, value]) => [key, String(value)])); window.history.replaceState(null, "", `${window.location.pathname}?${query}`); }
  return <div className="tool-grid"><form className="tool-form" onSubmit={(event) => { event.preventDefault(); share(); }}><label>Does the task require natural-language understanding?{yesNo(form.register, "understandsLanguage")}</label><label>Is there one exact correct answer?{yesNo(form.register, "exactAnswer")}</label><label>Is the data sensitive?{yesNo(form.register, "sensitive")}</label><label>Quality requirement<select {...form.register("quality")}><option value="standard">Standard</option><option value="high">High</option><option value="critical">Critical</option></select></label><label>Failure impact<select {...form.register("failureImpact")}><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option></select></label><label>Request volume<select {...form.register("volume")}><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option></select></label><label>Requires external knowledge?{yesNo(form.register, "externalKnowledge")}</label><label>Is latency critical?{yesNo(form.register, "latencyCritical")}</label><button className="button" type="submit">Update shareable URL</button></form><section className="result-card" aria-live="polite" aria-label="Routing advice"><p className="card-label">Educational recommendation</p><h2>{advice.result}</h2><p>{advice.reason}</p><p>This is a deterministic questionnaire result, not an automated production routing decision. Validate capability, quality, economics, and guardrails before implementation.</p></section></div>;
}
