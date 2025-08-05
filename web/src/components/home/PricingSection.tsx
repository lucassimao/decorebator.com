import React from 'react'
import { CheckCircleIcon } from './icons'
import { getTranslations } from 'next-intl/server'

const PricingSection: React.FC = async () => {
  const t = await getTranslations('pricing')
  const tCommon = await getTranslations('common')

  const plans = [
    {
      key: 'free',
      name: t('plans.free.name'),
      price: t('plans.free.price'),
      frequency: t('plans.free.period'),
      features: t.raw('plans.free.features') as string[],
      cta: tCommon('downloadApp'),
      bestValue: false,
      href: '#download',
      description: t('plans.free.description'),
    },
    {
      key: 'monthly',
      name: t('plans.monthly.name'),
      price: t('plans.monthly.price'),
      frequency: t('plans.monthly.period'),
      features: t.raw('plans.monthly.features') as string[],
      cta: tCommon('downloadApp'),
      bestValue: true,
      href: '#download',
      description: t('plans.monthly.description'),
      badge: t('plans.monthly.badge'),
    },
    {
      key: 'annual',
      name: t('plans.annual.name'),
      price: t('plans.annual.price'),
      frequency: t('plans.annual.period'),
      features: t.raw('plans.annual.features') as string[],
      cta: tCommon('downloadApp'),
      bestValue: false,
      href: '#download',
      description: t('plans.annual.description'),
    },
  ]

  return (
    <section id="pricing" className="bg-white py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-16 text-center">
          <h2 className="mb-4 text-4xl font-bold lg:text-5xl">
            <span>{t('title.part1')}</span>
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
              {t('title.part2')}
            </span>
          </h2>
          <p className="mx-auto max-w-3xl text-xl text-[#636E72]">{t('subtitle')}</p>
        </div>

        <div className="mx-auto grid max-w-6xl grid-cols-1 gap-8 md:grid-cols-3">
          {plans.map((plan, index) => (
            <div
              key={plan.name}
              className={`relative transform rounded-3xl border-2 bg-white shadow-xl transition-all duration-500 hover:-translate-y-2 hover:shadow-2xl ${
                plan.bestValue ? 'scale-105 border-[#FF7B54]' : 'border-gray-100'
              } overflow-hidden`}
              style={{ animationDelay: `${index * 0.1}s` }}
            >
              {plan.bestValue && (
                <div className="absolute top-0 right-0 left-0 bg-gradient-to-r from-[#FF7B54] to-orange-600 py-2 text-center text-sm font-semibold text-white">
                  {plan.badge || t('plans.monthly.badge')}
                </div>
              )}

              <div className={`p-8 ${plan.bestValue ? 'pt-16' : 'pt-8'}`}>
                <div className="mb-8 text-center">
                  <h3 className="mb-2 text-2xl font-bold text-[#2D3436]">{plan.name}</h3>
                  <p className="mb-4 text-sm text-[#636E72]">{plan.description}</p>
                  <div className="mb-6">
                    <span className="text-4xl font-bold text-[#2D3436]">{plan.price}</span>
                    <span className="ml-1 text-[#636E72]">/{plan.frequency}</span>
                  </div>
                </div>

                <ul className="mb-8 space-y-4">
                  {plan.features.map((feature) => (
                    <li key={feature} className="flex items-center">
                      <CheckCircleIcon className="mr-3 h-5 w-5 flex-shrink-0 text-[#4CAF50]" />
                      <span className="text-[#636E72]">{feature}</span>
                    </li>
                  ))}
                </ul>

                <a
                  href={plan.href}
                  className={`block w-full rounded-xl px-6 py-4 text-center font-semibold transition-all duration-300 ${
                    plan.bestValue
                      ? 'transform bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white hover:scale-105 hover:shadow-2xl'
                      : plan.name === 'Free'
                        ? 'bg-[#4CAF50] text-white hover:bg-green-600 hover:shadow-lg'
                        : 'border-2 border-[#FF7B54] text-[#FF7B54] hover:bg-[#FF7B54] hover:text-white'
                  }`}
                >
                  {plan.cta}
                </a>
              </div>
            </div>
          ))}
        </div>

        <div className="mt-16 text-center">
          <div className="mx-auto max-w-4xl rounded-2xl border border-blue-200 bg-gradient-to-r from-blue-50 to-indigo-50 p-8">
            <h3 className="mb-4 text-xl font-bold text-[#2D3436]">{t('allPlansInclude')}</h3>
            <div className="grid grid-cols-1 gap-6 text-sm text-[#636E72] md:grid-cols-2 lg:grid-cols-4">
              <div className="flex items-center space-x-2">
                <i className="fas fa-brain text-[#FF7B54]"></i>
                <span>{t('features.aiContent')}</span>
              </div>
              <div className="flex items-center space-x-2">
                <i className="fas fa-chart-line text-[#FF7B54]"></i>
                <span>{t('features.spacedRepetition')}</span>
              </div>
              <div className="flex items-center space-x-2">
                <i className="fas fa-mobile-alt text-[#FF7B54]"></i>
                <span>{t('features.mobileAccess')}</span>
              </div>
              <div className="flex items-center space-x-2">
                <i className="fas fa-shield-alt text-[#FF7B54]"></i>
                <span>{t('features.securePrivate')}</span>
              </div>
            </div>
          </div>

          <p className="mt-6 text-[#636E72]">{t('footer')}</p>
        </div>
      </div>
    </section>
  )
}

export default PricingSection
