"use client"

type Props = {
  isCompleted: boolean
  isWriteMode: boolean
  showFeedback: boolean
  canSubmitWrite: boolean
  onSubmitWrite: () => void
  onSkip: () => void
  onNext: () => void
}

export default function StickyActionBar({
  isCompleted,
  isWriteMode,
  showFeedback,
  canSubmitWrite,
  onSubmitWrite,
  onSkip,
  onNext,
}: Props) {
  if (isCompleted) return null
  return (
    <div className="border-t border-gray-200 bg-white/95 p-3 supports-[padding:env(safe-area-inset-bottom)]:pb-[calc(env(safe-area-inset-bottom)+0.5rem)] absolute bottom-0 left-0 right-0 z-20 sm:static">
      {isWriteMode ? (
        !showFeedback ? (
          <div className="flex gap-2">
            <button
              onClick={onSubmitWrite}
              disabled={!canSubmitWrite}
              className="flex-1 rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-3 font-semibold text-white transition-transform duration-150 hover:shadow-lg active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
            >
              Submit Answer
            </button>
            <button
              onClick={onSkip}
              className="rounded-xl border-2 border-gray-300 px-6 py-3 font-semibold text-gray-700 transition-transform duration-150 hover:border-gray-400 active:scale-[0.98]"
            >
              Skip
            </button>
          </div>
        ) : (
          <button
            onClick={onNext}
            className="w-full rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-3 font-semibold text-white transition-transform duration-150 hover:shadow-lg active:scale-[0.98]"
          >
            Next Question
          </button>
        )
      ) : showFeedback ? (
        <button
          onClick={onNext}
          className="w-full rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-3 font-semibold text-white transition-transform duration-150 hover:shadow-lg active:scale-[0.98]"
        >
          Next Question
        </button>
      ) : (
        <button
          onClick={onSkip}
          className="w-full rounded-xl border-2 border-gray-300 px-6 py-3 font-semibold text-gray-700 transition-transform duration-150 hover:border-gray-400 active:scale-[0.98]"
        >
          Skip Question
        </button>
      )}
    </div>
  )
}


