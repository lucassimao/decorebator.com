import React from 'react';

const NewFeaturesSection: React.FC = () => {
  const features = [
    {
      title: "7 Native AI Languages",
      description: "Native AI processing in English, Spanish, French, German, Italian, Portuguese, and Japanese with language-specific grammar rules.",
      icon: "fas fa-globe",
      bg: "bg-gradient-to-br from-indigo-50 to-purple-50",
      border: "border-indigo-100",
      iconBg: "bg-gradient-to-br from-[#6366F1] to-indigo-600",
      linkColor: "text-[#6366F1]"
    },
    {
      title: "AI-Powered Content Generation",
      description: "Automatically generates definitions, examples, images, and audio pronunciations for any word using advanced AI models.",
      icon: "fas fa-brain",
      bg: "bg-gradient-to-br from-orange-50 to-amber-50",
      border: "border-orange-100",
      iconBg: "bg-gradient-to-br from-[#FF7B54] to-orange-600",
      linkColor: "text-[#FF7B54]"
    },
    {
      title: "Leitner Spaced Repetition",
      description: "Scientifically proven system with 7 review boxes that optimizes when you review words for maximum long-term retention.",
      icon: "fas fa-clock",
      bg: "bg-gradient-to-br from-green-50 to-emerald-50",
      border: "border-green-100",
      iconBg: "bg-gradient-to-br from-[#4CAF50] to-green-600",
      linkColor: "text-[#4CAF50]"
    },
    {
      title: "8 Engaging Quiz Modes",
      description: "Multiple quiz types: meaning matching, image association, audio comprehension, sentence completion, write from definition, and more.",
      icon: "fas fa-gamepad",
      bg: "bg-gradient-to-br from-purple-50 to-pink-50",
      border: "border-purple-100",
      iconBg: "bg-gradient-to-br from-[#9C27B0] to-purple-600",
      linkColor: "text-[#9C27B0]"
    },
    {
      title: "Visual Learning with AI Images",
      description: "AI-generated images help you associate words with visual concepts for better memory retention.",
      icon: "fas fa-image",
      bg: "bg-gradient-to-br from-blue-50 to-cyan-50",
      border: "border-blue-100",
      iconBg: "bg-gradient-to-br from-[#2196F3] to-blue-600",
      linkColor: "text-[#2196F3]"
    },
    {
      title: "Multi-Language Audio",
      description: "Language-optimized voice selection with advanced text-to-speech. Each language uses native-sounding voices for authentic pronunciation.",
      icon: "fas fa-headphones",
      bg: "bg-gradient-to-br from-yellow-50 to-amber-50",
      border: "border-yellow-100",
      iconBg: "bg-gradient-to-br from-[#FFD700] to-yellow-600",
      linkColor: "text-yellow-600"
    },
    {
      title: "Interactive Flashcards",
      description: "Immersive full-screen flashcards with flip animations, rich content display, and integrated error reporting.",
      icon: "fas fa-layer-group",
      bg: "bg-gradient-to-br from-red-50 to-pink-50",
      border: "border-red-100",
      iconBg: "bg-gradient-to-br from-[#FF6B6B] to-red-600",
      linkColor: "text-[#FF6B6B]"
    },
    {
      title: "Advanced Analytics",
      description: "Comprehensive progress tracking with word mastery levels, learning velocity, retention rates, and performance insights.",
      icon: "fas fa-chart-line",
      bg: "bg-gradient-to-br from-teal-50 to-cyan-50",
      border: "border-teal-100",
      iconBg: "bg-gradient-to-br from-[#14B8A6] to-teal-600",
      linkColor: "text-[#14B8A6]"
    },
    {
      title: "Error Reporting System",
      description: "User-driven quality control with 5 error types for AI content. Report issues and get automatic content regeneration.",
      icon: "fas fa-exclamation-triangle",
      bg: "bg-gradient-to-br from-gray-50 to-slate-50",
      border: "border-gray-100",
      iconBg: "bg-gradient-to-br from-[#64748B] to-slate-600",
      linkColor: "text-[#64748B]"
    },
    {
      title: "Offline Support",
      description: "Premium users can download wordlists for offline access. Complete quiz functionality without internet connection.",
      icon: "fas fa-wifi-slash",
      bg: "bg-gradient-to-br from-emerald-50 to-green-50",
      border: "border-emerald-100",
      iconBg: "bg-gradient-to-br from-[#10B981] to-emerald-600",
      linkColor: "text-[#10B981]"
    }
  ];

  return (
    <section id="features" className="py-20 bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            Everything You Need to 
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent"> Master Vocabulary</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
            Our comprehensive platform combines cutting-edge AI technology with proven learning science to accelerate your language journey.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature) => (
            <div 
              key={feature.title}
              className={`p-8 rounded-2xl ${feature.bg} border ${feature.border} hover:shadow-2xl transition-all duration-500 transform hover:-translate-y-2`}
            >
              <div className={`w-16 h-16 ${feature.iconBg} rounded-2xl flex items-center justify-center mb-6 transition-transform duration-300`}>
                <i className={`${feature.icon} text-white text-2xl`}></i>
              </div>
              <h3 className="text-xl font-bold mb-3">{feature.title}</h3>
              <p className="text-[#636E72] leading-relaxed">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default NewFeaturesSection;