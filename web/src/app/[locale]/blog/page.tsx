import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Metadata } from 'next'
import { setRequestLocale } from 'next-intl/server'
import { getTranslations } from 'next-intl/server'
import PageLayout from '../../../components/layout/PageLayout'
import { getBlogPosts } from '../../../content/blog'

interface BlogPageProps {
  params: Promise<{
    locale: string
  }>
}

const BlogPage: React.FC<BlogPageProps> = async ({ params }) => {
  const { locale } = await params
  setRequestLocale(locale)

  const t = await getTranslations('blog')
  const posts = getBlogPosts()
  const featuredPost = posts[0]

  return (
    <PageLayout>
      <div className="min-h-screen pt-24 pb-20">
        <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
          <div className="mb-12 overflow-hidden rounded-[32px] border border-slate-200 bg-white shadow-2xl shadow-slate-900/10">
            <div className="relative overflow-hidden px-8 pt-10 pb-12 sm:px-12">
              <div
                className="bg-primary-200/40 pointer-events-none absolute -top-24 right-10 h-64 w-64 rounded-full blur-3xl"
                aria-hidden="true"
              />
              <div className="pointer-events-none absolute -bottom-24 left-6 h-64 w-64 rounded-full bg-amber-200/40 blur-3xl" />
              <div className="relative">
                <div className="border-primary-200/60 text-primary-600 mb-4 inline-flex items-center gap-2 rounded-full border bg-white/80 px-4 py-1 text-xs font-semibold tracking-[0.2em] uppercase shadow-sm">
                  {t('title')}
                </div>
                <h1 className="mb-4 text-4xl font-bold text-slate-900 lg:text-5xl">{t('title')}</h1>
                <p className="max-w-2xl text-lg text-slate-600">{t('subtitle')}</p>
              </div>
            </div>
          </div>

          <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
            {posts.map((post) => {
              const dateLabel = new Date(post.date).toLocaleDateString(locale, {
                year: 'numeric',
                month: 'short',
                day: 'numeric',
              })

              return (
                <Link
                  key={post.slug}
                  href={`/${locale}/blog/${post.slug}`}
                  className="group relative overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-xl shadow-slate-900/5 transition-all duration-200 hover:-translate-y-1 hover:shadow-2xl hover:shadow-slate-900/10"
                >
                  {post.cover.image ? (
                    <div className="relative h-48 w-full overflow-hidden">
                      <Image
                        src={post.cover.image}
                        alt={post.content.title}
                        fill
                        className="object-cover transition-transform duration-300 group-hover:scale-105"
                      />
                      <div
                        className={`absolute inset-0 bg-gradient-to-t from-white via-transparent to-transparent`}
                      />
                    </div>
                  ) : (
                    <div
                      className={`absolute inset-0 bg-gradient-to-br opacity-15 blur-2xl transition-opacity duration-200 group-hover:opacity-30 ${post.cover.gradient}`}
                      aria-hidden="true"
                    />
                  )}
                  <div className="relative flex h-full flex-col justify-between p-8">
                    <div>
                      <div className="mb-6 flex flex-wrap items-center gap-3">
                        <span className="rounded-full bg-slate-900 px-3 py-1 text-xs font-semibold text-white">
                          {post.cover.eyebrow}
                        </span>
                        <div className="flex flex-wrap items-center gap-2">
                          {post.tags.map((tag) => (
                            <span
                              key={tag}
                              className="rounded-full border border-slate-200 bg-white/80 px-2.5 py-0.5 text-[11px] font-semibold tracking-[0.18em] text-slate-500 uppercase"
                            >
                              {tag}
                            </span>
                          ))}
                        </div>
                      </div>
                      <h2 className="group-hover:text-primary-600 mb-3 text-2xl font-semibold text-slate-900">
                        {post.content.title}
                      </h2>
                      <p className="mb-6 text-sm text-slate-600 sm:text-base">
                        {post.content.description}
                      </p>
                    </div>
                    <div className="flex flex-wrap items-center justify-between gap-3 text-sm text-slate-500">
                      <span>{t('minutes', { count: post.readingMinutes })}</span>
                      <span>{post.author.name}</span>
                      <span>{dateLabel}</span>
                      <span className="text-primary-600 font-semibold">{t('readMore')}</span>
                    </div>
                  </div>
                </Link>
              )
            })}

            <div className="flex flex-col gap-6">
              {posts.length > 1 && featuredPost && (
                <div className="rounded-3xl border border-slate-200 bg-gradient-to-br from-slate-50 via-white to-slate-100 p-8 shadow-xl shadow-slate-900/5">
                  <h3 className="mb-3 text-xl font-semibold text-slate-900">{t('latestPosts')}</h3>
                  <p className="text-sm text-slate-600">{featuredPost.content.excerpt}</p>
                  <div className="mt-6 flex flex-wrap gap-2">
                    {featuredPost.tags.map((tag) => (
                      <span
                        key={tag}
                        className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-medium text-slate-600"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
              )}
              <div className="border-primary-100 bg-primary-50 rounded-3xl border p-8">
                <h3 className="mb-3 text-xl font-semibold text-slate-900">{t('cta.title')}</h3>
                <p className="text-sm text-slate-600">{t('cta.subtitle')}</p>
                <Link
                  href={`/${locale}/#download`}
                  className="bg-primary-600 hover:bg-primary-700 mt-5 inline-flex items-center justify-center rounded-full px-5 py-2 text-sm font-semibold text-white transition"
                >
                  {t('cta.button')}
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </PageLayout>
  )
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>
}): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'blog' })

  return {
    title: `Decorebator Blog | ${t('title')}`,
    description: t('subtitle'),
  }
}

export default BlogPage
