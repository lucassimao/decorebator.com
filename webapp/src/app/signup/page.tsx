"use client";
import Link from "next/link";
import { authenticate, signup } from "@/lib/auth";
import {
  Alert,
  Button,
  Loader,
  PasswordInput,
  TextInput,
  rem,
} from "@mantine/core";
import { useForm } from "@mantine/form";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useFormStatus } from "react-dom";
import { IconLock } from "@tabler/icons-react";
import { useMutation } from "@tanstack/react-query";
import { ValidationError } from "@/lib/common";

const icon = (
  <IconLock style={{ width: rem(18), height: rem(18) }} stroke={1.5} />
);

export default function SignupPage() {
  const { pending: isFormPending } = useFormStatus();
  const router = useRouter();
  const form = useForm({
    mode: "uncontrolled",
    initialValues: {
      email: "",
      password: "",
      firstName: "",
      lastName: "",
      confirmPassword: "",
    },
    validate: {
      firstName: (value) => (value?.trim() ? null : "First name is required"),
      lastName: (value) => (value?.trim() ? null : "Last name is required"),
      password: (value) => (value?.trim() ? null : "Password is required"),
      email: (value) =>
        value?.trim() && (/^\S+@\S+$/.test(value) ? null : "Invalid email"),
      confirmPassword: (value, { password }) =>
        value && value === password ? null : "Passwords do not match",
    },
  });

  const { isError, error, mutateAsync } = useMutation({
    mutationFn: () => signup(form.getValues()),
    onSuccess: () => {
      router.push("/wordlists");
    },
  });

  const handleSubmit = async (values: typeof form.values) => {
    await mutateAsync();
  };

  return (
    <main className="bg-gray-100 flex items-center justify-center min-h-screen">
      <div className="w-full max-w-xs p-6 mx-4 bg-white rounded shadow-md sm:max-w-sm lg:max-w-md">
        {isError && (
          <div className="mb-6 p-0">
            <Alert
              // onClose={() => null}
              // withCloseButton
              variant="light"
              color="red"
            >
              {error instanceof ValidationError ? (
                Object.values(error.validationErrors).map((message, i) => (
                  <div className="normal-case" key={i}>
                    {message}
                  </div>
                ))
              ) : (
                <span className="normal-case">{error.message}</span>
              )}
            </Alert>
          </div>
        )}
        <form
          onReset={form.onReset}
          onSubmit={form.onSubmit(handleSubmit)}
          method="POST"
        >
          <div className="grid grid-cols-2 gap-4">
            <TextInput
              withAsterisk
              label="First name"
              classNames={{
                label: "mb-2",
                root: "mb-4",
              }}
              placeholder="Your first name"
              {...form.getInputProps("firstName")}
              key="firstName"
            />

            <TextInput
              withAsterisk
              label="Last name"
              classNames={{
                label: "mb-2",
                root: "mb-4",
              }}
              placeholder="Your last name"
              {...form.getInputProps("lastName")}
              key="lastName"
            />
          </div>

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
            placeholder="Password"
            key="password"
          />

          <PasswordInput
            withAsterisk
            label="Confirm Password"
            classNames={{
              label: "mb-2",
              root: "mb-4",
            }}
            leftSection={icon}
            placeholder="Confirm Password"
            {...form.getInputProps("confirmPassword")}
            key="confirmPassword"
          />

          <div className="flex items-center justify-between">
            <Button
              variant="default"
              type="button"
              component={Link}
              href="/login"
            >
              Have an account? Sign in
            </Button>

            {isFormPending ? (
              <Loader color="blue" />
            ) : (
              <Button variant="filled" type="submit">
                Register
              </Button>
            )}
          </div>
        </form>
      </div>
    </main>
  );
}
