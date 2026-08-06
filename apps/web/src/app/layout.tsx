import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "AI Execution Advisor",
  description: "Evidence-based AI workload operations and optimization.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
