"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { api, apiMessage } from "@/lib/api";

const schema = z.object({ email: z.email(), password: z.string().min(8) });
type Values = z.infer<typeof schema>;

export function SignInForm() {
  const router = useRouter();
  const search = useSearchParams();
  const [serverError, setServerError] = useState("");
  const form = useForm<Values>({ defaultValues: { email: "", password: "" } });
  async function submit(raw: Values) {
    const parsed = schema.safeParse(raw);
    if (!parsed.success) { setServerError("Enter a valid email and a password of at least eight characters."); return; }
    setServerError("");
    const response = await api.POST("/api/v1/auth/login", { body: parsed.data });
    if (response.error) { setServerError(apiMessage(response.error)); return; }
    const next = search.get("next");
    router.replace(next?.startsWith("/dashboard") ? next : "/dashboard");
    router.refresh();
  }
  return <form className="sign-in-form" onSubmit={form.handleSubmit(submit)} noValidate><label>Email<input type="email" autoComplete="email" {...form.register("email")} /></label><label>Password<input type="password" autoComplete="current-password" {...form.register("password")} /></label>{serverError && <p className="form-error" role="alert">{serverError}</p>}<button className="button" type="submit" disabled={form.formState.isSubmitting}>{form.formState.isSubmitting ? "Signing in…" : "Sign in"}</button></form>;
}
