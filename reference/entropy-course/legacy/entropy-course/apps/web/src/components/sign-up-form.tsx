import { Button } from "@entropy-course/ui/components/button";
import { Input } from "@entropy-course/ui/components/input";
import { Label } from "@entropy-course/ui/components/label";
import { useForm } from "@tanstack/react-form";
import { ArrowRight, Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import z from "zod";

import { authClient } from "@/lib/auth-client";
import { DEFAULT_LESSON_PATH } from "@/lib/learning-paths";

import AuthCard from "./auth-card";
import Loader from "./loader";

export default function SignUpForm({ onSwitchToSignIn }: { onSwitchToSignIn: () => void }) {
  const router = useRouter();
  const { isPending } = authClient.useSession();

  const form = useForm({
    defaultValues: {
      email: "",
      password: "",
      name: "",
    },
    onSubmit: async ({ value }) => {
      await authClient.signUp.email(
        {
          email: value.email,
          password: value.password,
          name: value.name,
        },
        {
          onSuccess: () => {
            router.replace(DEFAULT_LESSON_PATH);
            router.refresh();
            toast.success("Sign up successful");
          },
          onError: (error) => {
            toast.error(error.error.message || error.error.statusText);
          },
        },
      );
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(2, "Name must be at least 2 characters"),
        email: z.email("Invalid email address"),
        password: z.string().min(8, "Password must be at least 8 characters"),
      }),
    },
  });

  if (isPending) {
    return <Loader />;
  }

  return (
    <AuthCard
      description="Create an account before starting the JavaScript workbook."
      eyebrow="New workspace"
      onSwitch={onSwitchToSignIn}
      switchActionLabel="Sign in"
      switchPrompt="Already registered?"
      title="Create account"
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          void form.handleSubmit();
        }}
        className="flex flex-col gap-4"
      >
        <form.Field name="name">
          {(field) => {
            const hasErrors = field.state.meta.errors.length > 0;

            return (
              <div className="flex flex-col gap-2" data-invalid={hasErrors || undefined}>
                <Label
                  className="font-mono text-[10.5px] uppercase tracking-[0.08em] text-muted-foreground"
                  htmlFor={field.name}
                >
                  Name
                </Label>
                <Input
                  aria-invalid={hasErrors}
                  autoComplete="name"
                  className="h-10 rounded-md bg-background px-3 text-sm md:text-sm"
                  id={field.name}
                  name={field.name}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  value={field.state.value}
                />
                {field.state.meta.errors.map((error) =>
                  error?.message ? (
                    <p className="font-mono text-[11px] text-destructive" key={error.message}>
                      {error.message}
                    </p>
                  ) : null,
                )}
              </div>
            );
          }}
        </form.Field>

        <form.Field name="email">
          {(field) => {
            const hasErrors = field.state.meta.errors.length > 0;

            return (
              <div className="flex flex-col gap-2" data-invalid={hasErrors || undefined}>
                <Label
                  className="font-mono text-[10.5px] uppercase tracking-[0.08em] text-muted-foreground"
                  htmlFor={field.name}
                >
                  Email
                </Label>
                <Input
                  aria-invalid={hasErrors}
                  autoComplete="email"
                  className="h-10 rounded-md bg-background px-3 text-sm md:text-sm"
                  id={field.name}
                  name={field.name}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  type="email"
                  value={field.state.value}
                />
                {field.state.meta.errors.map((error) =>
                  error?.message ? (
                    <p className="font-mono text-[11px] text-destructive" key={error.message}>
                      {error.message}
                    </p>
                  ) : null,
                )}
              </div>
            );
          }}
        </form.Field>

        <form.Field name="password">
          {(field) => {
            const hasErrors = field.state.meta.errors.length > 0;

            return (
              <div className="flex flex-col gap-2" data-invalid={hasErrors || undefined}>
                <Label
                  className="font-mono text-[10.5px] uppercase tracking-[0.08em] text-muted-foreground"
                  htmlFor={field.name}
                >
                  Password
                </Label>
                <Input
                  aria-invalid={hasErrors}
                  autoComplete="new-password"
                  className="h-10 rounded-md bg-background px-3 text-sm md:text-sm"
                  id={field.name}
                  name={field.name}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  type="password"
                  value={field.state.value}
                />
                {field.state.meta.errors.map((error) =>
                  error?.message ? (
                    <p className="font-mono text-[11px] text-destructive" key={error.message}>
                      {error.message}
                    </p>
                  ) : null,
                )}
              </div>
            );
          }}
        </form.Field>

        <form.Subscribe
          selector={(state) => ({ canSubmit: state.canSubmit, isSubmitting: state.isSubmitting })}
        >
          {({ canSubmit, isSubmitting }) => (
            <Button
              className="mt-2 h-10 w-full rounded-md border-foreground bg-foreground text-sm text-background hover:bg-foreground/90 hover:text-background"
              disabled={!canSubmit || isSubmitting}
              type="submit"
            >
              {isSubmitting ? (
                <>
                  <Loader2 aria-hidden="true" className="animate-spin" data-icon="inline-start" />
                  Creating account
                </>
              ) : (
                <>
                  Create account
                  <ArrowRight aria-hidden="true" data-icon="inline-end" />
                </>
              )}
            </Button>
          )}
        </form.Subscribe>
      </form>
    </AuthCard>
  );
}
