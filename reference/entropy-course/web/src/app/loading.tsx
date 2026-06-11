import { ReaderRouteState } from "@/components/reader-route-state";

export default function Loading() {
  return (
    <ReaderRouteState
      busy
      description="Published courses are being fetched from the course service."
      label="Loading"
      title="Preparing Catalog…"
      tone="loading"
    />
  );
}
