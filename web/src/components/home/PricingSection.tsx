import React from 'react';
import { CheckCircleIcon } from './icons';
import { getTranslations } from 'next-intl/server';

const PricingSection: React.FC = async () => {
  const t = await getTranslations('pricing');
  const tCommon = await getTranslations('common');

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
  ];

  return (
    <section id="pricing" className="py-20 bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            <span>{t('title.part1')}</span>
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">{t('title.part2')}</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
            {t('subtitle')}
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          {plans.map((plan, index) => (
            <div 
              key={plan.name} 
              className={`relative bg-white rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border-2 ${
                plan.bestValue ? 'border-[#FF7B54] scale-105' : 'border-gray-100'
              } overflow-hidden`}
              style={{ animationDelay: `${index * 0.1}s` }}
            >
              {plan.bestValue && (
                <div className="absolute top-0 left-0 right-0 bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white text-center py-2 text-sm font-semibold">
                  {plan.badge || t('plans.monthly.badge')}
                </div>
              )}
              
              <div className={`p-8 ${plan.bestValue ? 'pt-16' : 'pt-8'}`}>
                <div className="text-center mb-8">
                  <h3 className="text-2xl font-bold text-[#2D3436] mb-2">{plan.name}</h3>
                  <p className="text-[#636E72] text-sm mb-4">{plan.description}</p>
                  <div className="mb-6">
                    <span className="text-4xl font-bold text-[#2D3436]">{plan.price}</span>
                    <span className="text-[#636E72] ml-1">/{plan.frequency}</span>
                  </div>
                </div>

                <ul className="space-y-4 mb-8">
                  {plan.features.map((feature) => (
                    <li key={feature} className="flex items-center">
                      <CheckCircleIcon className="h-5 w-5 text-[#4CAF50] mr-3 flex-shrink-0" />
                      <span className="text-[#636E72]">{feature}</span>
                    </li>
                  ))}
                </ul>

                <a
                  href={plan.href}
                  className={`block w-full py-4 px-6 rounded-xl font-semibold text-center transition-all duration-300 ${
                    plan.bestValue
                      ? 'bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white hover:shadow-2xl transform hover:scale-105'
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
          <div className="bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-2xl p-8 max-w-4xl mx-auto">
            <h3 className="text-xl font-bold text-[#2D3436] mb-4">
              {t('allPlansInclude')}
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 text-sm text-[#636E72]">
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
          
          <p className="text-[#636E72] mt-6">
            {t('footer')}
          </p>
        </div>
      </div>
    </section>
  );
};

export default PricingSection;