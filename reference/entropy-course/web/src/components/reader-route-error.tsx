"use client";

import { useEffect } from "react";

import { ReaderRouteState } from "@/components/reader-route-state";

type ReaderRouteErrorBoundaryProps = {
  description: string;
  error: Error & { digest?: string };
  label: string;
  title: string;
  unstable_retry: () => void;
};

export function ReaderRouteErrorBoundary({
  description,
  error,
  label,
  title,
  unstable_retry,
}: ReaderRouteErrorBoundaryProps) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <ReaderRouteState
      description={description}
      label={label}
      title={title}
      tone="error"
    >
      <div className="route-state-actions">
        <button
          className="route-state-action route-state-action--primary"
          onClick={() => unstable_retry()}
          type="button"
        >
          Try Again
        </button>
      </div>
    </ReaderRouteState>
  );
}
