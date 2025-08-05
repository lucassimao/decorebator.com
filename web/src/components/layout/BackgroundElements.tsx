import React from 'react'

const BackgroundElements: React.FC = () => {
  return (
    <div className="pointer-events-none fixed inset-0 overflow-hidden">
      <div className="float-animation absolute top-20 left-10 h-64 w-64 rounded-full bg-orange-300 opacity-20 blur-3xl"></div>
      <div
        className="float-animation absolute right-10 bottom-20 h-96 w-96 rounded-full bg-yellow-300 opacity-20 blur-3xl"
        style={{ animationDelay: '3s' }}
      ></div>
      <div className="pulse-glow absolute top-1/2 left-1/2 h-80 w-80 rounded-full bg-amber-300 opacity-10 blur-3xl"></div>
    </div>
  )
}

export default BackgroundElements
