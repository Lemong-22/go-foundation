"use client";

import { ReaderRouteErrorBoundary } from "@/components/reader-route-error";

type ErrorProps = {
  error: Error & { digest?: string };
  unstable_retry: () => void;
};

export default function Error({ error, unstable_retry }: ErrorProps) {
  return (
    <ReaderRouteErrorBoundary
      description="The catalog route stopped while rendering. Try loading it again."
      error={error}
      label="Route error"
      title="Catalog Interrupted"
      unstable_retry={unstable_retry}
    />
  );
}
