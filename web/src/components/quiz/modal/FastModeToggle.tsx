"use client"

type Props = { enabled: boolean; onToggle: () => void; hidden?: boolean }

export default function FastModeToggle({ enabled, onToggle, hidden }: Props) {
  if (hidden) return null
  return (
    <div className="flex items-center justify-center border-b border-gray-100 py-3 sm:py-4">
      <div className="flex items-center space-x-3">
        <span className="text-sm font-medium text-[#2D3436]">Fast Mode</span>
        <button
          onClick={onToggle}
          className={`relative h-6 w-12 rounded-full transition-colors duration-300 ${
            enabled ? 'bg-[#FF7B54]' : 'bg-gray-300'
          }`}
        >
          <div
            className={`absolute top-1 h-4 w-4 rounded-full bg-white transition-transform duration-300 ${
              enabled ? 'translate-x-7' : 'translate-x-1'
            }`}
          />
        </button>
        <span className="text-xs text-[#636E72]">
          {enabled ? 'Auto-advance enabled' : 'Manual progression'}
        </span>
      </div>
    </div>
  )
}


