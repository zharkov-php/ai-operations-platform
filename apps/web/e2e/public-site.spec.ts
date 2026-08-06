import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { publicRoutes } from "../src/lib/content";

for (const route of publicRoutes) {
  test(`${route} renders with indexable metadata`, async ({ page }) => {
    const response = await page.goto(route);
    expect(response?.status()).toBe(200);
    await expect(page.locator("h1")).toBeVisible();
    await expect(page.locator('link[rel="canonical"]')).toHaveAttribute("href", new RegExp(`${route === "/" ? "/?$" : `${route}$`}`));
    await expect(page.locator('meta[name="robots"]')).toHaveCount(0);
  });
}

test("homepage has accessible primary navigation and no serious axe violations", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("navigation", { name: "Primary navigation" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Skip to content" })).toBeAttached();
  const results = await new AxeBuilder({ page }).analyze();
  expect(results.violations.filter((violation) => ["serious", "critical"].includes(violation.impact ?? ""))).toEqual([]);
});

test("all public internal links resolve", async ({ page, request }) => {
  await page.goto("/");
  const pending = new Set<string>(["/"]);
  const visited = new Set<string>();
  while (pending.size > 0) {
    const path = pending.values().next().value as string;
    pending.delete(path);
    if (visited.has(path) || path.startsWith("/dashboard")) continue;
    visited.add(path);
    const response = await request.get(path);
    expect(response.status(), `Expected ${path} to resolve`).toBeLessThan(400);
    await page.goto(path);
    const hrefs = await page.locator('a[href^="/"]').evaluateAll((links) => links.map((link) => link.getAttribute("href")).filter((href): href is string => Boolean(href)));
    for (const href of hrefs) pending.add(new URL(href, "http://localhost").pathname);
  }
  expect(visited.size).toBeGreaterThanOrEqual(publicRoutes.length);
});

test("sitemap lists public pages and excludes the dashboard", async ({ request }) => {
  const response = await request.get("/sitemap.xml");
  expect(response.status()).toBe(200);
  const xml = await response.text();
  for (const route of publicRoutes) expect(xml).toContain(route === "/" ? "http://localhost:3000</loc>" : `http://localhost:3000${route}</loc>`);
  expect(xml).not.toContain("/dashboard");
});

test("robots allows public pages and protects private routes", async ({ request }) => {
  const response = await request.get("/robots.txt");
  expect(response.status()).toBe(200);
  const body = await response.text();
  expect(body).toContain("Allow: /");
  expect(body).toContain("Disallow: /dashboard/");
  expect(body).toContain("Disallow: /api/");
  expect(body).toContain("Sitemap: http://localhost:3000/sitemap.xml");
});

test("unauthenticated dashboard redirects to a noindex sign-in page", async ({ page }) => {
  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/sign-in\?next=%2Fdashboard/);
  await expect(page.locator('meta[name="robots"]')).toHaveAttribute("content", /noindex/);
  await expect(page.getByRole("heading", { level: 1 })).toContainText("Sign in");
});

test("cost calculator updates exact estimates and shareable URL state", async ({ page }) => {
  await page.goto("/tools/llm-cost-calculator");
  await page.getByLabel("Monthly requests").fill("100000");
  await page.getByLabel("Average input tokens").fill("1000");
  await page.getByLabel("Average output tokens").fill("250");
  await page.getByLabel("Cached input tokens").fill("500");
  await expect(page.getByLabel("Cost estimate")).toContainText("$750 / month");
  await page.getByRole("button", { name: "Update shareable URL" }).click();
  await expect(page).toHaveURL(/requests=100000/);
  await page.reload();
  await expect(page.getByLabel("Cost estimate")).toContainText("$750 / month");
});

test("prompt estimator keeps pasted content browser-local", async ({ page }) => {
  const requests: string[] = [];
  page.on("request", (request) => requests.push(request.url()));
  await page.goto("/tools/prompt-cost-estimator");
  requests.length = 0;
  await page.getByLabel("Prompt text").fill("private-example-12345678");
  await expect(page.getByLabel("Token estimate")).toContainText("6 estimated tokens");
  await page.waitForTimeout(150);
  expect(requests.some((url) => url.includes("private-example"))).toBe(false);
  expect(requests.some((url) => new URL(url).pathname.startsWith("/api/"))).toBe(false);
  expect(page.url()).not.toContain("private-example");
});

test("routing advisor explains a deterministic code candidate", async ({ page }) => {
  await page.goto("/tools/model-routing-advisor");
  await page.getByLabel("Does the task require natural-language understanding?").selectOption("false");
  await page.getByLabel("Is there one exact correct answer?").selectOption("true");
  await expect(page.getByLabel("Routing advice")).toContainText("Deterministic code");
  await expect(page.getByLabel("Routing advice")).toContainText("one exact answer");
  await page.getByRole("button", { name: "Update shareable URL" }).click();
  await expect(page).toHaveURL(/understandsLanguage=false/);
});
