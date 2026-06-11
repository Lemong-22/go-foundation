import { ReaderRouteState } from "@/components/reader-route-state";

export default function Loading() {
  return (
    <ReaderRouteState
      busy
      description="The sample course test surface is being prepared."
      label="Loading"
      title="Opening Course Test…"
      tone="loading"
    />
  );
}
