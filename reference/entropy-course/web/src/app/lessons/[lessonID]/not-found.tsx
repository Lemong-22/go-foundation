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
      description="This published lesson is not available in the reader."
      label="Lesson not found"
      title="Reader Unavailable"
      tone="not-found"
    />
  );
}
