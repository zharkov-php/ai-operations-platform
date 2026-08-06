export type IllustrativeModel = {
  id: string;
  provider: string;
  name: string;
  inputNanoUSDPerMillion: bigint;
  outputNanoUSDPerMillion: bigint;
  cachedInputNanoUSDPerMillion: bigint | null;
};

export const pricingEffectiveDate = "2026-08-01";

export const illustrativeModels: IllustrativeModel[] = [
  { id: "northstar-frontier", provider: "Northstar AI (fictional)", name: "Northstar Frontier", inputNanoUSDPerMillion: BigInt("4000000000"), outputNanoUSDPerMillion: BigInt("12000000000"), cachedInputNanoUSDPerMillion: BigInt("1000000000") },
  { id: "northstar-compact", provider: "Northstar AI (fictional)", name: "Northstar Compact", inputNanoUSDPerMillion: BigInt("600000000"), outputNanoUSDPerMillion: BigInt("1800000000"), cachedInputNanoUSDPerMillion: BigInt("150000000") },
  { id: "lantern-general", provider: "Lantern Models (fictional)", name: "Lantern General", inputNanoUSDPerMillion: BigInt("1250000000"), outputNanoUSDPerMillion: BigInt("3750000000"), cachedInputNanoUSDPerMillion: null },
];

export function callCostNanoUSD(model: IllustrativeModel, inputTokens: number, outputTokens: number, cachedInputTokens: number): bigint {
  const million = BigInt(1_000_000);
  const cachedRate = model.cachedInputNanoUSDPerMillion ?? model.inputNanoUSDPerMillion;
  return (BigInt(inputTokens) * model.inputNanoUSDPerMillion + BigInt(outputTokens) * model.outputNanoUSDPerMillion + BigInt(cachedInputTokens) * cachedRate) / million;
}

export function scaleNanoUSD(value: bigint, multiplier: number): bigint { return value * BigInt(multiplier); }

export function formatNanoUSD(value: bigint): string {
  const nanoUSDPerUSD = BigInt(1_000_000_000);
  const dollars = value / nanoUSDPerUSD;
  const fraction = (value % nanoUSDPerUSD).toString().padStart(9, "0").slice(0, 6).replace(/0+$/, "");
  return `$${dollars.toLocaleString("en-US")}${fraction ? `.${fraction}` : ""}`;
}

export function approximateTokens(text: string): number {
  const normalized = text.trim();
  if (!normalized) return 0;
  return Math.max(1, Math.ceil(new TextEncoder().encode(normalized).length / 4));
}

export type AdvisorInput = {
  understandsLanguage: boolean;
  exactAnswer: boolean;
  sensitive: boolean;
  quality: "standard" | "high" | "critical";
  failureImpact: "low" | "medium" | "high";
  volume: "low" | "medium" | "high";
  externalKnowledge: boolean;
  latencyCritical: boolean;
};

export function adviseRoute(input: AdvisorInput): { result: string; reason: string } {
  if (!input.understandsLanguage && input.exactAnswer) return { result: "Deterministic code", reason: "The task has one exact answer and does not require natural-language understanding." };
  if (input.quality === "critical" || input.failureImpact === "high" || input.externalKnowledge) return { result: "Frontier hosted model", reason: "Critical quality, high failure impact, or external knowledge favors the strongest evaluated hosted path." };
  if (input.sensitive && input.volume !== "low") return { result: "Local model candidate", reason: "Sensitivity and sustained volume justify evaluating local economics and workload quality; privacy alone is not sufficient." };
  if (input.volume === "high" && input.failureImpact === "low") return { result: "Smaller hosted model candidate", reason: "High volume and low failure impact make a smaller model worth evaluating before any routing change." };
  if (input.exactAnswer && input.volume !== "low") return { result: "Exact cache", reason: "Stable exact outputs at repeated volume should be checked for exact-key reuse." };
  return { result: "Manual architecture review", reason: "The answers do not provide enough evidence for a safe automatic educational recommendation." };
}
