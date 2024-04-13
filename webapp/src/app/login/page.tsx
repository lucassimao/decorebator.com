"use client";
import { authenticate } from "@/lib/auth";
import {
  Alert,
  Button,
  Loader,
  PasswordInput,
  TextInput,
  rem,
} from "@mantine/core";
import { useForm } from "@mantine/form";
import { IconLock } from "@tabler/icons-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useFormStatus } from "react-dom";

const icon = (
  <IconLock style={{ width: rem(18), height: rem(18) }} stroke={1.5} />
);

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
      router.push("/wordlists");
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
            leftSection={icon}
            {...form.getInputProps("password")}
            key="password"
          />

          <div className="flex items-center justify-between">
            <Button
              variant="default"
              type="button"
              component={Link}
              href="/signup"
            >
              Don't have an account? Sign up
            </Button>

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
