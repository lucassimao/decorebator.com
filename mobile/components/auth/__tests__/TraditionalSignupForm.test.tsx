import React from "react";
import { render, fireEvent, waitFor } from "@testing-library/react-native";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import z from "zod";
import { TraditionalSignupForm } from "../TraditionalSignupForm";
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
    mode: "onChange",
    defaultValues: {
      email: "",
      fullName: "",
      password: "",
    },
  });

  // Create responsive values for medium screen (traditional form)
  const responsive: ResponsiveValues = {
    spacing: getResponsiveSpacing(430),
    fontSizes: getResponsiveFontSizes(430),
    screenWidth: 430,
    screenHeight: 900,
    category: "medium",
    keyboardBehavior: "padding",
    keyboardOffset: 10,
    isSmallPhone: false,
    isMediumPhone: true,
    isLargePhone: false,
    isExtraLargePhone: false,
    getValueForSize: (small, medium, large, xlarge) => medium,
    getScaledFont: (fontName) => getResponsiveFontSizes(430)[fontName],
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

describe("TraditionalSignupForm", () => {
  it("renders all fields at once", () => {
    const mockOnSubmit = jest.fn();

    const { getByPlaceholderText, getAllByText } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // All fields should be visible
    expect(getByPlaceholderText("Enter your email")).toBeTruthy();
    expect(getByPlaceholderText("Enter your full name")).toBeTruthy();
    expect(getByPlaceholderText("Choose a password")).toBeTruthy();
    const signUpButtons = getAllByText("Sign Up");
    expect(signUpButtons.length).toBeGreaterThan(0);
  });

  it("validates all fields", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId, getAllByText, queryByText } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    const submitButton = getByTestId("signup-submit-button");

    // Try to submit empty form
    fireEvent.press(submitButton);

    await waitFor(() => {
      const requiredElements = getAllByText("Required");
      expect(requiredElements.length).toBeGreaterThan(0);
    });

    // Enter invalid data
    const emailInput = getByTestId("email-input");
    const nameInput = getByTestId("fullname-input");
    const passwordInput = getByTestId("password-input");

    fireEvent.changeText(emailInput, "invalid");
    fireEvent(emailInput, "blur");

    fireEvent.changeText(nameInput, "John"); // No space
    fireEvent(nameInput, "blur");

    fireEvent.changeText(passwordInput, "123"); // Too short
    fireEvent(passwordInput, "blur");

    fireEvent.press(submitButton);

    await waitFor(() => {
      expect(queryByText("Invalid email")).toBeTruthy();
      expect(queryByText("Please enter your full name")).toBeTruthy();
    });
  });

  it("successfully submits valid form", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    const emailInput = getByTestId("email-input");
    const nameInput = getByTestId("fullname-input");
    const passwordInput = getByTestId("password-input");
    const submitButton = getByTestId("signup-submit-button");

    // Enter valid data
    fireEvent.changeText(emailInput, "test@example.com");
    fireEvent(emailInput, "blur");

    fireEvent.changeText(nameInput, "John Doe");
    fireEvent(nameInput, "blur");

    fireEvent.changeText(passwordInput, "password123");
    fireEvent(passwordInput, "blur");

    // Submit
    fireEvent.press(submitButton);

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

  it("toggles password visibility", () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    const passwordInput = getByTestId("password-input");
    const toggleButton = getByTestId("password-toggle");

    // Should be secure by default
    expect(passwordInput.props.secureTextEntry).toBe(true);

    // Toggle visibility
    fireEvent.press(toggleButton);
    expect(passwordInput.props.secureTextEntry).toBe(false);

    // Toggle back
    fireEvent.press(toggleButton);
    expect(passwordInput.props.secureTextEntry).toBe(true);
  });

  it("shows proper labels and icons", () => {
    const mockOnSubmit = jest.fn();

    const { getAllByText } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // Check for field labels
    expect(getAllByText("Email").length).toBeGreaterThan(0);
    expect(getAllByText("Full Name").length).toBeGreaterThan(0);
    expect(getAllByText("Password").length).toBeGreaterThan(0);
  });

  it("shows terms and privacy policy", () => {
    const mockOnSubmit = jest.fn();

    const { getByText } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // Terms text is rendered as a single block with Link components
    expect(getByText(/By signing up, you agree to our/)).toBeTruthy();
  });

  it("shows sign in link", () => {
    const mockOnSubmit = jest.fn();

    const { getByText } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    expect(getByText(/Already have an account/)).toBeTruthy();
    expect(getByText("Sign In")).toBeTruthy();
  });

  it("disables submit button when form is invalid", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    const submitButton = getByTestId("signup-submit-button");

    // Should be disabled initially
    expect(submitButton.props.disabled).toBe(true);

    // Fill in valid data
    const emailInput = getByTestId("email-input");
    const nameInput = getByTestId("fullname-input");
    const passwordInput = getByTestId("password-input");

    fireEvent.changeText(emailInput, "test@example.com");
    fireEvent(emailInput, "blur");

    fireEvent.changeText(nameInput, "John Doe");
    fireEvent(nameInput, "blur");

    fireEvent.changeText(passwordInput, "password123");
    fireEvent(passwordInput, "blur");

    // Should now be enabled - add slight delay for form validation
    await waitFor(
      () => {
        expect(getByTestId("signup-submit-button").props.disabled).toBe(false);
      },
      { timeout: 2000 },
    );
  });

  it("handles keyboard submit on password field", async () => {
    const mockOnSubmit = jest.fn();

    const { getByTestId } = render(
      <TestWrapper onSubmit={mockOnSubmit}>
        {({ form, theme, responsive }) => (
          <TraditionalSignupForm
            form={form}
            onSubmit={mockOnSubmit}
            theme={theme}
            responsive={responsive}
          />
        )}
      </TestWrapper>,
    );

    // Fill all fields
    const emailInput = getByTestId("email-input");
    const nameInput = getByTestId("fullname-input");
    const passwordInput = getByTestId("password-input");

    fireEvent.changeText(emailInput, "test@example.com");
    fireEvent(emailInput, "blur");

    fireEvent.changeText(nameInput, "John Doe");
    fireEvent(nameInput, "blur");

    fireEvent.changeText(passwordInput, "password123");
    fireEvent(passwordInput, "blur");

    // Submit via keyboard
    fireEvent(passwordInput, "submitEditing");

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
});
