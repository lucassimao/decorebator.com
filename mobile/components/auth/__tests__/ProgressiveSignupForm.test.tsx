import React from "react";
import { render, fireEvent, waitFor } from "@testing-library/react-native";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import z from "zod";
import { ProgressiveSignupForm } from "../ProgressiveSignupForm";
import { authLightTheme } from "@/theme/authTheme";
import {
  getResponsiveFontSizes,
  getResponsiveSpacing,
} from "@/utils/responsive";
import type { ResponsiveValues } from "@/contexts/ThemeContext";

// Mock expo-web-browser
jest.mock("expo-web-browser", () => ({
  openBrowserAsync: jest.fn(),
}));

// Mock expo-router Link component
jest.mock("expo-router", () => ({
  Link: ({ children }: { children: React.ReactNode }) => children,
}));

// Define the schema matching the real signup form
const signupSchema = z
  .object({
    fullName: z
      .string()
      .min(2, "Required")
      .regex(/\s/, "Please enter your full name"),
    email: z.string().email().min(2, "Required"),
    password: z.string().min(5, "Required"),
  })
  .required();

type SignupFormData = z.infer<typeof signupSchema>;

// Test wrapper component that provides form context
const TestWrapper: React.FC<{
  onSubmit: jest.Mock;
  children: (props: {
    form: ReturnType<typeof useForm<SignupFormData>>;
    theme: typeof authLightTheme;
    responsive: ResponsiveValues;
  }) => React.ReactNode;
}> = ({ onSubmit, children }) => {
  const form = useForm<SignupFormData>({
    resolver: zodResolver(signupSchema),
    mode: "onBlur",
    defaultValues: {
      email: "",
      fullName: "",
      password: "",
    },
  });

  // Create responsive values for small screen (progressive disclosure)
  const responsive: ResponsiveValues = {
    spacing: getResponsiveSpacing(375),
    fontSizes: getResponsiveFontSizes(375),
    screenWidth: 375,
    screenHeight: 667,
    category: "small",
    keyboardBehavior: "padding",
    keyboardOffset: 20,
    isSmallPhone: true,
    isMediumPhone: false,
    isLargePhone: false,
    isExtraLargePhone: false,
    getValueForSize: (small, medium, large, xlarge) => small,
    getScaledFont: (fontName) => getResponsiveFontSizes(375)[fontName],
  };

  return (
    <>
      {children({
        form,
        theme: authLightTheme,
        responsive,
      })}
    </>
  );
};

