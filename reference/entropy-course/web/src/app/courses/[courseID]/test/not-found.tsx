import { ReaderRouteState } from "@/components/reader-route-state";

export default function NotFound() {
  return (
    <ReaderRouteState
      actions={[
        {
          href: "/",
          label: "Back to Catalog",
          variant: "primary",
        },
      ]}
      description="This course test is not available for the published reader."
      label="Course test not found"
      title="Course Test Unavailable"
      tone="not-found"
    />
  );
}
