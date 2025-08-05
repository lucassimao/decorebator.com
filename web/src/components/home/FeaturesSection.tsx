import React from 'react'
import FeatureCard from './FeatureCard'
import { Feature } from '../../types'

interface FeaturesSectionProps {
  features: Feature[]
}

const FeaturesSection: React.FC<FeaturesSectionProps> = ({ features }) => {
  return (
    <section id="features" className="bg-white py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-16 text-center">
          <h2 className="mb-4 text-4xl font-bold lg:text-5xl">
            Everything You Need to
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
              {' '}
              Master Vocabulary
            </span>
          </h2>
          <p className="mx-auto max-w-3xl text-xl text-[#636E72]">
            Our comprehensive platform combines cutting-edge AI technology with proven learning
            science to accelerate your language journey.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
          {features.map((feature, index) => (
            <FeatureCard key={feature.title} feature={feature} index={index} />
          ))}
        </div>
      </div>
    </section>
  )
}

export default FeaturesSection
