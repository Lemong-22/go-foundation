import { ReaderRouteState } from "@/components/reader-route-state";

export default function NotFound() {
  return (
    <ReaderRouteState
      actions={[
        {
          href: "/",
          label: "Open Catalog",
          variant: "primary",
        },
      ]}
      description="The requested course reader page is not available yet."
      label="Not found"
      title="Reader Surface Unavailable"
      tone="not-found"
    />
  );
}
