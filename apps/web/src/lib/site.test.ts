import { describe, expect, it } from "vitest";
import sitemap from "@/app/sitemap";
import robots from "@/app/robots";
import { articles, publicRoutes } from "./content";
import { pageMetadata, siteConfig } from "./site";

describe("public SEO contracts", () => {
  it("uses a unique article slug and route for every article", () => {
    expect(new Set(articles.map(({ slug }) => slug)).size).toBe(articles.length);
    expect(new Set(publicRoutes).size).toBe(publicRoutes.length);
  });

  it("creates canonical, Open Graph, and Twitter metadata", () => {
    const metadata = pageMetadata("Evidence", "Inspectable evidence", "/evidence");
    expect(metadata.alternates).toEqual({ canonical: "/evidence" });
    expect(metadata.openGraph).toMatchObject({ title: "Evidence", url: "/evidence" });
    expect(metadata.twitter).toMatchObject({ title: "Evidence" });
  });

  it("maps every public route into the sitemap", () => {
    expect(sitemap().map(({ url }) => url)).toEqual(publicRoutes.map((route) => `${siteConfig.url}${route === "/" ? "" : route}`));
  });

  it("keeps dashboard and API routes out of crawler access", () => {
    expect(robots().rules).toMatchObject({ allow: "/", disallow: ["/dashboard/", "/api/"] });
  });
});
