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
      description="This published course is not available in the reader catalog."
      label="Course not found"
      title="Syllabus Unavailable"
      tone="not-found"
    />
  );
}
