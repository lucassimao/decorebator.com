'use client'

import React, { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { motion } from 'framer-motion'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  RadialLinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line, Bar, Doughnut } from 'react-chartjs-2'

ChartJS.register(
  CategoryScale,
  LinearScale,
  RadialLinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const AnalyticsSection: React.FC = () => {
  const t = useTranslations('analytics')
  const [animatedValues, setAnimatedValues] = useState({
    wordsStudiedToday: 0,
    currentStreak: 0,
    wordsMastered: 0,
    accuracyToday: 0,
  })
  const [chartsVisible, setChartsVisible] = useState({
    stats: false,
    learningProgress: false,
    practiceTime: false,
    quizPerformance: false,
    wordMastery: false,
    boxDistribution: false,
    historicalBox: false,
    topWords: false,
  })

  // Animate numbers on mount
  useEffect(() => {
    const timer = setTimeout(() => {
      setAnimatedValues({
        wordsStudiedToday: 12,
        currentStreak: 42,
        wordsMastered: 1923,
        accuracyToday: 89,
      })
    }, 500)

    return () => clearTimeout(timer)
  }, [])

  // Intersection Observer for scroll-triggered animations
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const chartId = entry.target.getAttribute('data-chart')
            if (chartId) {
              setTimeout(() => {
                setChartsVisible((prev) => ({ ...prev, [chartId]: true }))
              }, 100)
            }
          }
        })
      },
      { threshold: 0.1 }
    )

    const chartElements = document.querySelectorAll('[data-chart]')
    chartElements.forEach((el) => observer.observe(el))

    return () => observer.disconnect()
  }, [])

  // Stats Grid with theme colors
  const statsGrid = [
    {
      label: t('stats.wordsStudiedToday'),
      value: animatedValues.wordsStudiedToday,
      icon: '📚',
      gradient: 'from-primary-500 to-primary-600',
      bgColor: 'bg-primary-50',
      textColor: 'text-primary-600',
      borderColor: 'border-primary-200',
    },
    {
      label: t('stats.currentStreak'),
      value: animatedValues.currentStreak,
      suffix: t('stats.daysSuffix'),
      icon: '🔥',
      gradient: 'from-orange-500 to-red-500',
      bgColor: 'bg-orange-50',
      textColor: 'text-orange-600',
      borderColor: 'border-orange-200',
      highlight: true,
    },
    {
      label: t('stats.wordsMastered'),
      value: animatedValues.wordsMastered,
      icon: '🏆',
      gradient: 'from-success-500 to-success-600',
      bgColor: 'bg-success-50',
      textColor: 'text-success-600',
      borderColor: 'border-success-200',
    },
    {
      label: t('stats.accuracyToday'),
      value: animatedValues.accuracyToday,
      suffix: '%',
      icon: '🎯',
      gradient: 'from-accent-500 to-accent-600',
      bgColor: 'bg-accent-50',
      textColor: 'text-accent-600',
      borderColor: 'border-accent-200',
    },
  ]

  // Chart.js base configuration
  const baseChartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    animation: {
      duration: 1000,
      easing: 'easeInOutQuart' as const,
    },
    plugins: {
      legend: {
        display: false,
      },
      tooltip: {
        backgroundColor: 'rgba(15, 23, 42, 0.9)',
        titleColor: '#FFFFFF',
        bodyColor: '#FFFFFF',
        borderColor: '#FF6B35',
        borderWidth: 1,
        cornerRadius: 8,
        padding: 12,
      },
    },
  }

  // Theme colors for charts
  const chartColors = {
    primary: '#FF6B35',
    secondary: '#004E89',
    accent: '#7B2CBF',
    success: '#06D6A0',
  }

  // 1. Learning Progress Chart
  const learningProgressData = {
    labels: ['12/20', '12/21', '12/22', '12/23', '12/24', '12/25', '12/26'],
    datasets: [
      {
        label: t('charts.wordsStudied'),
        data: [15, 12, 18, 22, 16, 25, 20],
        borderColor: chartColors.primary,
        backgroundColor: 'rgba(255, 107, 53, 0.1)',
        pointBackgroundColor: chartColors.primary,
        pointBorderColor: '#FFFFFF',
        pointBorderWidth: 2,
        pointRadius: 5,
        pointHoverRadius: 7,
        tension: 0.4,
        fill: true,
      },
    ],
  }

  // 2. Practice Time Chart
  const practiceTimeData = {
    labels: t.raw('charts.weekDays') as string[],
    datasets: [
      {
        label: t('charts.minutes'),
        data: [25, 18, 32, 28, 22, 45, 38],
        backgroundColor: chartColors.secondary,
        borderRadius: 6,
        hoverBackgroundColor: '#003A6B',
      },
    ],
  }

  // 3. Quiz Performance Chart
  const quizPerformanceData = {
    labels: t.raw('charts.quizTypes') as string[],
    datasets: [
      {
        label: t('charts.successRate'),
        data: [94, 88, 87, 85],
        backgroundColor: chartColors.success,
        borderRadius: 6,
        hoverBackgroundColor: '#05AB80',
      },
    ],
  }

  // 4. Word Mastery Progress
  const topWordsMastery = [
    { word: 'Bonjour', mastery: 92, color: chartColors.primary },
    { word: 'Merci', mastery: 88, color: chartColors.success },
    { word: 'Au revoir', mastery: 85, color: '#EAB308' },
    { word: "S'il vous plaît", mastery: 79, color: chartColors.accent },
    { word: 'Excusez-moi', mastery: 74, color: chartColors.secondary },
    { word: 'Bonne journée', mastery: 68, color: '#F97316' },
  ]

  // 5. Top Words List
  const topWordsList = [
    { rank: 1, word: 'Bonjour', mastery: 92, box: 7 },
    { rank: 2, word: 'Merci', mastery: 88, box: 7 },
    { rank: 3, word: 'Au revoir', mastery: 85, box: 6 },
    { rank: 4, word: "S'il vous plaît", mastery: 79, box: 6 },
    { rank: 5, word: 'Excusez-moi', mastery: 74, box: 5 },
  ]

  return (
    <section
      id="analytics"
      className="relative overflow-hidden bg-gradient-to-b from-white via-slate-50 to-white py-16 sm:py-20 lg:py-24"
    >
      <div className="pointer-events-none absolute inset-0">
        <div className="bg-primary-100/60 absolute -top-32 right-[-10%] h-[22rem] w-[22rem] rounded-full blur-[120px]" />
        <div className="bg-accent-100/50 absolute -bottom-40 left-[-5%] h-[26rem] w-[26rem] rounded-full blur-[120px]" />
      </div>
      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-12 text-center lg:mb-16">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="border-accent-200 mb-4 inline-flex items-center gap-2 rounded-full border bg-white/70 px-4 py-2 shadow-sm backdrop-blur"
          >
            <span className="text-accent-700 text-sm font-semibold">Premium Feature</span>
          </motion.div>
          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="mb-4 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl"
          >
            {t('title')}
          </motion.h2>
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="mx-auto max-w-2xl text-lg text-slate-600"
          >
            {t('subtitle')}
          </motion.p>
        </div>

        {/* Stats Grid */}
        <div
          data-chart="stats"
          className={`mb-6 transition-all duration-700 sm:mb-8 ${
            chartsVisible.stats ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
          }`}
        >
          <div className="grid grid-cols-2 gap-4 sm:gap-5 lg:grid-cols-4">
            {statsGrid.map((stat, index) => (
              <div
                key={stat.label}
                className={`group relative overflow-hidden rounded-2xl border ${stat.borderColor} ${stat.bgColor} p-4 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg sm:p-5 ${
                  stat.highlight ? 'ring-2 ring-orange-300/70' : ''
                }`}
                style={{ transitionDelay: `${index * 100}ms` }}
              >
                <div className="mb-3 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-white/80 text-2xl shadow-sm">
                  {stat.icon}
                </div>
                <div className={`mb-1 text-2xl font-bold ${stat.textColor} sm:text-3xl`}>
                  {stat.value.toLocaleString()}
                  {stat.suffix || ''}
                </div>
                <div className="text-sm font-medium text-slate-600">{stat.label}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Charts Grid */}
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
          <div
            data-chart="learningProgress"
            className={`rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition-all duration-700 hover:-translate-y-0.5 hover:shadow-lg sm:p-6 lg:col-span-8 ${
              chartsVisible.learningProgress
                ? 'translate-y-0 opacity-100'
                : 'translate-y-4 opacity-0'
            }`}
          >
            <h3 className="mb-4 text-lg font-bold text-slate-900">
              {t('charts.learningProgressTitle')}
            </h3>
            <div className="h-60 sm:h-64">
              {chartsVisible.learningProgress && (
                <Line
                  data={learningProgressData}
                  options={{
                    ...baseChartOptions,
                    scales: {
                      x: {
                        grid: { display: false },
                        ticks: { color: '#64748B' },
                      },
                      y: {
                        grid: { color: '#F1F5F9' },
                        ticks: { color: '#64748B' },
                        beginAtZero: true,
                      },
                    },
                  }}
                />
              )}
            </div>
            <p className="mt-4 text-center text-sm text-slate-500">
              {t('charts.wordsStudiedDescription')}
            </p>
          </div>

          <div className="flex flex-col gap-6 lg:col-span-4">
            <div
              data-chart="practiceTime"
              className={`rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition-all duration-700 hover:-translate-y-0.5 hover:shadow-lg sm:p-6 ${
                chartsVisible.practiceTime ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
              }`}
            >
              <div className="mb-4 flex flex-col gap-2">
                <h3 className="text-lg font-bold text-slate-900">
                  {t('charts.practiceTimeTitle')}
                </h3>
                <div className="flex flex-wrap gap-4 text-sm">
                  <div>
                    <span className="text-slate-500">{t('charts.total')}: </span>
                    <span className="font-semibold text-slate-900">3h 38m</span>
                  </div>
                  <div>
                    <span className="text-slate-500">{t('charts.dailyAvg')}: </span>
                    <span className="font-semibold text-slate-900">31m</span>
                  </div>
                </div>
              </div>
              <div className="h-44 sm:h-48">
                {chartsVisible.practiceTime && (
                  <Bar
                    data={practiceTimeData}
                    options={{
                      ...baseChartOptions,
                      scales: {
                        x: {
                          grid: { display: false },
                          ticks: { color: '#64748B' },
                        },
                        y: {
                          grid: { color: '#F1F5F9' },
                          ticks: {
                            color: '#64748B',
                            callback: (value) => `${value}m`,
                          },
                          beginAtZero: true,
                        },
                      },
                    }}
                  />
                )}
              </div>
            </div>

            <div
              data-chart="quizPerformance"
              className={`rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition-all duration-700 hover:-translate-y-0.5 hover:shadow-lg sm:p-6 ${
                chartsVisible.quizPerformance
                  ? 'translate-y-0 opacity-100'
                  : 'translate-y-4 opacity-0'
              }`}
            >
              <h3 className="mb-4 text-lg font-bold text-slate-900">
                {t('charts.quizPerformanceTitle')}
              </h3>
              <div className="h-40 sm:h-44">
                {chartsVisible.quizPerformance && (
                  <Bar
                    data={quizPerformanceData}
                    options={{
                      ...baseChartOptions,
                      indexAxis: 'y',
                      scales: {
                        x: {
                          grid: { color: '#F1F5F9' },
                          ticks: {
                            color: '#64748B',
                            callback: (value) => `${value}%`,
                          },
                          beginAtZero: true,
                          max: 100,
                        },
                        y: {
                          grid: { display: false },
                          ticks: { color: '#64748B' },
                        },
                      },
                    }}
                  />
                )}
              </div>
            </div>
          </div>

          <div
            data-chart="wordMastery"
            className={`rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition-all duration-700 hover:-translate-y-0.5 hover:shadow-lg sm:p-6 lg:col-span-7 ${
              chartsVisible.wordMastery ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
            }`}
          >
            <h3 className="mb-6 text-lg font-bold text-slate-900">
              {t('charts.wordMasteryTitle')}
            </h3>
            <div className="grid grid-cols-2 gap-6 sm:grid-cols-3 lg:grid-cols-6">
              {topWordsMastery.map((word, index) => (
                <div
                  key={word.word}
                  className="text-center transition-all duration-500"
                  style={{
                    transitionDelay: chartsVisible.wordMastery ? `${index * 100}ms` : '0ms',
                    opacity: chartsVisible.wordMastery ? 1 : 0,
                    transform: chartsVisible.wordMastery ? 'scale(1)' : 'scale(0.8)',
                  }}
                >
                  <div className="relative mx-auto mb-3 h-20 w-20">
                    <Doughnut
                      data={{
                        datasets: [
                          {
                            data: [word.mastery, 100 - word.mastery],
                            backgroundColor: [word.color, '#F1F5F9'],
                            borderWidth: 0,
                          },
                        ],
                      }}
                      options={{
                        ...baseChartOptions,
                        cutout: '70%',
                        plugins: {
                          tooltip: { enabled: false },
                        },
                      }}
                    />
                    <div className="absolute inset-0 flex items-center justify-center">
                      <span className="text-sm font-bold text-slate-900">{word.mastery}%</span>
                    </div>
                  </div>
                  <p className="text-sm font-medium text-slate-700">{word.word}</p>
                </div>
              ))}
            </div>
          </div>

          <div
            data-chart="topWords"
            className={`rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition-all duration-700 hover:-translate-y-0.5 hover:shadow-lg sm:p-6 lg:col-span-5 ${
              chartsVisible.topWords ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
            }`}
          >
            <h3 className="mb-6 text-lg font-bold text-slate-900">{t('charts.topWordsTitle')}</h3>
            <div className="space-y-3">
              {topWordsList.map((word, index) => (
                <div
                  key={word.rank}
                  className="flex items-center justify-between rounded-xl border border-slate-200/70 bg-white/90 p-4 transition-all duration-300 hover:bg-slate-50"
                  style={{
                    transitionDelay: chartsVisible.topWords ? `${index * 100}ms` : '0ms',
                    opacity: chartsVisible.topWords ? 1 : 0,
                    transform: chartsVisible.topWords ? 'translateX(0)' : 'translateX(-20px)',
                  }}
                >
                  <div className="flex items-center gap-4">
                    <div
                      className={`flex h-10 w-10 items-center justify-center rounded-full text-sm font-bold text-white ${
                        word.rank === 1
                          ? 'bg-yellow-500'
                          : word.rank === 2
                            ? 'bg-slate-400'
                            : word.rank === 3
                              ? 'bg-amber-600'
                              : 'bg-slate-300'
                      }`}
                    >
                      {word.rank}
                    </div>
                    <span className="font-semibold text-slate-900">{word.word}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="bg-success-100 text-success-700 rounded-full px-3 py-1 text-sm font-semibold">
                      {word.mastery}%
                    </span>
                    <span className="rounded-full bg-slate-200 px-3 py-1 text-sm font-medium text-slate-600">
                      Box {word.box}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

export default AnalyticsSection
