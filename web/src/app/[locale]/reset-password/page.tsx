"use client";

import React, { useState, FormEvent, Suspense } from "react";
import {
  LockClosedIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/solid";
import { useSearchParams } from "next/navigation";

// Separate component that uses useSearchParams
const ResetPasswordFormContent: React.FC = () => {
  const [newPassword, setNewPassword] = useState<string>("");
  const [confirmPassword, setConfirmPassword] = useState<string>("");
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const searchParams = useSearchParams();

  const token = searchParams.get("token");

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);
    setSuccessMessage(null);

    if (!newPassword || !confirmPassword) {
      setError("Please fill in both password fields.");
      return;
    }

    if (newPassword.length < 4) {
      setError("Password must be at least 4 characters long.");
      return;
    }

    if (newPassword !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    setIsLoading(true);

    try {
      const apiBase = process.env.NEXT_PUBLIC_API_BASE;
      const res = await fetch(`${apiBase}/password/reset`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          token,
          password: newPassword,
        }),
      });

      if (!res.ok) {
        // handle errors, e.g. 4xx or 5xx responses
        const err = await res.json();
        throw new Error(err.message);
      }

      setSuccessMessage("Password has been reset successfully!");
      setNewPassword("");
      setConfirmPassword("");
    } catch (apiError) {
      // Handle API errors (e.g., token expired, server error)
      setError("Failed to reset password. Please try again.");
      console.error("API Error:", apiError);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-900 to-slate-800 p-4">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-xl shadow-2xl">
        <div className="flex flex-col items-center">
          <LockClosedIcon className="w-16 h-16 text-blue-600 mb-4" />
          <h2 className="text-3xl font-bold text-center text-gray-800">
            Reset Your Password
          </h2>
          <p className="text-sm text-gray-600 mt-2 text-center">
            Enter your new password below. Make sure it&apos;s strong and memorable.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label
              htmlFor="newPassword"
              className="block text-sm font-medium text-gray-700 mb-1"
            >
              New Password
            </label>
            <input
              id="newPassword"
              name="newPassword"
              type="password"
              autoComplete="new-password"
              required
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="appearance-none block w-full px-4 py-3 border border-gray-300 rounded-lg shadow-sm placeholder-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm transition duration-150 ease-in-out"
              placeholder="Enter new password"
            />
          </div>

          <div>
            <label
              htmlFor="confirmPassword"
              className="block text-sm font-medium text-gray-700 mb-1"
            >
              Confirm New Password
            </label>
            <input
              id="confirmPassword"
              name="confirmPassword"
              type="password"
              autoComplete="new-password"
              required
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="appearance-none block w-full px-4 py-3 border border-gray-300 rounded-lg shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm transition duration-150 ease-in-out"
              placeholder="Confirm new password"
            />
          </div>

          {error && (
            <div
              className="flex items-center p-3 text-sm text-red-700 bg-red-100 border border-red-300 rounded-lg"
              role="alert"
            >
              <ExclamationCircleIcon className="w-5 h-5 mr-2 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {successMessage && (
            <div
              className="flex items-center p-3 text-sm text-green-700 bg-green-100 border border-green-300 rounded-lg"
              role="alert"
            >
              <CheckCircleIcon className="w-5 h-5 mr-2 flex-shrink-0" />
              <span>{successMessage}</span>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className={`w-full flex justify-center py-3 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white 
                          ${
                            isLoading
                              ? "bg-blue-400 cursor-not-allowed"
                              : "bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                          }
                          transition duration-150 ease-in-out`}
            >
              {isLoading ? (
                <>
                  <svg
                    className="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    ></circle>
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  Processing...
                </>
              ) : (
                "Reset Password"
              )}
            </button>
          </div>
        </form>

        {!successMessage && (
          <p className="mt-6 text-center text-xs text-gray-500">
            Remember to choose a strong password you haven&apos;t used before.
          </p>
        )}
      </div>
    </div>
  );
};

// Loading fallback component
const ResetPasswordFormSkeleton: React.FC = () => {
  return (
    <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-slate-900 to-slate-800 p-4">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-xl shadow-2xl">
        <div className="flex flex-col items-center">
          <div className="w-16 h-16 bg-gray-200 rounded-full animate-pulse mb-4" />
          <div className="h-8 w-48 bg-gray-200 rounded animate-pulse mb-2" />
          <div className="h-4 w-64 bg-gray-200 rounded animate-pulse" />
        </div>
        <div className="space-y-4">
          <div className="h-10 bg-gray-200 rounded animate-pulse" />
          <div className="h-10 bg-gray-200 rounded animate-pulse" />
          <div className="h-10 bg-gray-200 rounded animate-pulse" />
        </div>
      </div>
    </div>
  );
};

// Main component with Suspense boundary
const ResetPasswordForm: React.FC = () => {
  return (
    <Suspense fallback={<ResetPasswordFormSkeleton />}>
      <ResetPasswordFormContent />
    </Suspense>
  );
};

export default ResetPasswordForm;