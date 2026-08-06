import { NextRequest } from "next/server";
import { describe, expect, it } from "vitest";
import { proxy } from "./proxy";

describe("dashboard proxy", () => {
  it("redirects a missing session to sign in with a safe return path", () => {
    const response = proxy(new NextRequest("https://advisor.test/dashboard/projects"));
    expect(response.status).toBe(307);
    expect(response.headers.get("location")).toBe("https://advisor.test/sign-in?next=%2Fdashboard%2Fprojects");
  });

  it("allows a request carrying the HTTP-only access cookie", () => {
    const request = new NextRequest("https://advisor.test/dashboard", { headers: { cookie: "access_token=opaque" } });
    expect(proxy(request).headers.get("x-middleware-next")).toBe("1");
  });
});
