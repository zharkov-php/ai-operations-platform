import type { Metadata } from "next";

const configuredURL = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";

export const siteConfig = {
  name: "AI Execution Advisor",
  description:
    "Evidence-based operations, cost analysis, evaluation, and controlled optimization for AI workloads.",
  url: configuredURL.replace(/\/$/, ""),
} as const;

export function pageMetadata(
  title: string,
  description: string,
  path: string,
): Metadata {
  return {
    title,
    description,
    alternates: { canonical: path },
    openGraph: { title, description, url: path },
    twitter: { title, description },
  };
}