describe("ProgressiveSignupForm", () => {
  it("renders email field on initial load", () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId, getByText } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <ProgressiveSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // Should show email input
    expect(getByTestId("email-input")).toBeTruthy();
    // Should show next button
    expect(getByTestId("email-next-button")).toBeTruthy();
    // Should show sign in link
    expect(getByText(/Already have an account/)).toBeTruthy();
  });

  it("validates email and progresses to name field", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId, queryByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <ProgressiveSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    const emailInput = getByTestId("email-input");
    const nextButton = getByTestId("email-next-button");

    // Enter valid email
    fireEvent.changeText(emailInput, "test@example.com");
    fireEvent(emailInput, "blur");

    // Click next
    fireEvent.press(nextButton);

    // Wait for animation and step change
    await waitFor(() => {
      expect(queryByTestId("fullname-input")).toBeTruthy();
    });

    // Email field should be hidden
    expect(queryByTestId("email-input")).toBeNull();
  });

  it("shows error for invalid email", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId, queryByText } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <ProgressiveSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    const emailInput = getByTestId("email-input");
    const nextButton = getByTestId("email-next-button");

    // Enter invalid email
    fireEvent.changeText(emailInput, "notanemail");
    fireEvent(emailInput, "blur");

    // Try to proceed
    fireEvent.press(nextButton);

    // Should show validation error
    await waitFor(() => {
      expect(queryByText("Invalid email")).toBeTruthy();
    });
  });

  it("validates full name requires space", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId, queryByText } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <ProgressiveSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // First navigate to name step
    const emailInput = getByTestId("email-input");
    fireEvent.changeText(emailInput, "test@example.com");
    fireEvent(emailInput, "blur");
    fireEvent.press(getByTestId("email-next-button"));

    // Wait for name field
    await waitFor(() => {
      expect(getByTestId("fullname-input")).toBeTruthy();
    });

    const nameInput = getByTestId("fullname-input");
    const nextButton = getByTestId("name-next-button");

    // Enter name without space
    fireEvent.changeText(nameInput, "JohnDoe");
    fireEvent(nameInput, "blur");
    fireEvent.press(nextButton);

    // Should show validation error
    await waitFor(() => {
      expect(queryByText(/Please enter your full name/)).toBeTruthy();
    });
  });

  it("shows back button and navigates correctly", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId, queryByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <ProgressiveSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // Navigate to name step
    const emailInput = getByTestId("email-input");
    fireEvent.changeText(emailInput, "test@example.com");
    fireEvent(emailInput, "blur");
    fireEvent.press(getByTestId("email-next-button"));

    // Wait for name field and back button
    await waitFor(() => {
      expect(getByTestId("fullname-input")).toBeTruthy();
    });

    // Back button should be visible
    const backButton = getByTestId("back-button");
    expect(backButton).toBeTruthy();

    // Click back
    fireEvent.press(backButton);

    // Should return to email step
    await waitFor(() => {
      expect(getByTestId("email-input")).toBeTruthy();
    });

    // Back button should be hidden on email step
    expect(queryByTestId("back-button")).toBeNull();
  });

  it("completes full signup flow", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <ProgressiveSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // Step 1: Email
    const emailInput = getByTestId("email-input");
    fireEvent.changeText(emailInput, "test@example.com");
    fireEvent(emailInput, "blur");
    fireEvent.press(getByTestId("email-next-button"));

    // Step 2: Name
    await waitFor(() => {
      expect(getByTestId("fullname-input")).toBeTruthy();
    });

    const nameInput = getByTestId("fullname-input");
    fireEvent.changeText(nameInput, "John Doe");
    fireEvent(nameInput, "blur");
    fireEvent.press(getByTestId("name-next-button"));

    // Step 3: Password
    await waitFor(() => {
      expect(getByTestId("password-input")).toBeTruthy();
    });

    const passwordInput = getByTestId("password-input");
    fireEvent.changeText(passwordInput, "password123");
    fireEvent(passwordInput, "blur");

    // Submit
    const submitButton = getByTestId("signup-submit-button");
    fireEvent.press(submitButton);

    // Verify onSubmit was called with correct data
    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalled();
      const callArgs = mockOnSubmit.mock.calls[0][0];
      expect(callArgs).toEqual({
        email: "test@example.com",
        fullName: "John Doe",
        password: "password123",
      });
    });
  });

  it("disables buttons when fields are invalid", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <ProgressiveSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // Next button should be disabled with empty email
    const nextButton = getByTestId("email-next-button");
    expect(nextButton.props.disabled).toBe(true);

    // Enter email
    const emailInput = getByTestId("email-input");
    fireEvent.changeText(emailInput, "test@example.com");

    // Next button should now be enabled
    await waitFor(() => {
      expect(getByTestId("email-next-button").props.disabled).toBe(false);
    });
  });

  it("shows password visibility toggle", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <ProgressiveSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // Navigate to password step
    const emailInput = getByTestId("email-input");
    fireEvent.changeText(emailInput, "test@example.com");
    fireEvent(emailInput, "blur");
    fireEvent.press(getByTestId("email-next-button"));

    await waitFor(() => {
      expect(getByTestId("fullname-input")).toBeTruthy();
    });

    const nameInput = getByTestId("fullname-input");
    fireEvent.changeText(nameInput, "John Doe");
    fireEvent(nameInput, "blur");
    fireEvent.press(getByTestId("name-next-button"));

    await waitFor(() => {
      expect(getByTestId("password-input")).toBeTruthy();
    });

    // Password should be secure by default
    const passwordInput = getByTestId("password-input");
    expect(passwordInput.props.secureTextEntry).toBe(true);

    // Toggle visibility
    const toggleButton = getByTestId("password-toggle");
    fireEvent.press(toggleButton);

    // Password should now be visible
    expect(passwordInput.props.secureTextEntry).toBe(false);
  });
});
