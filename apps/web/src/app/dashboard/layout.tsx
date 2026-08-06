import type { Metadata } from "next";
import { DashboardProviders } from "@/components/dashboard/providers";
import { DashboardShell } from "@/components/dashboard/shell";

export const metadata: Metadata = { title: "Dashboard", robots: { index: false, follow: false, nocache: true } };
export default function DashboardLayout({ children }: { children: React.ReactNode }) { return <DashboardProviders><DashboardShell>{children}</DashboardShell></DashboardProviders>; }
