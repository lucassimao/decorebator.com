import React from 'react';
import { getTranslations } from 'next-intl/server';

const NewFeaturesSection: React.FC = async () => {
  const t = await getTranslations('newFeatures');

  const features = [
    {
      key: 'nativeLanguages',
      title: t('items.nativeLanguages.title'),
      description: t('items.nativeLanguages.description'),
      icon: "fas fa-globe",
      bg: "bg-gradient-to-br from-indigo-50 to-purple-50",
      border: "border-indigo-100",
      iconBg: "bg-gradient-to-br from-[#6366F1] to-indigo-600",
      linkColor: "text-[#6366F1]"
    },
    {
      key: 'aiContent',
      title: t('items.aiContent.title'),
      description: t('items.aiContent.description'),
      icon: "fas fa-brain",
      bg: "bg-gradient-to-br from-orange-50 to-amber-50",
      border: "border-orange-100",
      iconBg: "bg-gradient-to-br from-[#FF7B54] to-orange-600",
      linkColor: "text-[#FF7B54]"
    },
    {
      key: 'spacedRepetition',
      title: t('items.spacedRepetition.title'),
      description: t('items.spacedRepetition.description'),
      icon: "fas fa-clock",
      bg: "bg-gradient-to-br from-green-50 to-emerald-50",
      border: "border-green-100",
      iconBg: "bg-gradient-to-br from-[#4CAF50] to-green-600",
      linkColor: "text-[#4CAF50]"
    },
    {
      key: 'quizModes',
      title: t('items.quizModes.title'),
      description: t('items.quizModes.description'),
      icon: "fas fa-gamepad",
      bg: "bg-gradient-to-br from-purple-50 to-pink-50",
      border: "border-purple-100",
      iconBg: "bg-gradient-to-br from-[#9C27B0] to-purple-600",
      linkColor: "text-[#9C27B0]"
    },
    {
      key: 'visualLearning',
      title: t('items.visualLearning.title'),
      description: t('items.visualLearning.description'),
      icon: "fas fa-image",
      bg: "bg-gradient-to-br from-blue-50 to-cyan-50",
      border: "border-blue-100",
      iconBg: "bg-gradient-to-br from-[#2196F3] to-blue-600",
      linkColor: "text-[#2196F3]"
    },
    {
      key: 'multiLanguageAudio',
      title: t('items.multiLanguageAudio.title'),
      description: t('items.multiLanguageAudio.description'),
      icon: "fas fa-headphones",
      bg: "bg-gradient-to-br from-yellow-50 to-amber-50",
      border: "border-yellow-100",
      iconBg: "bg-gradient-to-br from-[#FFD700] to-yellow-600",
      linkColor: "text-yellow-600"
    },
    {
      key: 'flashcards',
      title: t('items.flashcards.title'),
      description: t('items.flashcards.description'),
      icon: "fas fa-layer-group",
      bg: "bg-gradient-to-br from-red-50 to-pink-50",
      border: "border-red-100",
      iconBg: "bg-gradient-to-br from-[#FF6B6B] to-red-600",
      linkColor: "text-[#FF6B6B]"
    },
    {
      key: 'analytics',
      title: t('items.analytics.title'),
      description: t('items.analytics.description'),
      icon: "fas fa-chart-line",
      bg: "bg-gradient-to-br from-teal-50 to-cyan-50",
      border: "border-teal-100",
      iconBg: "bg-gradient-to-br from-[#14B8A6] to-teal-600",
      linkColor: "text-[#14B8A6]"
    },
    {
      key: 'errorReporting',
      title: t('items.errorReporting.title'),
      description: t('items.errorReporting.description'),
      icon: "fas fa-exclamation-triangle",
      bg: "bg-gradient-to-br from-gray-50 to-slate-50",
      border: "border-gray-100",
      iconBg: "bg-gradient-to-br from-[#64748B] to-slate-600",
      linkColor: "text-[#64748B]"
    },
    {
      key: 'offlineSupport',
      title: t('items.offlineSupport.title'),
      description: t('items.offlineSupport.description'),
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
            <span>{t('title.part1')}</span>
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">{t('title.part2')}</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
            {t('subtitle')}
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