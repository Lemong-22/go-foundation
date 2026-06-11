"use client";

import { ReaderRouteErrorBoundary } from "@/components/reader-route-error";

type ErrorProps = {
  error: Error & { digest?: string };
  unstable_retry: () => void;
};

export default function Error({ error, unstable_retry }: ErrorProps) {
  return (
    <ReaderRouteErrorBoundary
      description="The sample course test route stopped while rendering. Try loading it again."
      error={error}
      label="Route error"
      title="Course Test Interrupted"
      unstable_retry={unstable_retry}
    />
  );
}
