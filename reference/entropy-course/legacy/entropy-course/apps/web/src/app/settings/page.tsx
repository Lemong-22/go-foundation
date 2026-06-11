import { buttonVariants } from "@entropy-course/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@entropy-course/ui/components/card";
import type { Route } from "next";
import Link from "next/link";

import { DEFAULT_LESSON_PATH } from "@/lib/learning-paths";
import { requireSession } from "@/lib/server-session";

export default async function SettingsPage() {
  const session = await requireSession();
  const accountFields = [
    { label: "Name", value: session.user.name ?? "Not set" },
    { label: "Email", value: session.user.email },
  ];

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-5 py-10 sm:px-8">
      <header className="flex flex-col gap-2">
        <p className="font-mono text-[11px] uppercase tracking-[0.1em] text-muted-foreground">Settings</p>
        <h1 className="font-serif text-[42px] leading-none tracking-normal text-foreground">Account</h1>
      </header>

      <Card className="rounded-lg">
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Basic account details for your Crashcourse workspace.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {accountFields.map((field) => (
            <div
              key={field.label}
              className="flex flex-col gap-1 rounded-md border border-border bg-background px-3 py-2.5 sm:flex-row sm:items-center sm:justify-between"
            >
              <span className="font-mono text-[10.5px] uppercase tracking-[0.08em] text-muted-foreground">
                {field.label}
              </span>
              <span className="text-sm font-medium text-foreground">{field.value}</span>
            </div>
          ))}
        </CardContent>
        <CardFooter>
          <Link className={buttonVariants({ variant: "outline" })} href={DEFAULT_LESSON_PATH as Route}>
            Back to lesson
          </Link>
        </CardFooter>
      </Card>
    </main>
  );
}
