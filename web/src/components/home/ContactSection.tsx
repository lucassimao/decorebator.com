'use client'

import React, { useState } from 'react'

const ContactSection: React.FC = () => {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    subject: '',
    message: '',
  })

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    })
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // Handle form submission here
    console.log('Form submitted:', formData)
  }

  return (
    <section id="contact" className="bg-white py-20">
      <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
        <div className="mb-16 text-center">
          <h2 className="mb-4 text-4xl font-bold lg:text-5xl">
            Get in
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
              {' '}
              Touch
            </span>
          </h2>
          <p className="mx-auto max-w-2xl text-xl text-[#636E72]">
            Have questions about Decorebator? We&apos;d love to hear from you. Send us a message and
            we&apos;ll respond as soon as possible.
          </p>
        </div>

        <div className="rounded-3xl bg-gradient-to-br from-orange-50 to-amber-50 p-8 shadow-xl lg:p-12">
          <form className="space-y-6" onSubmit={handleSubmit}>
            <div className="grid gap-6 md:grid-cols-2">
              {/* Name */}
              <div>
                <label htmlFor="name" className="mb-3 block text-sm font-semibold text-[#2D3436]">
                  Full Name <span className="text-[#FF7B54]">*</span>
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  required
                  value={formData.name}
                  onChange={handleChange}
                  className="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 transition-all duration-300 outline-none focus:border-[#FF7B54] focus:ring-2 focus:ring-[#FF7B54]/20"
                  placeholder="Enter your full name"
                />
              </div>

              {/* Email */}
              <div>
                <label htmlFor="email" className="mb-3 block text-sm font-semibold text-[#2D3436]">
                  Email Address <span className="text-[#FF7B54]">*</span>
                </label>
                <input
                  type="email"
                  id="email"
                  name="email"
                  required
                  value={formData.email}
                  onChange={handleChange}
                  className="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 transition-all duration-300 outline-none focus:border-[#FF7B54] focus:ring-2 focus:ring-[#FF7B54]/20"
                  placeholder="Enter your email address"
                />
              </div>
            </div>

            {/* Subject */}
            <div>
              <label htmlFor="subject" className="mb-3 block text-sm font-semibold text-[#2D3436]">
                Subject <span className="text-[#FF7B54]">*</span>
              </label>
              <select
                id="subject"
                name="subject"
                required
                value={formData.subject}
                onChange={handleChange}
                className="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 transition-all duration-300 outline-none focus:border-[#FF7B54] focus:ring-2 focus:ring-[#FF7B54]/20"
              >
                <option value="">Select a subject...</option>
                <option value="general">General Inquiry</option>
                <option value="support">Technical Support</option>
                <option value="billing">Billing & Subscriptions</option>
                <option value="feature">Feature Request</option>
                <option value="bug">Bug Report</option>
                <option value="partnership">Partnership Opportunity</option>
              </select>
            </div>

            {/* Message */}
            <div>
              <label htmlFor="message" className="mb-3 block text-sm font-semibold text-[#2D3436]">
                Message <span className="text-[#FF7B54]">*</span>
              </label>
              <textarea
                id="message"
                name="message"
                rows={6}
                required
                value={formData.message}
                onChange={handleChange}
                className="w-full resize-none rounded-xl border border-gray-200 bg-white px-4 py-3 transition-all duration-300 outline-none focus:border-[#FF7B54] focus:ring-2 focus:ring-[#FF7B54]/20"
                placeholder="Tell us how we can help you..."
              ></textarea>
            </div>

            {/* Submit Button */}
            <div className="text-center">
              <button
                type="submit"
                className="group transform rounded-full bg-gradient-to-r from-[#FF7B54] to-orange-600 px-10 py-4 text-lg font-semibold text-white transition-all duration-300 hover:scale-105 hover:shadow-2xl"
              >
                <span>Send Message</span>
                <i className="fas fa-paper-plane ml-2 transition-transform group-hover:translate-x-1"></i>
              </button>
            </div>

            {/* Alternative Contact Info */}
            <div className="mt-8 border-t border-orange-200 pt-8">
              <div className="grid gap-6 text-center md:grid-cols-3">
                <div className="flex flex-col items-center">
                  <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-[#FF7B54]">
                    <i className="fas fa-envelope text-white"></i>
                  </div>
                  <h4 className="mb-1 font-semibold text-[#2D3436]">Email Us</h4>
                  <p className="text-sm text-[#636E72]">support@decorebator.com</p>
                </div>

                <div className="flex flex-col items-center">
                  <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-[#4CAF50]">
                    <i className="fas fa-comments text-white"></i>
                  </div>
                  <h4 className="mb-1 font-semibold text-[#2D3436]">Live Chat</h4>
                  <p className="text-sm text-[#636E72]">Available 24/7</p>
                </div>

                <div className="flex flex-col items-center">
                  <div className="mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-[#9C27B0]">
                    <i className="fas fa-question-circle text-white"></i>
                  </div>
                  <h4 className="mb-1 font-semibold text-[#2D3436]">Help Center</h4>
                  <p className="text-sm text-[#636E72]">Find quick answers</p>
                </div>
              </div>
            </div>
          </form>
        </div>
      </div>
    </section>
  )
}

export default ContactSection
