import React from 'react'
import { Feature } from '../../types'

interface FeatureCardProps {
  feature: Feature
  index?: number
}

const FeatureCard: React.FC<FeatureCardProps> = ({ feature, index = 0 }) => {
  const animationDelay = `${index * 0.1}s`

  const cardStyles = [
    {
      bg: 'bg-gradient-to-br from-indigo-50 to-purple-50',
      border: 'border-indigo-100',
      iconBg: 'bg-gradient-to-br from-[#6366F1] to-indigo-600',
      icon: 'fas fa-globe',
      linkColor: 'text-[#6366F1]',
    },
    {
      bg: 'bg-gradient-to-br from-orange-50 to-amber-50',
      border: 'border-orange-100',
      iconBg: 'bg-gradient-to-br from-[#FF7B54] to-orange-600',
      icon: 'fas fa-brain',
      linkColor: 'text-[#FF7B54]',
    },
    {
      bg: 'bg-gradient-to-br from-green-50 to-emerald-50',
      border: 'border-green-100',
      iconBg: 'bg-gradient-to-br from-[#4CAF50] to-green-600',
      icon: 'fas fa-clock',
      linkColor: 'text-[#4CAF50]',
    },
    {
      bg: 'bg-gradient-to-br from-purple-50 to-pink-50',
      border: 'border-purple-100',
      iconBg: 'bg-gradient-to-br from-[#9C27B0] to-purple-600',
      icon: 'fas fa-gamepad',
      linkColor: 'text-[#9C27B0]',
    },
    {
      bg: 'bg-gradient-to-br from-blue-50 to-cyan-50',
      border: 'border-blue-100',
      iconBg: 'bg-gradient-to-br from-[#2196F3] to-blue-600',
      icon: 'fas fa-image',
      linkColor: 'text-[#2196F3]',
    },
    {
      bg: 'bg-gradient-to-br from-yellow-50 to-amber-50',
      border: 'border-yellow-100',
      iconBg: 'bg-gradient-to-br from-[#FFD700] to-yellow-600',
      icon: 'fas fa-headphones',
      linkColor: 'text-yellow-600',
    },
  ]

  const style = cardStyles[index % cardStyles.length]

  return (
    <div
      className={`group rounded-2xl p-8 ${style.bg} border ${style.border} transform transition-all duration-500 hover:-translate-y-2 hover:shadow-2xl`}
      style={{ animationDelay }}
    >
      <div
        className={`h-16 w-16 ${style.iconBg} mb-6 flex items-center justify-center rounded-2xl transition-transform duration-300 group-hover:scale-110`}
      >
        <i className={`${style.icon} text-2xl text-white`}></i>
      </div>
      <h3 className="mb-3 text-xl font-bold">{feature.title}</h3>
      <p className="mb-4 leading-relaxed text-[#636E72]">{feature.description}</p>
      <div
        className={`flex items-center text-sm ${style.linkColor} font-medium transition-transform group-hover:translate-x-2`}
      >
        <span>Learn more</span>
        <i className="fas fa-arrow-right ml-2"></i>
      </div>
    </div>
  )
}

export default FeatureCard
