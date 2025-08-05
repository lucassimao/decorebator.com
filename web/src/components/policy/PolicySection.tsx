import React from 'react'

interface PolicySectionProps {
  title: string
  content: string
}

export const PolicySection: React.FC<PolicySectionProps> = ({ title, content }) => {
  return (
    <section className="mb-8 border-b border-slate-200 py-4 last-of-type:border-b-0">
      <h2 className="mb-4 text-2xl font-semibold text-slate-800">{title}</h2>
      <div
        className="prose prose-slate max-w-none text-slate-700"
        dangerouslySetInnerHTML={{ __html: content }}
      ></div>
    </section>
  )
}

export default PolicySection
