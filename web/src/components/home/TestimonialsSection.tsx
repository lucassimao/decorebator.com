import React from 'react';

const TestimonialsSection: React.FC = () => {
  return (
    <section id="testimonials" className="py-20 bg-gradient-to-br from-orange-50 to-amber-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            Loved by 
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent"> Language Learners</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
            Join thousands of satisfied learners who have transformed their language skills with Decorebator.
          </p>
        </div>

        <div className="grid lg:grid-cols-3 gap-8">
          {/* Testimonial 1 */}
          <div className="bg-white p-8 rounded-3xl shadow-xl hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-2">
            <div className="flex items-center mb-4">
              <img 
                src="https://ui-avatars.com/api/?name=Sarah+Chen&background=FF7B54&color=fff&size=48&rounded=true" 
                alt="Sarah Chen" 
                className="w-12 h-12 rounded-full mr-4"
              />
              <div>
                <div className="font-bold">Sarah Chen</div>
                <div className="text-sm text-[#636E72]">Language Student</div>
              </div>
            </div>
            <div className="flex mb-4">
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
            </div>
            <p className="text-[#636E72] leading-relaxed italic">
              &quot;Decorebator transformed my Spanish learning. The AI-generated images help me remember words so much better than traditional methods. I&apos;ve learned more in 3 months than in years of traditional study!&quot;
            </p>
          </div>

          {/* Testimonial 2 */}
          <div className="bg-white p-8 rounded-3xl shadow-xl hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-2">
            <div className="flex items-center mb-4">
              <img 
                src="https://ui-avatars.com/api/?name=Marcus+Rodriguez&background=4CAF50&color=fff&size=48&rounded=true" 
                alt="Marcus Rodriguez" 
                className="w-12 h-12 rounded-full mr-4"
              />
              <div>
                <div className="font-bold">Marcus Rodriguez</div>
                <div className="text-sm text-[#636E72]">Professional Translator</div>
              </div>
            </div>
            <div className="flex mb-4">
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
            </div>
            <p className="text-[#636E72] leading-relaxed italic">
              &quot;The spaced repetition system is incredibly effective. I&apos;ve expanded my technical vocabulary by 300% in just 3 months. The 8 quiz modes keep learning fresh and engaging every day.&quot;
            </p>
          </div>

          {/* Testimonial 3 */}
          <div className="bg-white p-8 rounded-3xl shadow-xl hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-2">
            <div className="flex items-center mb-4">
              <img 
                src="https://ui-avatars.com/api/?name=Emma+Thompson&background=9C27B0&color=fff&size=48&rounded=true" 
                alt="Emma Thompson" 
                className="w-12 h-12 rounded-full mr-4"
              />
              <div>
                <div className="font-bold">Emma Thompson</div>
                <div className="text-sm text-[#636E72]">ESL Teacher</div>
              </div>
            </div>
            <div className="flex mb-4">
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
              <i className="fas fa-star text-yellow-400"></i>
            </div>
            <p className="text-[#636E72] leading-relaxed italic">
              &quot;I recommend Decorebator to all my students. The variety of quiz modes keeps them engaged and the progress tracking is excellent. It&apos;s the perfect blend of technology and pedagogy.&quot;
            </p>
          </div>
        </div>

        {/* Success Metrics */}
        <div className="mt-16 bg-white rounded-3xl p-8 shadow-xl">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-8 text-center">
            <div>
              <div className="text-4xl font-bold text-[#FF7B54] mb-2">2.5M+</div>
              <div className="text-[#636E72]">Words Learned</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-[#4CAF50] mb-2">85%</div>
              <div className="text-[#636E72]">Retention Rate</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-[#9C27B0] mb-2">7</div>
              <div className="text-[#636E72]">AI Languages</div>
            </div>
            <div>
              <div className="text-4xl font-bold text-[#FFD700] mb-2">4.9/5</div>
              <div className="text-[#636E72]">App Rating</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default TestimonialsSection;