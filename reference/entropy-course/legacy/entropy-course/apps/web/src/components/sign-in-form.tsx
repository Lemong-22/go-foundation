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

export default function SignInForm({ onSwitchToSignUp }: { onSwitchToSignUp: () => void }) {
  const router = useRouter();
  const { isPending } = authClient.useSession();

  const form = useForm({
    defaultValues: {
      email: "",
      password: "",
    },
    onSubmit: async ({ value }) => {
      await authClient.signIn.email(
        {
          email: value.email,
          password: value.password,
        },
        {
          onSuccess: () => {
            router.replace(DEFAULT_LESSON_PATH);
            router.refresh();
            toast.success("Sign in successful");
          },
          onError: (error) => {
            toast.error(error.error.message || error.error.statusText);
          },
        },
      );
    },
    validators: {
      onSubmit: z.object({
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
      description="Use your Crashcourse account to return to the JavaScript workbook."
      eyebrow="Account access"
      onSwitch={onSwitchToSignUp}
      switchActionLabel="Create account"
      switchPrompt="No account yet?"
      title="Welcome back"
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          void form.handleSubmit();
        }}
        className="flex flex-col gap-4"
      >
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
                  autoComplete="current-password"
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
                  Signing in
                </>
              ) : (
                <>
                  Sign in
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
