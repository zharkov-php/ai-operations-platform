import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import Home from "./page";

describe("home page", () => {
  it("describes the evidence-based product", () => {
    render(<Home />);
    expect(screen.getByRole("heading", { name: /make ai execution decisions/i })).toBeInTheDocument();
  });
});
