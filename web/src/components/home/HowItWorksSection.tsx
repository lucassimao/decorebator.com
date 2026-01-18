import React from 'react'
import { getTranslations } from 'next-intl/server'
import { PlusCircle, Gamepad2, Brain, CheckCircle2, ChevronRight, Download } from 'lucide-react'

const HowItWorksSection: React.FC = async () => {
  const t = await getTranslations('howItWorks')
  const tCommon = await getTranslations('common')

  return (
    <section id="how-it-works" className="bg-white py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-16 text-center">
          <h2 className="mb-4 text-4xl font-bold lg:text-5xl">
            <span>{t('title.part1')}</span>
            <span className="from-primary-500 to-accent-500 bg-gradient-to-r bg-clip-text text-transparent">
              {t('title.part2')}
            </span>
          </h2>
          <p className="mx-auto max-w-3xl text-xl text-slate-600">{t('subtitle')}</p>
        </div>

        <div className="relative grid gap-8 lg:grid-cols-3">
          {/* Step 1 */}
          <div className="relative">
            <div className="from-primary-500 to-primary-600 shadow-primary-500/30 absolute -top-4 -left-4 z-10 flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br text-xl font-bold text-white shadow-lg">
              1
            </div>
            <div className="border-primary-100 from-primary-50/50 hover:shadow-primary-500/10 h-full rounded-2xl border bg-gradient-to-br to-orange-50/50 p-8 transition-all duration-300 hover:shadow-xl">
              <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-white shadow-lg">
                <PlusCircle className="text-primary-500 h-8 w-8" strokeWidth={2} />
              </div>
              <h3 className="mb-4 text-2xl font-bold text-slate-900">{t('step1.title')}</h3>
              <p className="mb-6 text-slate-600">{t('step1.description')}</p>
              <ul className="space-y-3">
                {(t.raw('step1.features') as string[]).map((feature, index) => (
                  <li key={index} className="flex items-start">
                    <CheckCircle2
                      className="text-success-500 mt-0.5 mr-3 h-5 w-5 flex-shrink-0"
                      strokeWidth={2}
                    />
                    <span className="text-sm text-slate-700">{feature}</span>
                  </li>
                ))}
              </ul>
            </div>
            {/* Connector */}
            <div className="absolute top-1/2 -right-4 z-0 hidden -translate-y-1/2 transform lg:block">
              <ChevronRight className="text-primary-300 h-10 w-10" strokeWidth={2} />
            </div>
          </div>

          {/* Step 2 */}
          <div className="relative">
            <div className="from-accent-500 to-accent-600 shadow-accent-500/30 absolute -top-4 -left-4 z-10 flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br text-xl font-bold text-white shadow-lg">
              2
            </div>
            <div className="border-accent-100 from-accent-50/50 hover:shadow-accent-500/10 h-full rounded-2xl border bg-gradient-to-br to-purple-50/50 p-8 transition-all duration-300 hover:shadow-xl">
              <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-white shadow-lg">
                <Gamepad2 className="text-accent-500 h-8 w-8" strokeWidth={2} />
              </div>
              <h3 className="mb-4 text-2xl font-bold text-slate-900">{t('step2.title')}</h3>
              <p className="mb-6 text-slate-600">{t('step2.description')}</p>
              <ul className="space-y-3">
                {(t.raw('step2.features') as string[]).map((feature, index) => (
                  <li key={index} className="flex items-start">
                    <CheckCircle2
                      className="text-success-500 mt-0.5 mr-3 h-5 w-5 flex-shrink-0"
                      strokeWidth={2}
                    />
                    <span className="text-sm text-slate-700">{feature}</span>
                  </li>
                ))}
              </ul>
            </div>
            {/* Connector */}
            <div className="absolute top-1/2 -right-4 z-0 hidden -translate-y-1/2 transform lg:block">
              <ChevronRight className="text-accent-300 h-10 w-10" strokeWidth={2} />
            </div>
          </div>

          {/* Step 3 */}
          <div className="relative">
            <div className="from-success-500 to-success-600 shadow-success-500/30 absolute -top-4 -left-4 z-10 flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br text-xl font-bold text-white shadow-lg">
              3
            </div>
            <div className="border-success-100 from-success-50/50 hover:shadow-success-500/10 h-full rounded-2xl border bg-gradient-to-br to-emerald-50/50 p-8 transition-all duration-300 hover:shadow-xl">
              <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-white shadow-lg">
                <Brain className="text-success-500 h-8 w-8" strokeWidth={2} />
              </div>
              <h3 className="mb-4 text-2xl font-bold text-slate-900">{t('step3.title')}</h3>
              <p className="mb-6 text-slate-600">{t('step3.description')}</p>
              <ul className="space-y-3">
                {(t.raw('step3.features') as string[]).map((feature, index) => (
                  <li key={index} className="flex items-start">
                    <CheckCircle2
                      className="text-success-500 mt-0.5 mr-3 h-5 w-5 flex-shrink-0"
                      strokeWidth={2}
                    />
                    <span className="text-sm text-slate-700">{feature}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>

        {/* CTA */}
        <div className="mt-16 text-center">
          <a
            href="#download"
            className="group from-primary-500 to-primary-600 shadow-primary-500/30 hover:shadow-primary-500/40 inline-flex items-center gap-2 rounded-full bg-gradient-to-r px-10 py-4 text-lg font-semibold text-white shadow-lg transition-all duration-300 hover:scale-105 hover:shadow-xl"
          >
            <span>{tCommon('downloadApp')}</span>
            <Download
              className="h-5 w-5 transition-transform group-hover:translate-y-0.5"
              strokeWidth={2.5}
            />
          </a>
        </div>
      </div>
    </section>
  )
}

export default HowItWorksSection
