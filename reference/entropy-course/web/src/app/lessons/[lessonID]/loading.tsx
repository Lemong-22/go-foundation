import { ReaderRouteState } from "@/components/reader-route-state";

export default function Loading() {
  return (
    <ReaderRouteState
      busy
      description="Lesson metadata and ordered content blocks are being loaded."
      label="Loading"
      title="Opening Lesson…"
      tone="loading"
    />
  );
}
