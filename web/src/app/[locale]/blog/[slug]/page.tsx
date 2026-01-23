import React from 'react'
import Link from 'next/link'
import Image from 'next/image'
import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { setRequestLocale } from 'next-intl/server'
import { getTranslations } from 'next-intl/server'
import PageLayout from '../../../../components/layout/PageLayout'
import { getBlogPost, blogPosts, type BlogPost } from '../../../../content/blog'
import { routing } from '../../../../i18n/routing'

function ShareButtons({ url, title }: { url: string; title: string }) {
  const encodedUrl = encodeURIComponent(url)
  const encodedTitle = encodeURIComponent(title)

  return (
    <div className="flex items-center gap-3">
      <span className="text-sm font-medium text-slate-500">Share:</span>
      <a
        href={`https://twitter.com/intent/tweet?url=${encodedUrl}&text=${encodedTitle}`}
        target="_blank"
        rel="noopener noreferrer"
        className="flex h-9 w-9 items-center justify-center rounded-full bg-slate-100 text-slate-600 transition-colors hover:bg-slate-200 hover:text-slate-900"
        aria-label="Share on Twitter"
      >
        <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
        </svg>
      </a>
      <a
        href={`https://www.linkedin.com/sharing/share-offsite/?url=${encodedUrl}`}
        target="_blank"
        rel="noopener noreferrer"
        className="flex h-9 w-9 items-center justify-center rounded-full bg-slate-100 text-slate-600 transition-colors hover:bg-slate-200 hover:text-slate-900"
        aria-label="Share on LinkedIn"
      >
        <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z" />
        </svg>
      </a>
      <a
        href={`mailto:?subject=${encodedTitle}&body=${encodedUrl}`}
        className="flex h-9 w-9 items-center justify-center rounded-full bg-slate-100 text-slate-600 transition-colors hover:bg-slate-200 hover:text-slate-900"
        aria-label="Share via email"
      >
        <svg
          className="h-4 w-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
          />
        </svg>
      </a>
    </div>
  )
}

function ArticleStructuredData({ post, locale }: { post: BlogPost; locale: string }) {
  const articleData: Record<string, unknown> = {
    '@context': 'https://schema.org',
    '@type': 'Article',
    headline: post.content.title,
    description: post.content.description,
    datePublished: post.date,
    dateModified: post.date,
    author: {
      '@type': 'Organization',
      name: post.author.name,
    },
    publisher: {
      '@type': 'Organization',
      name: 'Decorebator',
      logo: {
        '@type': 'ImageObject',
        url: 'https://decorebator.com/images/logo.png',
      },
    },
    mainEntityOfPage: {
      '@type': 'WebPage',
      '@id': `https://decorebator.com/${locale}/blog/${post.slug}`,
    },
  }

  if (post.cover.image) {
    articleData.image = post.cover.image.startsWith('/')
      ? `https://decorebator.com${post.cover.image}`
      : post.cover.image
  }

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(articleData) }}
    />
  )
}

interface BlogPostPageProps {
  params: Promise<{
    locale: string
    slug: string
  }>
}

