"use client";

import { usePathname } from "next/navigation";

import Header from "@/components/header";

export default function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isFocusedRoute = pathname.startsWith("/course/") || pathname === "/login";

  if (isFocusedRoute) {
    return <div className="min-h-svh">{children}</div>;
  }

  return (
    <div className="grid min-h-svh grid-rows-[auto_1fr]">
      <Header />
      {children}
    </div>
  );
}
