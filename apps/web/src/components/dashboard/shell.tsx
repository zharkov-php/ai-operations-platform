"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { api } from "@/lib/api";

const links = [
  ["Overview", "/dashboard"], ["Projects", "/dashboard/projects"], ["Workloads", "/dashboard/workloads"],
  ["Calls", "/dashboard/calls"], ["Recommendations", "/dashboard/recommendations"], ["Evaluations", "/dashboard/evaluations"],
] as const;

export function DashboardShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const me = useQuery({ queryKey: ["me"], queryFn: async () => { const result = await api.GET("/api/v1/me"); if (result.error) throw result.error; return result.data; } });
  async function signOut() { await api.POST("/api/v1/auth/logout"); router.replace("/sign-in"); router.refresh(); }
  useEffect(() => { if (me.isError) router.replace(`/sign-in?next=${encodeURIComponent(pathname)}`); }, [me.isError, pathname, router]);
  if (me.isPending) return <main className="dashboard-loading" aria-live="polite">Verifying session…</main>;
  if (me.isError || !me.data) return <main className="dashboard-loading" aria-live="polite">Redirecting to sign in…</main>;
  return <div className="dashboard-layout"><aside className="dashboard-sidebar"><Link className="brand" href="/"><span className="brand-mark" aria-hidden="true">EA</span><span>Execution Advisor</span></Link><nav aria-label="Dashboard navigation">{links.map(([label, href]) => <Link aria-current={pathname === href ? "page" : undefined} href={href} key={href}>{label}</Link>)}</nav><div className="account"><span>{me.data.role}</span><small>Organization scoped</small><button type="button" onClick={signOut}>Sign out</button></div></aside><div className="dashboard-content">{children}</div></div>;
}
