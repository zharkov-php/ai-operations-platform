import type { Metadata } from "next";
import Link from "next/link";
import { Suspense } from "react";
import { SignInForm } from "@/components/auth/sign-in-form";

export const metadata: Metadata = { title: "Sign in", robots: { index: false, follow: false } };
export default function SignInPage() { return <main className="auth-page"><section><Link className="brand" href="/"><span className="brand-mark" aria-hidden="true">EA</span><span>AI Execution Advisor</span></Link><p className="eyebrow">Private platform</p><h1>Sign in to inspect operational evidence.</h1><p>Use an organization account created by the seed or authentication API. Credentials are never embedded in this page.</p></section><section className="auth-card"><h2>Organization account</h2><Suspense fallback={<p>Loading sign-in form…</p>}><SignInForm /></Suspense><p>API sessions use HTTP-only cookies. Contact your organization owner if access has not been provisioned.</p></section></main>; }
