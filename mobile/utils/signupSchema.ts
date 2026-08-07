import z from "zod";
import { isPasswordTooLong, isPasswordTooShort } from "./passwordPolicy";

export function createSignupSchema(
  shortPasswordMessage: string,
  longPasswordMessage: string,
) {
  return z.object({
    fullName: z
      .string()
      .min(2, "Required")
      .regex(/\s/, "Please enter your full name"),
    email: z.string().email().min(2, "Required"),
    password: z
      .string()
      .refine((value) => !isPasswordTooShort(value), shortPasswordMessage)
      .refine((value) => !isPasswordTooLong(value), longPasswordMessage),
  });
}

export type SignupFormData = z.infer<ReturnType<typeof createSignupSchema>>;
