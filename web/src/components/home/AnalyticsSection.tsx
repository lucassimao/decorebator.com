'use client'

import React, { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
import { motion } from 'framer-motion'
import { TrendingUp, Flame, Trophy, Target, BarChart3, BookOpen, Zap, Calendar } from 'lucide-react'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Filler,
} from 'chart.js'
import { Line, Bar } from 'react-chartjs-2'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Filler
)

const AnalyticsSection: React.FC = () => {
  const t = useTranslations('analytics')
  const [isVisible, setIsVisible] = useState(false)
  const [animatedStats, setAnimatedStats] = useState({
    wordsStudied: 0,
    streak: 0,
    mastered: 0,
    accuracy: 0,
  })

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true)
          // Animate numbers
          const duration = 1500
          const start = Date.now()
          const targets = { wordsStudied: 247, streak: 42, mastered: 1923, accuracy: 89 }

          const animate = () => {
            const elapsed = Date.now() - start
            const progress = Math.min(elapsed / duration, 1)
            const eased = 1 - Math.pow(1 - progress, 3)

            setAnimatedStats({
              wordsStudied: Math.round(targets.wordsStudied * eased),
              streak: Math.round(targets.streak * eased),
              mastered: Math.round(targets.mastered * eased),
              accuracy: Math.round(targets.accuracy * eased),
            })

            if (progress < 1) requestAnimationFrame(animate)
          }
          requestAnimationFrame(animate)
        }
      },
      { threshold: 0.2 }
    )

    const section = document.getElementById('analytics')
    if (section) observer.observe(section)
    return () => observer.disconnect()
  }, [])

  // Chart configurations
  const weeklyProgressData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Words Studied',
        data: [15, 22, 18, 28, 24, 35, 42],
        borderColor: '#6366f1',
        backgroundColor: 'rgba(99, 102, 241, 0.1)',
        pointBackgroundColor: '#6366f1',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6,
        tension: 0.4,
        fill: true,
      },
    ],
  }

  const practiceTimeData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Minutes',
        data: [25, 18, 32, 28, 22, 45, 38],
        backgroundColor: (context: { dataIndex: number }) => {
          const index = context.dataIndex
          return index === 6 ? '#6366f1' : '#e0e7ff'
        },
        borderRadius: 8,
        barThickness: 24,
      },
    ],
  }

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: '#1e293b',
        titleColor: '#fff',
        bodyColor: '#fff',
        cornerRadius: 8,
        padding: 12,
      },
    },
    scales: {
      x: {
        grid: { display: false },
        ticks: { color: '#94a3b8', font: { size: 12 } },
      },
      y: {
        grid: { color: '#f1f5f9' },
        ticks: { color: '#94a3b8', font: { size: 12 } },
        beginAtZero: true,
      },
    },
  }

  // Word mastery data
  const wordMasteryItems = [
    { word: 'Bonjour', mastery: 92, color: 'bg-emerald-500' },
    { word: 'Merci', mastery: 88, color: 'bg-indigo-500' },
    { word: 'Au revoir', mastery: 85, color: 'bg-amber-500' },
    { word: "S'il vous plaît", mastery: 79, color: 'bg-rose-500' },
  ]

  // Box distribution
  const boxDistribution = [
    { box: 1, count: 12, label: 'New', color: 'bg-slate-400' },
    { box: 2, count: 24, label: '6h', color: 'bg-rose-400' },
    { box: 3, count: 45, label: '1d', color: 'bg-amber-400' },
    { box: 4, count: 78, label: '3d', color: 'bg-emerald-400' },
    { box: 5, count: 156, label: '7d', color: 'bg-sky-400' },
    { box: 6, count: 234, label: '14d', color: 'bg-indigo-400' },
    { box: 7, count: 1374, label: '30d', color: 'bg-violet-500' },
  ]
  const maxBoxCount = Math.max(...boxDistribution.map((b) => b.count))

  return (
    <section
      id="analytics"
      className="relative overflow-hidden bg-gradient-to-b from-slate-50 via-white to-slate-50 py-16 sm:py-20 lg:py-24"
    >
      {/* Background decorations */}
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute -top-40 right-0 h-96 w-96 rounded-full bg-indigo-100/60 blur-3xl" />
        <div className="absolute -bottom-40 left-0 h-96 w-96 rounded-full bg-violet-100/60 blur-3xl" />
      </div>

      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-12 text-center lg:mb-16">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="mb-4 inline-flex items-center gap-2 rounded-full border border-indigo-200 bg-indigo-50 px-4 py-2"
          >
            <BarChart3 className="h-4 w-4 text-indigo-600" />
            <span className="text-sm font-semibold text-indigo-700">Premium Feature</span>
          </motion.div>
          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.1 }}
            className="mb-4 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl"
          >
            {t('title')}
          </motion.h2>
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: 0.2 }}
            className="mx-auto max-w-2xl text-lg text-slate-600"
          >
            {t('subtitle')}
          </motion.p>
        </div>

        {/* Modern Bento Grid Layout */}
        <div className="grid grid-cols-1 gap-4 sm:gap-5 lg:grid-cols-12 lg:grid-rows-[auto_auto_auto]">
          {/* Stats Cards Row - Spanning full width */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={isVisible ? { opacity: 1, y: 0 } : {}}
            transition={{ delay: 0.1 }}
            className="col-span-full grid grid-cols-2 gap-4 sm:grid-cols-4 sm:gap-5"
          >
            {/* Words This Week */}
            <div className="group relative overflow-hidden rounded-2xl border border-indigo-100 bg-gradient-to-br from-indigo-50 to-white p-5 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-indigo-100">
              <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-100 text-indigo-600 transition-transform duration-300 group-hover:scale-110">
                <BookOpen className="h-5 w-5" />
              </div>
              <div className="text-3xl font-bold text-slate-900">{animatedStats.wordsStudied}</div>
              <div className="text-sm font-medium text-slate-500">
                {t('stats.wordsStudiedToday')}
              </div>
              <div className="mt-2 flex items-center gap-1 text-xs font-semibold text-emerald-600">
                <TrendingUp className="h-3 w-3" />
                +18% vs last week
              </div>
            </div>

            {/* Streak */}
            <div className="group relative overflow-hidden rounded-2xl border border-amber-100 bg-gradient-to-br from-amber-50 to-white p-5 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-amber-100">
              <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-amber-100 text-amber-600 transition-transform duration-300 group-hover:scale-110">
                <Flame className="h-5 w-5" />
              </div>
              <div className="text-3xl font-bold text-slate-900">{animatedStats.streak}</div>
              <div className="text-sm font-medium text-slate-500">{t('stats.currentStreak')}</div>
              <div className="mt-2 text-xs font-semibold text-amber-600">🔥 Personal best!</div>
            </div>

            {/* Words Mastered */}
            <div className="group relative overflow-hidden rounded-2xl border border-emerald-100 bg-gradient-to-br from-emerald-50 to-white p-5 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-emerald-100">
              <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-100 text-emerald-600 transition-transform duration-300 group-hover:scale-110">
                <Trophy className="h-5 w-5" />
              </div>
              <div className="text-3xl font-bold text-slate-900">
                {animatedStats.mastered.toLocaleString()}
              </div>
              <div className="text-sm font-medium text-slate-500">{t('stats.wordsMastered')}</div>
              <div className="mt-2 flex items-center gap-1 text-xs font-semibold text-emerald-600">
                <Zap className="h-3 w-3" />
                Top 5% learner
              </div>
            </div>

            {/* Accuracy */}
            <div className="group relative overflow-hidden rounded-2xl border border-violet-100 bg-gradient-to-br from-violet-50 to-white p-5 transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-violet-100">
              <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-violet-100 text-violet-600 transition-transform duration-300 group-hover:scale-110">
                <Target className="h-5 w-5" />
              </div>
              <div className="text-3xl font-bold text-slate-900">{animatedStats.accuracy}%</div>
              <div className="text-sm font-medium text-slate-500">{t('stats.accuracyToday')}</div>
              <div className="mt-2 text-xs font-semibold text-violet-600">Above average</div>
            </div>
          </motion.div>

          {/* Weekly Progress Chart - Large Card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={isVisible ? { opacity: 1, y: 0 } : {}}
            transition={{ delay: 0.2 }}
            className="card-shine overflow-hidden rounded-2xl border border-slate-200 bg-white p-6 shadow-sm transition-all duration-300 hover:shadow-lg lg:col-span-8"
          >
            <div className="mb-6 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h3 className="text-lg font-bold text-slate-900">
                  {t('charts.learningProgressTitle')}
                </h3>
                <p className="text-sm text-slate-500">Words studied per day</p>
              </div>
              <div className="flex items-center gap-2 rounded-lg bg-slate-100 px-3 py-1.5">
                <Calendar className="h-4 w-4 text-slate-500" />
                <span className="text-sm font-medium text-slate-600">This Week</span>
              </div>
            </div>
            <div className="h-64">
              {isVisible && <Line data={weeklyProgressData} options={chartOptions} />}
            </div>
          </motion.div>

          {/* Word Mastery Progress - Side Card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={isVisible ? { opacity: 1, y: 0 } : {}}
            transition={{ delay: 0.3 }}
            className="overflow-hidden rounded-2xl border border-slate-200 bg-white p-6 shadow-sm transition-all duration-300 hover:shadow-lg lg:col-span-4"
          >
            <h3 className="mb-5 text-lg font-bold text-slate-900">
              {t('charts.wordMasteryTitle')}
            </h3>
            <div className="space-y-4">
              {wordMasteryItems.map((item, index) => (
                <div
                  key={item.word}
                  className="group"
                  style={{
                    opacity: isVisible ? 1 : 0,
                    transform: isVisible ? 'translateX(0)' : 'translateX(-10px)',
                    transition: `all 0.4s ease ${index * 0.1}s`,
                  }}
                >
                  <div className="mb-1.5 flex items-center justify-between">
                    <span className="font-medium text-slate-700 transition-colors group-hover:text-slate-900">
                      {item.word}
                    </span>
                    <span className="text-sm font-semibold text-slate-900">{item.mastery}%</span>
                  </div>
                  <div className="h-2 overflow-hidden rounded-full bg-slate-100">
                    <div
                      className={`h-full rounded-full ${item.color} transition-all duration-700`}
                      style={{
                        width: isVisible ? `${item.mastery}%` : '0%',
                        transitionDelay: `${index * 0.1 + 0.3}s`,
                      }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </motion.div>

          {/* Practice Time - Medium Card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={isVisible ? { opacity: 1, y: 0 } : {}}
            transition={{ delay: 0.4 }}
            className="overflow-hidden rounded-2xl border border-slate-200 bg-white p-6 shadow-sm transition-all duration-300 hover:shadow-lg lg:col-span-5"
          >
            <div className="mb-5 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <h3 className="text-lg font-bold text-slate-900">{t('charts.practiceTimeTitle')}</h3>
              <div className="flex gap-4 text-sm">
                <div>
                  <span className="text-slate-500">Total: </span>
                  <span className="font-semibold text-slate-900">3h 28m</span>
                </div>
              </div>
            </div>
            <div className="h-48">
              {isVisible && <Bar data={practiceTimeData} options={chartOptions} />}
            </div>
          </motion.div>

          {/* Box Distribution - Visual Card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={isVisible ? { opacity: 1, y: 0 } : {}}
            transition={{ delay: 0.5 }}
            className="overflow-hidden rounded-2xl border border-slate-200 bg-gradient-to-br from-slate-900 to-slate-800 p-6 text-white shadow-sm transition-all duration-300 hover:shadow-lg lg:col-span-7"
          >
            <div className="mb-5 flex items-center justify-between">
              <div>
                <h3 className="text-lg font-bold">Spaced Repetition Boxes</h3>
                <p className="text-sm text-slate-400">Your vocabulary distribution</p>
              </div>
              <div className="rounded-lg bg-white/10 px-3 py-1.5 text-sm font-medium backdrop-blur">
                1,923 total words
              </div>
            </div>

            <div className="flex items-end justify-between gap-2 pt-4">
              {boxDistribution.map((item, index) => (
                <div
                  key={item.box}
                  className="flex flex-1 flex-col items-center gap-2"
                  style={{
                    opacity: isVisible ? 1 : 0,
                    transform: isVisible ? 'translateY(0)' : 'translateY(20px)',
                    transition: `all 0.5s ease ${index * 0.08}s`,
                  }}
                >
                  <div className="text-xs font-semibold text-slate-300">{item.count}</div>
                  <div
                    className={`w-full rounded-t-lg ${item.color} transition-all duration-700`}
                    style={{
                      height: isVisible ? `${(item.count / maxBoxCount) * 100 + 20}px` : '4px',
                      transitionDelay: `${index * 0.08 + 0.2}s`,
                    }}
                  />
                  <div className="text-center">
                    <div className="text-xs font-medium text-white">Box {item.box}</div>
                    <div className="text-[10px] text-slate-400">{item.label}</div>
                  </div>
                </div>
              ))}
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  )
}

export default AnalyticsSection
