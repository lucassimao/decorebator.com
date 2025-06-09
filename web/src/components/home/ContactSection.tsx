'use client';

import React, { useState } from 'react';

const ContactSection: React.FC = () => {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    subject: '',
    message: ''
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // Handle form submission here
    console.log('Form submitted:', formData);
  };

  return (
    <section id="contact" className="py-20 bg-white">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            Get in 
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent"> Touch</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-2xl mx-auto">
            Have questions about Decorebator? We&apos;d love to hear from you. Send us a message and we&apos;ll respond as soon as possible.
          </p>
        </div>

        <div className="bg-gradient-to-br from-orange-50 to-amber-50 rounded-3xl p-8 lg:p-12 shadow-xl">
          <form className="space-y-6" onSubmit={handleSubmit}>
            <div className="grid md:grid-cols-2 gap-6">
              {/* Name */}
              <div>
                <label htmlFor="name" className="block text-sm font-semibold text-[#2D3436] mb-3">
                  Full Name <span className="text-[#FF7B54]">*</span>
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  required
                  value={formData.name}
                  onChange={handleChange}
                  className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:border-[#FF7B54] focus:ring-2 focus:ring-[#FF7B54]/20 outline-none transition-all duration-300 bg-white"
                  placeholder="Enter your full name"
                />
              </div>

              {/* Email */}
              <div>
                <label htmlFor="email" className="block text-sm font-semibold text-[#2D3436] mb-3">
                  Email Address <span className="text-[#FF7B54]">*</span>
                </label>
                <input
                  type="email"
                  id="email"
                  name="email"
                  required
                  value={formData.email}
                  onChange={handleChange}
                  className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:border-[#FF7B54] focus:ring-2 focus:ring-[#FF7B54]/20 outline-none transition-all duration-300 bg-white"
                  placeholder="Enter your email address"
                />
              </div>
            </div>

            {/* Subject */}
            <div>
              <label htmlFor="subject" className="block text-sm font-semibold text-[#2D3436] mb-3">
                Subject <span className="text-[#FF7B54]">*</span>
              </label>
              <select
                id="subject"
                name="subject"
                required
                value={formData.subject}
                onChange={handleChange}
                className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:border-[#FF7B54] focus:ring-2 focus:ring-[#FF7B54]/20 outline-none transition-all duration-300 bg-white"
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
              <label htmlFor="message" className="block text-sm font-semibold text-[#2D3436] mb-3">
                Message <span className="text-[#FF7B54]">*</span>
              </label>
              <textarea
                id="message"
                name="message"
                rows={6}
                required
                value={formData.message}
                onChange={handleChange}
                className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:border-[#FF7B54] focus:ring-2 focus:ring-[#FF7B54]/20 outline-none transition-all duration-300 bg-white resize-none"
                placeholder="Tell us how we can help you..."
              ></textarea>
            </div>

            {/* Submit Button */}
            <div className="text-center">
              <button
                type="submit"
                className="group bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-10 py-4 rounded-full font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300"
              >
                <span>Send Message</span>
                <i className="fas fa-paper-plane ml-2 group-hover:translate-x-1 transition-transform"></i>
              </button>
            </div>

            {/* Alternative Contact Info */}
            <div className="border-t border-orange-200 pt-8 mt-8">
              <div className="grid md:grid-cols-3 gap-6 text-center">
                <div className="flex flex-col items-center">
                  <div className="w-12 h-12 bg-[#FF7B54] rounded-full flex items-center justify-center mb-3">
                    <i className="fas fa-envelope text-white"></i>
                  </div>
                  <h4 className="font-semibold text-[#2D3436] mb-1">Email Us</h4>
                  <p className="text-[#636E72] text-sm">support@decorebator.com</p>
                </div>
                
                <div className="flex flex-col items-center">
                  <div className="w-12 h-12 bg-[#4CAF50] rounded-full flex items-center justify-center mb-3">
                    <i className="fas fa-comments text-white"></i>
                  </div>
                  <h4 className="font-semibold text-[#2D3436] mb-1">Live Chat</h4>
                  <p className="text-[#636E72] text-sm">Available 24/7</p>
                </div>
                
                <div className="flex flex-col items-center">
                  <div className="w-12 h-12 bg-[#9C27B0] rounded-full flex items-center justify-center mb-3">
                    <i className="fas fa-question-circle text-white"></i>
                  </div>
                  <h4 className="font-semibold text-[#2D3436] mb-1">Help Center</h4>
                  <p className="text-[#636E72] text-sm">Find quick answers</p>
                </div>
              </div>
            </div>
          </form>
        </div>
      </div>
    </section>
  );
};

export default ContactSection;