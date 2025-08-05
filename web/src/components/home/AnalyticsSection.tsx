'use client'

import React, { useState, useEffect } from 'react'
import { useTranslations } from 'next-intl'
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

    // Observe all chart containers
    const chartElements = document.querySelectorAll('[data-chart]')
    chartElements.forEach((el) => observer.observe(el))

    return () => observer.disconnect()
  }, [])

  // Stats Grid - matching mobile app exactly
  const statsGrid = [
    {
      label: t('stats.wordsStudiedToday'),
      value: animatedValues.wordsStudiedToday,
      icon: '📚',
      bgColor: 'bg-white',
      textColor: 'text-[#2D3436]',
    },
    {
      label: t('stats.currentStreak'),
      value: animatedValues.currentStreak,
      suffix: t('stats.daysSuffix'),
      icon: '🔥',
      bgColor: 'bg-orange-50',
      textColor: 'text-[#FF7B54]',
    },
    {
      label: t('stats.wordsMastered'),
      value: animatedValues.wordsMastered,
      icon: '🏆',
      bgColor: 'bg-white',
      textColor: 'text-[#4CAF50]',
    },
    {
      label: t('stats.accuracyToday'),
      value: animatedValues.accuracyToday,
      suffix: '%',
      icon: '🎯',
      bgColor: 'bg-white',
      textColor: 'text-[#2D3436]',
    },
  ]

  // Chart.js base configuration with animations
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
        backgroundColor: 'rgba(45, 52, 54, 0.9)',
        titleColor: '#FFFFFF',
        bodyColor: '#FFFFFF',
        borderColor: '#FF7B54',
        borderWidth: 1,
        cornerRadius: 8,
        padding: 12,
      },
    },
  }

  // Delayed animation options for staggered effects
  const getDelayedAnimation = (delay: number) => ({
    animation: {
      duration: 1200,
      easing: 'easeInOutQuart' as const,
      delay: (context: { type: string; mode: string; dataIndex: number }) => {
        let datasetDelay = 0
        if (context.type === 'data' && context.mode === 'default') {
          datasetDelay = context.dataIndex * 50 + delay
        }
        return datasetDelay
      },
    },
  })

  // 1. Learning Progress Chart (Line Chart - last 7 days)
  const learningProgressData = {
    labels: ['12/20', '12/21', '12/22', '12/23', '12/24', '12/25', '12/26'],
    datasets: [
      {
        label: t('charts.wordsStudied'),
        data: [15, 12, 18, 22, 16, 25, 20],
        borderColor: '#FF7B54',
        backgroundColor: 'rgba(255, 123, 84, 0.1)',
        pointBackgroundColor: '#FF7B54',
        pointBorderColor: '#FFFFFF',
        pointBorderWidth: 2,
        pointRadius: 6,
        pointHoverRadius: 8,
        tension: 0.4,
        fill: true,
      },
    ],
  }

  // 2. Practice Time Chart (Bar Chart)
  const practiceTimeData = {
    labels: t.raw('charts.weekDays') as string[],
    datasets: [
      {
        label: t('charts.minutes'),
        data: [25, 18, 32, 28, 22, 45, 38],
        backgroundColor: '#4A90E2',
        borderRadius: 4,
      },
    ],
  }

  // 3. Quiz Performance Chart (Bar Chart by quiz type)
  const quizPerformanceData = {
    labels: t.raw('charts.quizTypes') as string[],
    datasets: [
      {
        label: t('charts.successRate'),
        data: [94, 88, 87, 85],
        backgroundColor: '#4CAF50',
        borderRadius: 4,
      },
    ],
  }

  // 4. Current Box Distribution (Bar Chart with gradient colors)
  const boxDistributionData = {
    labels: t.raw('charts.boxLabels') as string[],
    datasets: [
      {
        label: t('charts.words'),
        data: [45, 38, 29, 22, 18, 12, 8],
        backgroundColor: [
          '#EF4444', // red
          '#F97316', // orange
          '#EAB308', // yellow
          '#84CC16', // lime
          '#22C55E', // green
          '#10B981', // emerald
          '#059669', // green-600
        ],
        borderRadius: 4,
      },
    ],
  }

  // 5. Historical Box Distribution (Stacked Bar Chart)
  const historicalBoxData = {
    labels: ['12/20', '12/21', '12/22', '12/23', '12/24', '12/25', '12/26'],
    datasets: [
      {
        label: t('charts.box1'),
        data: [47, 46, 45, 44, 45, 44, 45],
        backgroundColor: '#EF4444',
        stack: 'Stack 0',
      },
      {
        label: t('charts.box2'),
        data: [40, 39, 38, 37, 38, 37, 38],
        backgroundColor: '#F97316',
        stack: 'Stack 0',
      },
      {
        label: t('charts.box3'),
        data: [31, 30, 29, 28, 29, 28, 29],
        backgroundColor: '#EAB308',
        stack: 'Stack 0',
      },
      {
        label: t('charts.box4'),
        data: [24, 23, 22, 21, 22, 21, 22],
        backgroundColor: '#84CC16',
        stack: 'Stack 0',
      },
      {
        label: t('charts.box5'),
        data: [19, 18, 17, 16, 18, 17, 18],
        backgroundColor: '#22C55E',
        stack: 'Stack 0',
      },
      {
        label: t('charts.box6'),
        data: [13, 12, 11, 10, 12, 11, 12],
        backgroundColor: '#10B981',
        stack: 'Stack 0',
      },
      {
        label: t('charts.box7'),
        data: [6, 7, 8, 9, 8, 9, 8],
        backgroundColor: '#059669',
        stack: 'Stack 0',
      },
    ],
  }

  // 6. Word Mastery Progress (Ring/Doughnut Charts)
  const topWordsMastery = [
    { word: 'Bonjour', mastery: 92, color: '#FF7B54' },
    { word: 'Merci', mastery: 88, color: '#4CAF50' },
    { word: 'Au revoir', mastery: 85, color: '#FFD700' },
    { word: "S'il vous plaît", mastery: 79, color: '#9C27B0' },
    { word: 'Excusez-moi', mastery: 74, color: '#2196F3' },
    { word: 'Bonne journée', mastery: 68, color: '#FF6B3D' },
  ]

  // 7. Top Words Section
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
      className="bg-gradient-to-br from-[#FDF6E3] via-[#FFF9F0] to-[#FFE8D6] py-16"
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-12 text-center">
          <h2 className="mb-3 text-3xl font-bold text-[#2D3436] lg:text-4xl">{t('title')}</h2>
          <p className="mx-auto max-w-2xl text-lg text-[#636E72]">{t('subtitle')}</p>
        </div>

        {/* 1. Stats Grid */}
        <div
          data-chart="stats"
          className={`mb-8 rounded-2xl bg-white p-6 shadow-lg transition-all duration-700 ${
            chartsVisible.stats ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
          }`}
        >
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            {statsGrid.map((stat, index) => (
              <div
                key={stat.label}
                className={`${stat.bgColor} rounded-xl p-4 text-center ${
                  stat.bgColor === 'bg-orange-50' ? 'border-2 border-orange-200' : ''
                } transition-all duration-500 ${chartsVisible.stats ? 'scale-100' : 'scale-95'}`}
                style={{
                  transitionDelay: `${index * 100}ms`,
                }}
              >
                <div className="mb-2 text-3xl transition-transform duration-300 hover:scale-110">
                  {stat.icon}
                </div>
                <div
                  className={`text-2xl font-bold ${stat.textColor} mb-1 transition-all duration-1000`}
                >
                  {stat.value.toLocaleString()}
                  {stat.suffix || ''}
                </div>
                <div className="text-sm text-[#636E72]">{stat.label}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Charts Section */}
        <div className="space-y-6">
          {/* Row 1: Learning Progress & Practice Time */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Learning Progress Chart */}
            <div
              data-chart="learningProgress"
              className={`rounded-2xl bg-white p-6 shadow-lg transition-all duration-700 ${
                chartsVisible.learningProgress
                  ? 'translate-y-0 opacity-100'
                  : 'translate-y-4 opacity-0'
              }`}
            >
              <h3 className="mb-6 text-xl font-bold text-[#2D3436]">
                {t('charts.learningProgressTitle')}
              </h3>
              <div className="h-64">
                {chartsVisible.learningProgress && (
                  <Line
                    data={learningProgressData}
                    options={{
                      ...baseChartOptions,
                      ...getDelayedAnimation(0),
                      scales: {
                        x: {
                          grid: { display: false },
                          ticks: { color: '#636E72' },
                        },
                        y: {
                          grid: { color: '#F0F0F0' },
                          ticks: { color: '#636E72' },
                          beginAtZero: true,
                        },
                      },
                    }}
                  />
                )}
              </div>
              <p className="mt-4 text-center text-sm text-[#636E72]">
                {t('charts.wordsStudiedDescription')}
              </p>
            </div>

            {/* Practice Time Chart */}
            <div
              data-chart="practiceTime"
              className={`rounded-2xl bg-white p-6 shadow-lg transition-all duration-700 ${
                chartsVisible.practiceTime ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
              }`}
            >
              <h3 className="mb-4 text-xl font-bold text-[#2D3436]">
                {t('charts.practiceTimeTitle')}
              </h3>
              <div className="mb-4 flex justify-between">
                <div>
                  <span className="text-sm text-[#636E72]">{t('charts.total')}: </span>
                  <span className="font-bold text-[#2D3436]">3h 38m</span>
                </div>
                <div>
                  <span className="text-sm text-[#636E72]">{t('charts.dailyAvg')}: </span>
                  <span className="font-bold text-[#2D3436]">31m</span>
                </div>
              </div>
              <div className="h-56">
                {chartsVisible.practiceTime && (
                  <Bar
                    data={practiceTimeData}
                    options={{
                      ...baseChartOptions,
                      ...getDelayedAnimation(100),
                      plugins: {
                        ...baseChartOptions.plugins,
                      },
                      scales: {
                        x: {
                          grid: { display: false },
                          ticks: { color: '#636E72' },
                        },
                        y: {
                          grid: { color: '#F0F0F0' },
                          ticks: {
                            color: '#636E72',
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
          </div>

          {/* Row 2: Quiz Performance */}
          <div
            data-chart="quizPerformance"
            className={`rounded-2xl bg-white p-6 shadow-lg transition-all duration-700 ${
              chartsVisible.quizPerformance
                ? 'translate-y-0 opacity-100'
                : 'translate-y-4 opacity-0'
            }`}
          >
            <h3 className="mb-6 text-xl font-bold text-[#2D3436]">
              {t('charts.quizPerformanceTitle')}
            </h3>
            <div className="h-64">
              {chartsVisible.quizPerformance && (
                <Bar
                  data={quizPerformanceData}
                  options={{
                    ...baseChartOptions,
                    indexAxis: 'y',
                    plugins: {
                      ...baseChartOptions.plugins,
                    },
                    scales: {
                      x: {
                        grid: { color: '#F0F0F0' },
                        ticks: {
                          color: '#636E72',
                          callback: (value) => `${value}%`,
                        },
                        beginAtZero: true,
                        max: 100,
                      },
                      y: {
                        grid: { display: false },
                        ticks: { color: '#636E72' },
                      },
                    },
                    ...getDelayedAnimation(200),
                  }}
                />
              )}
            </div>
            <div className="mt-4 flex flex-wrap justify-center gap-4">
              {(t.raw('charts.quizTypes') as string[]).map((quiz, index) => (
                <div key={quiz} className="flex items-center space-x-2">
                  <div className="h-3 w-3 rounded-full bg-[#4CAF50]"></div>
                  <span className="text-sm text-[#636E72]">
                    {quiz}: {[94, 88, 87, 85][index]}%
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Row 3: Word Mastery Progress */}
          <div
            data-chart="wordMastery"
            className={`rounded-2xl bg-white p-6 shadow-lg transition-all duration-700 ${
              chartsVisible.wordMastery ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
            }`}
          >
            <h3 className="mb-6 text-xl font-bold text-[#2D3436]">
              {t('charts.wordMasteryTitle')}
            </h3>
            <div className="grid grid-cols-2 gap-6 md:grid-cols-3 lg:grid-cols-6">
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
                  <div className="relative mb-2 inline-flex h-24 w-24 items-center justify-center">
                    <Doughnut
                      data={{
                        datasets: [
                          {
                            data: [word.mastery, 100 - word.mastery],
                            backgroundColor: [word.color, '#F0F0F0'],
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
                      <span className="text-lg font-bold">{word.mastery}%</span>
                    </div>
                  </div>
                  <p className="text-sm font-medium text-[#2D3436]">{word.word}</p>
                </div>
              ))}
            </div>
          </div>

          {/* Row 4: Box Distribution */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Current Box Distribution */}
            <div
              data-chart="boxDistribution"
              className={`rounded-2xl bg-white p-6 shadow-lg transition-all duration-700 ${
                chartsVisible.boxDistribution
                  ? 'translate-y-0 opacity-100'
                  : 'translate-y-4 opacity-0'
              }`}
            >
              <h3 className="mb-4 text-xl font-bold text-[#2D3436]">
                {t('charts.currentBoxTitle')}
              </h3>
              <div className="h-64">
                {chartsVisible.boxDistribution && (
                  <Bar
                    data={boxDistributionData}
                    options={{
                      ...baseChartOptions,
                      ...getDelayedAnimation(300),
                      plugins: {
                        ...baseChartOptions.plugins,
                      },
                      scales: {
                        x: {
                          grid: { display: false },
                          ticks: { color: '#636E72' },
                        },
                        y: {
                          grid: { color: '#F0F0F0' },
                          ticks: { color: '#636E72' },
                          beginAtZero: true,
                        },
                      },
                    }}
                  />
                )}
              </div>
              <div className="mt-4 rounded-lg bg-blue-50 p-4">
                <p className="text-sm text-[#636E72]">{t('charts.leitnerExplanation')}</p>
              </div>
            </div>

            {/* Historical Box Distribution */}
            <div
              data-chart="historicalBox"
              className={`rounded-2xl bg-white p-6 shadow-lg transition-all duration-700 ${
                chartsVisible.historicalBox
                  ? 'translate-y-0 opacity-100'
                  : 'translate-y-4 opacity-0'
              }`}
            >
              <h3 className="mb-4 text-xl font-bold text-[#2D3436]">
                {t('charts.boxProgressTitle')}
              </h3>
              <div className="h-64">
                {chartsVisible.historicalBox && (
                  <Bar
                    data={historicalBoxData}
                    options={{
                      ...baseChartOptions,
                      ...getDelayedAnimation(400),
                      plugins: {
                        ...baseChartOptions.plugins,
                        legend: { display: false },
                      },
                      scales: {
                        x: {
                          stacked: true,
                          grid: { display: false },
                          ticks: { color: '#636E72' },
                        },
                        y: {
                          stacked: true,
                          grid: { color: '#F0F0F0' },
                          ticks: { color: '#636E72' },
                          beginAtZero: true,
                        },
                      },
                    }}
                  />
                )}
              </div>
              <div className="mt-4 text-center">
                <span className="text-sm text-[#636E72]">{t('charts.box7Description')}: </span>
                <span className="font-bold text-[#059669]">8 words</span>
              </div>
            </div>
          </div>

          {/* Row 5: Top Words Leaderboard */}
          <div
            data-chart="topWords"
            className={`rounded-2xl bg-white p-6 shadow-lg transition-all duration-700 ${
              chartsVisible.topWords ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
            }`}
          >
            <h3 className="mb-6 text-xl font-bold text-[#2D3436]">{t('charts.topWordsTitle')}</h3>
            <div className="space-y-3">
              {topWordsList.map((word, index) => (
                <div
                  key={word.rank}
                  className="flex items-center justify-between rounded-xl bg-[#F7F8FA] p-4 transition-all duration-500 hover:shadow-md"
                  style={{
                    transitionDelay: chartsVisible.topWords ? `${index * 100}ms` : '0ms',
                    opacity: chartsVisible.topWords ? 1 : 0,
                    transform: chartsVisible.topWords ? 'translateX(0)' : 'translateX(-20px)',
                  }}
                >
                  <div className="flex items-center space-x-4">
                    <div
                      className={`flex h-10 w-10 items-center justify-center rounded-full font-bold text-white ${
                        word.rank === 1
                          ? 'bg-[#FFD700]'
                          : word.rank === 2
                            ? 'bg-[#C0C0C0]'
                            : word.rank === 3
                              ? 'bg-[#CD7F32]'
                              : 'bg-gray-400'
                      }`}
                    >
                      {word.rank}
                    </div>
                    <span className="text-lg font-semibold text-[#2D3436]">{word.word}</span>
                  </div>
                  <div className="flex items-center space-x-3">
                    <span className="rounded-full bg-[#4CAF50] px-4 py-2 text-sm font-bold text-white">
                      {word.mastery}%
                    </span>
                    <span className="rounded-full bg-gray-200 px-4 py-2 text-sm font-medium text-[#636E72]">
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
