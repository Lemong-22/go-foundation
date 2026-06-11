import type { Route } from "next";
import Link from "next/link";
import type { ReactNode } from "react";

type ReaderRouteStateTone = "empty" | "error" | "loading" | "not-found";

type ReaderRouteStateAction = {
  href: Route;
  label: string;
  variant?: "primary" | "secondary";
};

type ReaderRouteStateProps = {
  actions?: ReaderRouteStateAction[];
  busy?: boolean;
  children?: ReactNode;
  description: string;
  label: string;
  mark?: string;
  title: string;
  tone?: ReaderRouteStateTone;
};

type ReaderInlineStateProps = {
  children?: ReactNode;
  className?: string;
  description: string;
  label: string;
  title: string;
};

export function ReaderRouteState({
  actions = [],
  busy = false,
  children,
  description,
  label,
  mark,
  title,
  tone = "empty",
}: ReaderRouteStateProps) {
  return (
    <main
      aria-busy={busy || undefined}
      className={`route-state route-state--${tone}`}
      id="main-content"
    >
      <section className="route-state-card">
        <span className="route-state-mark" aria-hidden="true">
          {mark ?? markForTone(tone)}
        </span>
        <p className="mono-label">{label}</p>
        <h1>{title}</h1>
        <p>{description}</p>
        {busy ? <LoadingBars /> : null}
        {children}
        {actions.length > 0 ? (
          <div className="route-state-actions">
            {actions.map((action) => (
              <Link
                className={`route-state-action route-state-action--${
                  action.variant ?? "secondary"
                }`}
                href={action.href}
                key={`${action.href}-${action.label}`}
              >
                {action.label}
              </Link>
            ))}
          </div>
        ) : null}
      </section>
    </main>
  );
}

export function ReaderInlineState({
  children,
  className,
  description,
  label,
  title,
}: ReaderInlineStateProps) {
  const classes = ["catalog-state", "reader-inline-state", className]
    .filter(Boolean)
    .join(" ");

  return (
    <section className={classes}>
      <p className="mono-label">{label}</p>
      <h2>{title}</h2>
      <p>{description}</p>
      {children}
    </section>
  );
}

function LoadingBars() {
  return (
    <div className="route-state-bars" aria-hidden="true">
      <span />
      <span />
      <span />
    </div>
  );
}

function markForTone(tone: ReaderRouteStateTone) {
  switch (tone) {
    case "error":
      return "!";
    case "loading":
      return "RD";
    case "not-found":
      return "404";
    default:
      return "EC";
  }
}
