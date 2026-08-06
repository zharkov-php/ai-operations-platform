import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  transpilePackages: ["@ai-operations/api-client"],
  async rewrites() {
    const apiURL = process.env.API_INTERNAL_URL ?? process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
    return [{ source: "/api/v1/:path*", destination: `${apiURL}/api/v1/:path*` }];
  },
};

export default nextConfig;
