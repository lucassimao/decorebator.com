'use client'

import React, { useState, FormEvent, Suspense } from 'react'
import { LockClosedIcon, CheckCircleIcon, ExclamationCircleIcon } from '@heroicons/react/24/solid'
import { useSearchParams } from 'next/navigation'
import { useTranslations } from 'next-intl'

// Separate component that uses useSearchParams
const ResetPasswordFormContent: React.FC = () => {
  const t = useTranslations('resetPassword')
  const [newPassword, setNewPassword] = useState<string>('')
  const [confirmPassword, setConfirmPassword] = useState<string>('')
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const searchParams = useSearchParams()

  const token = searchParams.get('token')

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)
    setSuccessMessage(null)

    if (!newPassword || !confirmPassword) {
      setError(t('errors.fillBothFields'))
      return
    }

    if (newPassword.length < 4) {
      setError(t('errors.tooShort'))
      return
    }

    if (newPassword !== confirmPassword) {
      setError(t('errors.noMatch'))
      return
    }

    if (!token) {
      setError(t('errors.missingToken'))
      return
    }

    setIsLoading(true)

    try {
      const apiBase = process.env.NEXT_PUBLIC_API_URL
      const res = await fetch(`${apiBase}/password/reset`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          token,
          password: newPassword,
        }),
      })

      if (!res.ok) {
        if (res.status === 400) {
          const err = await res.json().catch(() => null)
          if (err?.error === 'token_expired') {
            setError(t('errors.tokenExpired'))
            return
          }
          if (err?.error === 'token_invalid') {
            setError(t('errors.tokenInvalid'))
            return
          }
        }
        throw new Error('reset_failed')
      }

      setSuccessMessage(t('success'))
      setNewPassword('')
      setConfirmPassword('')
    } catch (apiError) {
      // Handle API errors (e.g., token expired, server error)
      if (!error) {
        setError(t('errors.resetFailed'))
      }
      console.error('API Error:', apiError)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800 p-4">
      <div className="w-full max-w-md space-y-6 rounded-xl bg-white p-8 shadow-2xl">
        <div className="flex flex-col items-center">
          <LockClosedIcon className="mb-4 h-16 w-16 text-blue-600" />
          <h2 className="text-center text-3xl font-bold text-gray-800">{t('title')}</h2>
          <p className="mt-2 text-center text-sm text-gray-600">{t('subtitle')}</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="newPassword" className="mb-1 block text-sm font-medium text-gray-700">
              {t('newPassword')}
            </label>
            <input
              id="newPassword"
              name="newPassword"
              type="password"
              autoComplete="new-password"
              required
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="block w-full appearance-none rounded-lg border border-gray-300 px-4 py-3 text-gray-900 placeholder-gray-400 shadow-sm transition duration-150 ease-in-out focus:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:outline-none sm:text-sm"
              placeholder={t('newPasswordPlaceholder')}
            />
          </div>

          <div>
            <label
              htmlFor="confirmPassword"
              className="mb-1 block text-sm font-medium text-gray-700"
            >
              {t('confirmPassword')}
            </label>
            <input
              id="confirmPassword"
              name="confirmPassword"
              type="password"
              autoComplete="new-password"
              required
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="block w-full appearance-none rounded-lg border border-gray-300 px-4 py-3 text-gray-900 placeholder-gray-400 shadow-sm transition duration-150 ease-in-out focus:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:outline-none sm:text-sm"
              placeholder={t('confirmPasswordPlaceholder')}
            />
          </div>

          {error && (
            <div
              className="flex items-center rounded-lg border border-red-300 bg-red-100 p-3 text-sm text-red-700"
              role="alert"
            >
              <ExclamationCircleIcon className="mr-2 h-5 w-5 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {successMessage && (
            <div
              className="flex items-center rounded-lg border border-green-300 bg-green-100 p-3 text-sm text-green-700"
              role="alert"
            >
              <CheckCircleIcon className="mr-2 h-5 w-5 flex-shrink-0" />
              <span>{successMessage}</span>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className={`flex w-full justify-center rounded-lg border border-transparent px-4 py-3 text-sm font-medium text-white shadow-sm ${
                isLoading
                  ? 'cursor-not-allowed bg-blue-400'
                  : 'bg-blue-600 hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none'
              } transition duration-150 ease-in-out`}
            >
              {isLoading ? (
                <>
                  <svg
                    className="mr-3 -ml-1 h-5 w-5 animate-spin text-white"
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
                t('resetButton')
              )}
            </button>
          </div>
        </form>

        {!successMessage && (
          <p className="mt-6 text-center text-xs text-gray-500">{t('helpText')}</p>
        )}
      </div>
    </div>
  )
}

// Loading fallback component
const ResetPasswordFormSkeleton: React.FC = () => {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800 p-4">
      <div className="w-full max-w-md space-y-6 rounded-xl bg-white p-8 shadow-2xl">
        <div className="flex flex-col items-center">
          <div className="mb-4 h-16 w-16 animate-pulse rounded-full bg-gray-200" />
          <div className="mb-2 h-8 w-48 animate-pulse rounded bg-gray-200" />
          <div className="h-4 w-64 animate-pulse rounded bg-gray-200" />
        </div>
        <div className="space-y-4">
          <div className="h-10 animate-pulse rounded bg-gray-200" />
          <div className="h-10 animate-pulse rounded bg-gray-200" />
          <div className="h-10 animate-pulse rounded bg-gray-200" />
        </div>
      </div>
    </div>
  )
}

// Main component with Suspense boundary
const ResetPasswordForm: React.FC = () => {
  return (
    <Suspense fallback={<ResetPasswordFormSkeleton />}>
      <ResetPasswordFormContent />
    </Suspense>
  )
}

export default ResetPasswordForm
