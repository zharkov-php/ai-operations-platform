import { describe, expect, it } from "vitest";
import { adviseRoute, approximateTokens, callCostNanoUSD, formatNanoUSD, illustrativeModels, scaleNanoUSD } from "./tools";

describe("public optimization tools", () => {
  it("calculates and scales illustrative pricing with integer monetary arithmetic", () => {
    const cost = callCostNanoUSD(illustrativeModels[0], 1_000, 250, 500);
    expect(formatNanoUSD(cost)).toBe("$0.0075");
    expect(formatNanoUSD(scaleNanoUSD(cost, 100_000))).toBe("$750");
  });

  it("uses the input rate when a model does not support cached pricing", () => {
    expect(callCostNanoUSD(illustrativeModels[2], 0, 0, 1_000)).toBe(BigInt(1_250_000));
  });

  it("documents a reproducible UTF-8 byte approximation", () => {
    expect(approximateTokens("")).toBe(0);
    expect(approximateTokens("12345678")).toBe(2);
    expect(approximateTokens("🙂")).toBe(1);
  });

  it.each([
    [{ understandsLanguage: false, exactAnswer: true, sensitive: false, quality: "standard", failureImpact: "low", volume: "high", externalKnowledge: false, latencyCritical: false }, "Deterministic code"],
    [{ understandsLanguage: true, exactAnswer: false, sensitive: false, quality: "critical", failureImpact: "medium", volume: "low", externalKnowledge: false, latencyCritical: false }, "Frontier hosted model"],
    [{ understandsLanguage: true, exactAnswer: false, sensitive: true, quality: "standard", failureImpact: "medium", volume: "high", externalKnowledge: false, latencyCritical: false }, "Local model candidate"],
    [{ understandsLanguage: true, exactAnswer: false, sensitive: false, quality: "standard", failureImpact: "low", volume: "high", externalKnowledge: false, latencyCritical: false }, "Smaller hosted model candidate"],
    [{ understandsLanguage: true, exactAnswer: true, sensitive: false, quality: "standard", failureImpact: "medium", volume: "medium", externalKnowledge: false, latencyCritical: false }, "Exact cache"],
  ] as const)("returns a transparent deterministic route", (input, expected) => expect(adviseRoute(input).result).toBe(expected));
});
