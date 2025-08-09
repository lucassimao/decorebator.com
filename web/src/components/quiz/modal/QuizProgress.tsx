"use client"

type Props = { current: number; total: number; hidden?: boolean }

export default function QuizProgress({ current, total, hidden }: Props) {
  if (hidden || total <= 0) return null
  return (
    <div className="h-2 bg-gray-100">
      <div
        className="h-full bg-gradient-to-r from-[#FF7B54] to-orange-600 transition-[width] duration-500 ease-out will-change-[width]"
        style={{ width: `${(Math.min(total, Math.max(1, current)) / total) * 100}%` }}
      />
    </div>
  )
}


