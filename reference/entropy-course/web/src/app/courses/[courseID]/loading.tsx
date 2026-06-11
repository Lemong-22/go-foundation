import { ReaderRouteState } from "@/components/reader-route-state";

export default function Loading() {
  return (
    <ReaderRouteState
      busy
      description="The course outline and lesson sequence are being loaded."
      label="Loading"
      title="Opening Syllabus…"
      tone="loading"
    />
  );
}
