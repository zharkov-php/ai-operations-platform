"use client";

import { useState } from "react";
import { approximateTokens } from "@/lib/tools";

export function PromptEstimator() {
  const [text, setText] = useState("");
  const tokens = approximateTokens(text);
  const bytes = new TextEncoder().encode(text.trim()).length;
  return <div className="tool-grid"><div className="tool-form"><label htmlFor="prompt-text">Prompt text<textarea id="prompt-text" rows={12} value={text} onChange={(event) => setText(event.target.value)} placeholder="Paste text to estimate locally…" /></label><button className="button secondary-button" type="button" onClick={() => setText("")}>Clear text</button></div><section className="result-card" aria-live="polite" aria-label="Token estimate"><p className="card-label">Local approximation</p><h2>{tokens.toLocaleString("en-US")} <span>estimated tokens</span></h2><dl className="result-list"><div><dt>UTF-8 bytes</dt><dd>{bytes.toLocaleString("en-US")}</dd></div><div><dt>Method</dt><dd>ceil(bytes ÷ 4)</dd></div></dl><p>Text stays in this browser component and is not submitted to the API. The approximation is not a provider tokenizer and may differ substantially by language, formatting, and model.</p></section></div>;
}
