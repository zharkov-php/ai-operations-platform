import AxeBuilder from "@axe-core/playwright";
import { expect, type Page, test } from "@playwright/test";

const projectID = "11111111-1111-4111-8111-111111111111";
const workloadID = "22222222-2222-4222-8222-222222222222";
const recommendationID = "33333333-3333-4333-8333-333333333333";
const now = "2026-08-06T10:00:00Z";

const project = { id: projectID, organization_id: "44444444-4444-4444-8444-444444444444", name: "Demo Operations", slug: "demo-operations", environment: "production", monthly_budget: "2000", currency: "USD", status: "active", created_at: now, updated_at: now };
const workload = { id: workloadID, project_id: projectID, name: "Ticket Classifier", slug: "ticket-classifier", type: "feature", owner: "Support Engineering", criticality: "medium", quality_requirement: "standard", privacy_classification: "internal", status: "active", created_at: now, updated_at: now };
const overview = { observed_cost: "126.50", previous_period_cost: "110", estimated_monthly_projection: "180.25", currency: "USD", call_count: 12500, input_tokens: 8000000, output_tokens: 1200000, cached_input_tokens: 500000, average_latency_ms: "842", error_rate: "0.012", retry_rate: "0.021", projection_method: "elapsed_day_linear" };
const recommendation = { id: recommendationID, workload_id: workloadID, recommendation_type: "smaller_hosted_model", priority: "high", confidence_level: "high", status: "new", currency: "USD", quality_risk: "medium", operational_risk: "low", required_next_action: "evaluation_required", rule_version: "v1", confidence: "0.82", estimated_monthly_savings: "64.25", estimated_implementation_cost: "120", estimated_break_even_months: "1.87", reason_codes: ["REPETITIVE_CLASSIFICATION"], evidence_summary: { analyzed_calls: 12500 }, confidence_inputs: { pricing_complete: true }, current_execution: { model: "northstar-frontier" }, proposed_execution: { model: "northstar-compact" }, created_at: now, updated_at: now };

async function authenticate(page: Page) {
  await page.context().addCookies([{ name: "access_token", value: "e2e-session", url: "http://127.0.0.1:3100", httpOnly: true, sameSite: "Lax" }]);
}

async function mockAPI(page: Page, overrides: Record<string, { status?: number; body: unknown }> = {}) {
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const overridden = overrides[url.pathname];
    if (overridden) return route.fulfill({ status: overridden.status ?? 200, contentType: "application/json", body: JSON.stringify(overridden.body) });
    let body: unknown;
    if (url.pathname === "/api/v1/me") body = { id: "55555555-5555-4555-8555-555555555555", organization_id: project.organization_id, role: "owner" };
    else if (url.pathname === "/api/v1/analytics/overview") body = overview;
    else if (url.pathname === "/api/v1/analytics/tokens") body = { input: overview.input_tokens, output: overview.output_tokens, cached_input: overview.cached_input_tokens };
    else if (url.pathname === "/api/v1/analytics/latency") body = { average_ms: "842", p50_ms: "710", p95_ms: "1420" };
    else if (url.pathname === "/api/v1/analytics/errors") body = { calls: 12500, errors: 150, retried_calls: 262, error_rate: "0.012", retry_rate: "0.021" };
    else if (url.pathname === "/api/v1/analytics/costs") body = { dimension: url.searchParams.get("dimension") ?? "project", breakdown: [{ key: url.searchParams.get("dimension") === "model" ? "northstar-frontier" : url.searchParams.get("dimension") === "provider" ? "Northstar AI" : project.name, cost: "126.50", calls: 12500, currency: "USD" }], daily: [{ date: "2026-08-05", cost: "61.25", calls: 6000 }, { date: "2026-08-06", cost: "65.25", calls: 6500 }], currency: "USD" };
    else if (url.pathname === "/api/v1/projects") body = { items: [project], limit: 100, offset: 0 };
    else if (url.pathname === `/api/v1/projects/${projectID}`) body = project;
    else if (url.pathname === "/api/v1/workloads") body = { items: [workload], limit: 100, offset: 0 };
    else if (url.pathname === `/api/v1/workloads/${workloadID}`) body = workload;
    else if (url.pathname === "/api/v1/llm-calls") body = { items: [{ id: "66666666-6666-4666-8666-666666666666", organization_id: project.organization_id, project_id: projectID, workload_id: workloadID, external_call_id: "call-100", provider: "Northstar AI", model: "northstar-frontier", estimated_cost: "0.0125", currency: "USD", created_at: now, idempotent_replay: false }], limit: 100, offset: 0 };
    else if (url.pathname === "/api/v1/recommendations") body = { items: [recommendation] };
    else if (url.pathname === `/api/v1/recommendations/${recommendationID}`) body = recommendation;
    else return route.fulfill({ status: 404, contentType: "application/json", body: JSON.stringify({ error: { code: "not_found", message: "not found", request_id: "e2e" } }) });
    return route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(body) });
  });
}

