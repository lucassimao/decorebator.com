import PageLayout from '@/components/layout/PageLayout'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'

export default async function DeleteAccountPage({
  params,
}: {
  params: Promise<{ locale: string }>
}) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'accountDeletion' })

  return (
    <PageLayout>
      <div className="relative overflow-hidden bg-[linear-gradient(180deg,#fffaf1_0%,#fff_72%)] pt-28 pb-20 text-[#25302d] sm:pt-36">
        <div
          className="pointer-events-none absolute -top-40 right-[-12rem] h-[32rem] w-[32rem] rounded-full bg-[#e26f4e]/15 blur-3xl"
          aria-hidden="true"
        />
        <div className="relative mx-auto w-[min(100%-2rem,48rem)]">
          <section className="border-b border-[#d9c9ad] pb-10" aria-labelledby="page-title">
            <p className="font-mono text-xs font-bold tracking-[0.16em] text-[#8f2e1b] uppercase">
              {t('eyebrow')}
            </p>
            <h1
              id="page-title"
              className="mt-3 max-w-[12ch] font-serif text-[clamp(2.75rem,11vw,5rem)] leading-[0.98] font-medium tracking-[-0.045em]"
            >
              {t('title')}
            </h1>
            <p className="mt-5 max-w-2xl text-base leading-7 text-[#596560] sm:text-lg">
              {t('lede')}
            </p>
          </section>

          <aside
            className="mt-8 rounded-xl border border-l-[0.4rem] border-[#9c5b16] bg-[#fff0d6] px-5 py-4"
            aria-labelledby="subscription-warning"
          >
            <h2 id="subscription-warning" className="text-base leading-6 font-bold">
              {t('warning.title')}
            </h2>
            <p className="mt-1 leading-7 text-[#3f4b47]">{t('warning.body')}</p>
          </aside>

          <section className="mt-8 grid gap-4 md:grid-cols-2" aria-label={t('routesLabel')}>
            <article className="relative overflow-hidden rounded-2xl border border-[#d9c9ad] bg-white/85 p-6 shadow-[0_1rem_3rem_rgb(72_53_31_/_8%)]">
              <div
                className="absolute top-0 right-5 h-9 w-6 bg-[#b83e25] [clip-path:polygon(0_0,100%_0,100%_100%,50%_75%,0_100%)]"
                aria-hidden="true"
              />
              <p className="font-mono text-xs font-bold tracking-[0.14em] text-[#8f2e1b] uppercase">
                {t('inApp.eyebrow')}
              </p>
              <h2 className="mt-2 font-serif text-2xl leading-tight font-bold">
                {t('inApp.title')}
              </h2>
              <ol className="mt-4 list-decimal space-y-1 pl-5 leading-7 text-[#596560]">
                <li>{t('inApp.step1')}</li>
                <li>{t('inApp.step2')}</li>
                <li>{t('inApp.step3')}</li>
              </ol>
              <p className="mt-5 leading-7 text-[#596560]">{t('inApp.success')}</p>
            </article>

            <article className="relative overflow-hidden rounded-2xl border border-[#d9c9ad] bg-white/85 p-6 shadow-[0_1rem_3rem_rgb(72_53_31_/_8%)]">
              <div
                className="absolute top-0 right-5 h-9 w-6 bg-[#b83e25] [clip-path:polygon(0_0,100%_0,100%_100%,50%_75%,0_100%)]"
                aria-hidden="true"
              />
              <p className="font-mono text-xs font-bold tracking-[0.14em] text-[#8f2e1b] uppercase">
                {t('email.eyebrow')}
              </p>
              <h2 className="mt-2 font-serif text-2xl leading-tight font-bold">
                {t('email.title')}
              </h2>
              <p className="mt-4 leading-7 text-[#596560]">{t('email.instructions')}</p>
              <p className="mt-4 leading-7 text-[#596560]">
                {t.rich('email.verification', {
                  privacyPolicy: (chunks) => (
                    <Link
                      href={`/${locale}/privacy`}
                      className="font-semibold text-[#8f2e1b] underline decoration-[#b83e25]/50 underline-offset-4 focus-visible:rounded focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#165d64]"
                    >
                      {chunks}
                    </Link>
                  ),
                })}
              </p>
              <a
                href="mailto:privacy@decorebator.com"
                className="mt-4 block w-fit font-bold text-[#8f2e1b] underline decoration-[#b83e25]/50 underline-offset-4 focus-visible:rounded focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#165d64]"
              >
                {t('email.address')}
              </a>
              <a
                href="mailto:privacy@decorebator.com?subject=Decorebator%20account%20deletion%20request"
                className="mt-4 inline-flex min-h-12 items-center justify-center rounded-xl bg-[#8f2e1b] px-5 py-3 font-bold text-white transition-colors hover:bg-[#702313] focus-visible:outline-3 focus-visible:outline-offset-4 focus-visible:outline-[#165d64]"
              >
                {t('email.button')}
              </a>
            </article>
          </section>

          <section className="mt-8 border-y border-[#d9c9ad] py-7" aria-labelledby="outcome-title">
            <p className="font-mono text-xs font-bold tracking-[0.14em] text-[#8f2e1b] uppercase">
              {t('outcome.eyebrow')}
            </p>
            <h2 id="outcome-title" className="mt-2 font-serif text-2xl leading-tight font-bold">
              {t('outcome.title')}
            </h2>
            <ul className="mt-4 list-disc space-y-2 pl-5 leading-7 text-[#596560]">
              <li>{t('outcome.profile')}</li>
              <li>{t('outcome.retention')}</li>
              <li>{t('outcome.refund')}</li>
            </ul>
          </section>

          <nav className="mt-7 flex flex-wrap gap-x-6 gap-y-3 text-sm" aria-label={t('help')}>
            <Link
              href={`/${locale}/privacy`}
              className="text-[#8f2e1b] underline underline-offset-4"
            >
              {t('privacyPolicy')}
            </Link>
            <Link href={`/${locale}/terms`} className="text-[#8f2e1b] underline underline-offset-4">
              {t('termsOfService')}
            </Link>
            <a
              href="mailto:privacy@decorebator.com"
              className="text-[#8f2e1b] underline underline-offset-4"
            >
              {t('help')}: {t('email.address')}
            </a>
          </nav>
        </div>
      </div>
    </PageLayout>
  )
}
