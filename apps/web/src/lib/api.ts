"use client";

import { createAPIClient } from "@ai-operations/api-client";

export const api = createAPIClient("");

export function apiMessage(error: unknown): string {
  if (error && typeof error === "object" && "error" in error) {
    const body = error as { error?: { message?: string } };
    if (body.error?.message) return body.error.message;
  }
  return "The platform API could not complete this request.";
}