test("sign-in establishes a session and enters the protected dashboard", async ({ page }) => {
  await mockAPI(page);
  await page.route("**/api/v1/auth/login", async (route) => { await authenticate(page); await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ expires_at: now }) }); });
  await page.goto("/sign-in");
  await page.getByLabel("Email").fill("owner@example.test");
  await page.getByLabel("Password").fill("portfolio-password");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("AI operations at a glance");
});

test("overview renders observed metrics, trend, and recommendation status", async ({ page }) => {
  await authenticate(page); await mockAPI(page); await page.goto("/dashboard");
  await expect(page.getByText("USD 126.50")).toBeVisible();
  await expect(page.getByText("Not available")).toBeVisible();
  await expect(page.getByLabel("Daily observed cost chart")).toBeVisible();
  await expect(page.getByLabel("Token distribution chart")).toBeVisible();
  await expect(page.getByLabel("Latency and error chart")).toBeVisible();
  await expect(page.locator('meta[name="robots"]')).toHaveAttribute("content", /noindex/);
  const accessibility = await new AxeBuilder({ page }).analyze();
  expect(accessibility.violations.filter((item) => ["serious", "critical"].includes(item.impact ?? ""))).toEqual([]);
});

test("project filtering and details preserve organization-scoped data", async ({ page }) => {
  await authenticate(page); await mockAPI(page); await page.goto("/dashboard/projects");
  await page.getByLabel("Environment").selectOption("production");
  await expect(page.getByRole("link", { name: project.name })).toBeVisible();
  await page.getByRole("link", { name: project.name }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(project.name);
  await expect(page.getByRole("link", { name: workload.name })).toBeVisible();
});

test("workload detail shows tokens, execution mix, and recommendation history", async ({ page }) => {
  await authenticate(page); await mockAPI(page); await page.goto(`/dashboard/workloads/${workloadID}`);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(workload.name);
  await expect(page.getByText("8,000,000")).toBeVisible();
  await expect(page.getByRole("link", { name: "smaller hosted model" })).toBeVisible();
});

test("call explorer renders redacted operational records", async ({ page }) => {
  await authenticate(page); await mockAPI(page); await page.goto("/dashboard/calls");
  await expect(page.getByRole("cell", { name: "Northstar AI" })).toBeVisible();
  await expect(page.getByText(/Full prompts and responses are not displayed/)).toBeVisible();
});

test("call explorer rejects malformed project filters before API execution", async ({ page }) => {
  await authenticate(page); await mockAPI(page); await page.goto("/dashboard/calls");
  await page.getByLabel("Project ID").fill("not-a-uuid");
  await expect(page.locator(".form-error")).toHaveText("Enter a valid project UUID.");
});

test("API errors render a safe recoverable state", async ({ page }) => {
  await authenticate(page); await mockAPI(page, { "/api/v1/projects": { status: 500, body: { error: { code: "internal_error", message: "private database detail", request_id: "e2e" } } } });
  await page.goto("/dashboard/projects");
  await expect(page.locator(".state-error")).toContainText("Data unavailable");
  await expect(page.locator(".state-error")).not.toContainText("private database detail");
});

test("empty project data renders an actionable empty state", async ({ page }) => {
  await authenticate(page); await mockAPI(page, { "/api/v1/projects": { body: { items: [], limit: 100, offset: 0 } } });
  await page.goto("/dashboard/projects");
  await expect(page.getByText("No matching projects")).toBeVisible();
});

test("an invalid API session returns to sign in", async ({ page }) => {
  await authenticate(page); await mockAPI(page, { "/api/v1/me": { status: 401, body: { error: { code: "unauthorized", message: "authentication is required", request_id: "e2e" } } } });
  await page.goto("/dashboard");
  await expect(page).toHaveURL(/\/sign-in\?next=%2Fdashboard/);
});
