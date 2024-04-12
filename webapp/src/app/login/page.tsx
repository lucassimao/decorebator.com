"use client";
import { authenticate, isAuthenticated } from "@/lib/auth";
import { Button, Loader, PasswordInput, TextInput, Alert } from "@mantine/core";
import { useForm } from "@mantine/form";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useFormStatus } from "react-dom";

export default function LoginPage() {
  const { pending: isFormPending } = useFormStatus();
  const router = useRouter();
  const form = useForm({
    mode: "uncontrolled",
    initialValues: { email: "", password: "" },
    validate: {
      email: (value) => (/^\S+@\S+$/.test(value) ? null : "Invalid email"),
    },
  });
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (values: typeof form.values) => {
    try {
      await authenticate(values.email, values.password);
      router.push("/dashboard");
    } catch (error: any) {
      setError(error.message || "Something went wrong");
    }
  };

  return (
    <main className="bg-gray-100 flex items-center justify-center min-h-screen">
      <div className="w-full max-w-xs p-6 mx-4 bg-white rounded shadow-md sm:max-w-sm lg:max-w-md">
        {error && (
          <div className="mb-6 p-0">
            <Alert
              onClose={() => setError(null)}
              withCloseButton
              variant="light"
              color="red"
            >
              {error}
            </Alert>
          </div>
        )}
        <form
          onReset={form.onReset}
          onSubmit={form.onSubmit(handleSubmit)}
          method="POST"
        >
          <TextInput
            withAsterisk
            label="Email"
            classNames={{
              label: "mb-2",
              root: "mb-4",
            }}
            placeholder="your@email.com"
            {...form.getInputProps("email")}
            key="email"
          />
          <PasswordInput
            withAsterisk
            label="Password"
            classNames={{
              label: "mb-2",
              root: "mb-4",
            }}
            {...form.getInputProps("password")}
            key="password"
          />

          <div className="flex items-center justify-between">
            {isFormPending ? (
              <Loader color="blue" />
            ) : (
              <Button variant="filled" type="submit">
                Sign In
              </Button>
            )}
          </div>
        </form>
      </div>
    </main>
  );
}
