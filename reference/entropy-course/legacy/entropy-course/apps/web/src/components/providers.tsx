"use client";

import { Toaster } from "@entropy-course/ui/components/sonner";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

import { queryClient } from "@/utils/trpc";

import { ThemeProvider } from "./theme-provider";

const showReactQueryDevtools = process.env.NEXT_PUBLIC_REACT_QUERY_DEVTOOLS === "true";

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
      <QueryClientProvider client={queryClient}>
        {children}
        {showReactQueryDevtools ? <ReactQueryDevtools /> : null}
      </QueryClientProvider>
      <Toaster richColors />
    </ThemeProvider>
  );
}
