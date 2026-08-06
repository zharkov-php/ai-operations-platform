import createClient from "openapi-fetch";
import type { paths } from "./generated";

export type { components, operations, paths } from "./generated";

export function createAPIClient(baseUrl = "", headers?: Record<string, string>) {
  return createClient<paths>({ baseUrl, credentials: "include", headers });
}
