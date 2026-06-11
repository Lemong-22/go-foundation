import { Button } from "@entropy-course/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@entropy-course/ui/components/card";
import type { ReactNode } from "react";

interface AuthCardProps {
  children: ReactNode;
  description: string;
  eyebrow: string;
  onSwitch: () => void;
  switchActionLabel: string;
  switchPrompt: string;
  title: string;
}

export default function AuthCard({
  children,
  description,
  eyebrow,
  onSwitch,
  switchActionLabel,
  switchPrompt,
  title,
}: AuthCardProps) {
  return (
    <Card className="min-w-0 gap-0 rounded-[14px] border border-border bg-card py-0 text-card-foreground shadow-[0_20px_70px_oklch(0.19_0.012_70_/_0.1)] ring-0">
      <CardHeader className="gap-2 border-b border-border px-6 py-6 sm:px-7">
        <p className="font-mono text-[10.5px] uppercase tracking-[0.1em] text-muted-foreground">
          {eyebrow}
        </p>
        <CardTitle className="font-serif text-[34px] leading-none tracking-normal text-foreground">
          {title}
        </CardTitle>
        <CardDescription className="max-w-[38ch] text-[14px] leading-6">
          {description}
        </CardDescription>
      </CardHeader>
      <CardContent className="px-6 py-6 sm:px-7">{children}</CardContent>
      <CardFooter className="justify-center gap-2 border-t border-border bg-muted/55 px-6 py-4 text-xs text-muted-foreground sm:px-7">
        <span>{switchPrompt}</span>
        <Button
          className="h-7 rounded-md px-2 text-primary hover:bg-accent hover:text-accent-foreground"
          onClick={onSwitch}
          size="sm"
          type="button"
          variant="ghost"
        >
          {switchActionLabel}
        </Button>
      </CardFooter>
    </Card>
  );
}
