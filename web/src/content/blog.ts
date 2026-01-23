export type BlogPostSection =
  | { type: 'paragraph'; text: string }
  | { type: 'heading'; text: string }
  | { type: 'list'; items: string[] }
  | { type: 'callout'; title: string; text: string }

export type LocalizedBlogPost = {
  title: string
  description: string
  excerpt: string
  sections: BlogPostSection[]
}

export type BlogPost = {
  slug: string
  date: string
  readingMinutes: number
  tags: string[]
  author: {
    name: string
    role: string
  }
  cover: {
    eyebrow: string
    gradient: string
    image?: string
  }
  content: LocalizedBlogPost
}

export const blogPosts: BlogPost[] = [
  {
    slug: 'word-from-meaning-audio',
    date: '2026-01-22',
    readingMinutes: 3,
    tags: ['Listening', 'Quiz modes', 'Audio'],
    author: {
      name: 'Decorebator Team',
      role: 'Product',
    },
    cover: {
      eyebrow: 'New quiz mode',
      gradient: 'from-primary-500 via-amber-500 to-rose-500',
    },
    content: {
      title: 'New quiz: Word from Meaning (Audio)',
      description:
        'Hear the meaning, then choose the correct word. It is a fast way to strengthen listening comprehension and recall.',
      excerpt:
        'A new listening quiz that plays the meaning out loud and asks you to pick the right word.',
      sections: [
        {
          type: 'paragraph',
          text: 'We just added a new quiz mode that flips the usual prompt. Instead of seeing the meaning on screen, you hear it. Then you select the word that matches the meaning you heard.',
        },
        {
          type: 'heading',
          text: 'Why this helps',
        },
        {
          type: 'list',
          items: [
            'Trains listening comprehension for meanings, not just word sounds.',
            'Improves recall under time pressure by forcing quick recognition.',
            'Balances visual-heavy practice with audio-driven practice.',
          ],
        },
        {
          type: 'heading',
          text: 'How it works',
        },
        {
          type: 'list',
          items: [
            'We generate a short audio clip for each definition meaning.',
            'When a meaning audio clip is available, the quiz becomes eligible.',
            'You hear the meaning and choose the right word from the options.',
          ],
        },
        {
          type: 'callout',
          title: 'Tip',
          text: 'Try this mode after a few reviews of a word. It is great for reinforcing long-term memory.',
        },
      ],
    },
  },
]

export const getBlogPost = (slug: string): BlogPost | null => {
  const post = blogPosts.find((entry) => entry.slug === slug)
  return post ?? null
}

export const getBlogPosts = (): BlogPost[] => blogPosts
