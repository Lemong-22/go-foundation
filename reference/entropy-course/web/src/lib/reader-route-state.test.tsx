import { describe, expect, test } from "bun:test";
import { renderToStaticMarkup } from "react-dom/server";

import {
  ReaderInlineState,
  ReaderRouteState,
} from "@/components/reader-route-state";

describe("reader route states", () => {
  test("renders a busy loading state with reader skeleton markup", () => {
    const html = renderToStaticMarkup(
      <ReaderRouteState
        busy
        description="Published courses are being fetched from the course service."
        label="Loading"
        title="Preparing Catalog…"
        tone="loading"
      />,
    );

    expect(html).toContain('aria-busy="true"');
    expect(html).toContain("Preparing Catalog");
    expect(html).toContain("route-state-bars");
  });

  test("renders route actions and inline empty states consistently", () => {
    const routeHtml = renderToStaticMarkup(
      <ReaderRouteState
        actions={[
          {
            href: "/",
            label: "Back to Catalog",
            variant: "primary",
          },
        ]}
        description="This published lesson is not available in the reader."
        label="Lesson not found"
        title="Reader Unavailable"
        tone="not-found"
      />,
    );
    const inlineHtml = renderToStaticMarkup(
      <ReaderInlineState
        description="Lesson blocks will appear here as soon as they are published."
        label="No blocks"
        title="The lesson is empty"
      />,
    );

    expect(routeHtml).toContain('href="/"');
    expect(routeHtml).toContain("Back to Catalog");
    expect(routeHtml).toContain("route-state--not-found");
    expect(inlineHtml).toContain("reader-inline-state");
    expect(inlineHtml).toContain("The lesson is empty");
  });
});