const BlogPostPage: React.FC<BlogPostPageProps> = async ({ params }) => {
  const { locale, slug } = await params
  setRequestLocale(locale)

  const t = await getTranslations('blog')
  const post = getBlogPost(slug)

  if (!post) {
    notFound()
  }

  const publishedDate = new Date(post.date).toLocaleDateString(locale, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })

  const postUrl = `https://decorebator.com/${locale}/blog/${post.slug}`

  return (
    <PageLayout>
      <ArticleStructuredData post={post} locale={locale} />
      <div className="min-h-screen pt-24 pb-20">
        <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
          <Link
            href={`/${locale}/blog`}
            className="text-primary-600 mb-6 inline-flex items-center text-sm font-semibold"
          >
            ← {t('backToBlog')}
          </Link>

          <article className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-xl shadow-slate-900/5">
            {post.cover.image && (
              <div className="relative h-64 w-full sm:h-80">
                <Image
                  src={post.cover.image}
                  alt={post.content.title}
                  fill
                  className="object-cover"
                  priority
                />
              </div>
            )}
            <div className="p-8">
              <div className="mb-6 flex flex-wrap items-center gap-3">
                <span className="rounded-full bg-slate-900 px-3 py-1 text-xs font-semibold text-white">
                  {post.cover.eyebrow}
                </span>
                <span className="text-xs font-semibold tracking-[0.15em] text-slate-500 uppercase">
                  {post.tags.join(' · ')}
                </span>
              </div>

              <h1 className="mb-4 text-3xl font-bold text-slate-900 sm:text-4xl">
                {post.content.title}
              </h1>
              <p className="mb-6 text-lg text-slate-600">{post.content.description}</p>

              <div className="mb-8 flex flex-wrap items-center justify-between gap-4 border-b border-slate-100 pb-6">
                <div className="flex flex-wrap items-center gap-4 text-sm text-slate-500">
                  <span>{t('publishedOn', { date: publishedDate })}</span>
                  <span>•</span>
                  <span>{t('minutes', { count: post.readingMinutes })}</span>
                </div>
                <ShareButtons url={postUrl} title={post.content.title} />
              </div>

              <div className="space-y-6 text-slate-700">
                {post.content.sections.map((section, index) => {
                  const sectionKey =
                    section.type === 'list'
                      ? `${section.type}-${section.items.join('|')}-${index}`
                      : `${section.type}-${section.text}-${index}`
                  if (section.type === 'heading') {
                    return (
                      <h2 key={sectionKey} className="text-xl font-semibold text-slate-900">
                        {section.text}
                      </h2>
                    )
                  }

                  if (section.type === 'paragraph') {
                    return (
                      <p key={sectionKey} className="leading-relaxed">
                        {section.text}
                      </p>
                    )
                  }

                  if (section.type === 'list') {
                    return (
                      <ul key={sectionKey} className="space-y-3">
                        {section.items.map((item, itemIndex) => (
                          <li key={itemIndex} className="flex gap-3">
                            <span className="bg-primary-500 mt-1 h-2 w-2 rounded-full" />
                            <span>{item}</span>
                          </li>
                        ))}
                      </ul>
                    )
                  }

                  if (section.type === 'callout') {
                    return (
                      <div
                        key={sectionKey}
                        className="border-primary-100 bg-primary-50 rounded-2xl border p-5"
                      >
                        <h3 className="text-primary-600 mb-2 text-sm font-semibold tracking-[0.2em] uppercase">
                          {section.title}
                        </h3>
                        <p className="text-sm text-slate-700">{section.text}</p>
                      </div>
                    )
                  }

                  return null
                })}
              </div>

              {/* Share + CTA section at end of article */}
              <div className="mt-10 border-t border-slate-100 pt-8">
                <div className="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
                  <ShareButtons url={postUrl} title={post.content.title} />
                </div>
              </div>
            </div>
          </article>

          {/* CTA Card */}
          <div className="border-primary-100 from-primary-50 mt-8 rounded-3xl border bg-gradient-to-br via-white to-amber-50 p-8">
            <div className="flex flex-col items-center text-center sm:flex-row sm:text-left">
              <div className="from-primary-500 mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br to-amber-500 text-2xl text-white sm:mr-6 sm:mb-0">
                🎧
              </div>
              <div className="flex-1">
                <h3 className="mb-2 text-xl font-semibold text-slate-900">{t('cta.title')}</h3>
                <p className="mb-4 text-slate-600 sm:mb-0">{t('cta.subtitle')}</p>
              </div>
              <Link
                href={`/${locale}/#download`}
                className="bg-primary-600 hover:bg-primary-700 inline-flex items-center justify-center rounded-full px-6 py-3 text-sm font-semibold text-white transition"
              >
                {t('cta.button')}
              </Link>
            </div>
          </div>
        </div>
      </div>
    </PageLayout>
  )
}

export async function generateStaticParams() {
  return blogPosts.flatMap((post) =>
    routing.locales.map((locale) => ({
      locale,
      slug: post.slug,
    }))
  )
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string; slug: string }>
}): Promise<Metadata> {
  const { slug } = await params
  const post = getBlogPost(slug)
  if (!post) {
    return {}
  }

  const metadata: Metadata = {
    title: `${post.content.title} | Decorebator Blog`,
    description: post.content.description,
    authors: [{ name: post.author.name }],
  }

  if (post.cover.image) {
    const imageUrl = post.cover.image.startsWith('/')
      ? `https://decorebator.com${post.cover.image}`
      : post.cover.image
    metadata.openGraph = {
      images: [{ url: imageUrl, alt: post.content.title }],
    }
    metadata.twitter = {
      card: 'summary_large_image',
      images: [imageUrl],
    }
  }

  return metadata
}

export default BlogPostPage
