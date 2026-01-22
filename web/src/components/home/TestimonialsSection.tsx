import React from 'react'
import Image from 'next/image'
import { statsConfig } from '@/config/statsConfig'

const TestimonialsSection: React.FC = () => {
  return (
    <section id="testimonials" className="bg-gradient-to-br from-orange-50 to-amber-50 py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-16 text-center">
          <h2 className="mb-4 text-4xl font-bold lg:text-5xl">
            Loved by
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
              {' '}
              Language Learners
            </span>
          </h2>
          <p className="mx-auto max-w-3xl text-xl text-[#636E72]">
            Join thousands of satisfied learners who have transformed their language skills with
            Decorebator.
          </p>
        </div>

        <div className="grid gap-8 lg:grid-cols-3">
          {/* Testimonial 1 */}
          <div className="transform rounded-3xl bg-white p-8 shadow-xl transition-all duration-300 hover:-translate-y-2 hover:shadow-2xl">
            <div className="mb-4 flex items-center">
              <Image
                src="https://ui-avatars.com/api/?name=Sarah+Chen&background=FF7B54&color=fff&size=48&rounded=true"
                alt="Sarah Chen"
                width={48}
                height={48}
                className="mr-4 h-12 w-12 rounded-full"
              />
              <div>
                <div className="font-bold">Sarah Chen</div>
                <div className="text-sm text-[#636E72]">Language Student</div>
              </div>
            </div>
            <div className="mb-4 flex">
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
            </div>
            <p className="leading-relaxed text-[#636E72] italic">
              &quot;Decorebator transformed my Spanish learning. The AI-generated images help me
              remember words so much better than traditional methods. I&apos;ve learned more in 3
              months than in years of traditional study!&quot;
            </p>
          </div>

          {/* Testimonial 2 */}
          <div className="transform rounded-3xl bg-white p-8 shadow-xl transition-all duration-300 hover:-translate-y-2 hover:shadow-2xl">
            <div className="mb-4 flex items-center">
              <Image
                src="https://ui-avatars.com/api/?name=Marcus+Rodriguez&background=4CAF50&color=fff&size=48&rounded=true"
                alt="Marcus Rodriguez"
                width={48}
                height={48}
                className="mr-4 h-12 w-12 rounded-full"
              />
              <div>
                <div className="font-bold">Marcus Rodriguez</div>
                <div className="text-sm text-[#636E72]">Professional Translator</div>
              </div>
            </div>
            <div className="mb-4 flex">
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
            </div>
            <p className="leading-relaxed text-[#636E72] italic">
              &quot;The spaced repetition system is incredibly effective. I&apos;ve expanded my
              technical vocabulary by 300% in just 3 months. The 9 quiz modes keep learning fresh
              and engaging every day.&quot;
            </p>
          </div>

          {/* Testimonial 3 */}
          <div className="transform rounded-3xl bg-white p-8 shadow-xl transition-all duration-300 hover:-translate-y-2 hover:shadow-2xl">
            <div className="mb-4 flex items-center">
              <Image
                src="https://ui-avatars.com/api/?name=Emma+Thompson&background=9C27B0&color=fff&size=48&rounded=true"
                alt="Emma Thompson"
                width={48}
                height={48}
                className="mr-4 h-12 w-12 rounded-full"
              />
              <div>
                <div className="font-bold">Emma Thompson</div>
                <div className="text-sm text-[#636E72]">ESL Teacher</div>
              </div>
            </div>
            <div className="mb-4 flex">
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
            </div>
            <p className="leading-relaxed text-[#636E72] italic">
              &quot;I recommend Decorebator to all my students. The variety of quiz modes keeps them
              engaged and the progress tracking is excellent. It&apos;s the perfect blend of
              technology and pedagogy.&quot;
            </p>
          </div>
        </div>

        {/* Success Metrics */}
        {statsConfig.locations.testimonials.showStats && (
          <div className="mt-16 rounded-3xl bg-white p-8 shadow-xl">
            <div className="grid grid-cols-2 gap-8 text-center lg:grid-cols-4">
              <div>
                <div className="mb-2 text-4xl font-bold text-[#FF7B54]">2.5M+</div>
                <div className="text-[#636E72]">Words Learned</div>
              </div>
              <div>
                <div className="mb-2 text-4xl font-bold text-[#4CAF50]">85%</div>
                <div className="text-[#636E72]">Retention Rate</div>
              </div>
              {statsConfig.showStats.languageCount && (
                <div>
                  <div className="mb-2 text-4xl font-bold text-[#9C27B0]">
                    {statsConfig.values.languageCount}
                  </div>
                  <div className="text-[#636E72]">AI Languages</div>
                </div>
              )}
              {statsConfig.showStats.starRating && (
                <div>
                  <div className="mb-2 text-4xl font-bold text-[#FFD700]">
                    {statsConfig.values.starRating}
                  </div>
                  <div className="text-[#636E72]">App Rating</div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </section>
  )
}

export default TestimonialsSection
