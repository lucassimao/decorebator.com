'use client';

import React, { useState, useEffect } from 'react';
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
} from 'chart.js';
import { Line, Bar, Doughnut, PolarArea } from 'react-chartjs-2';

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
);

const AnalyticsSection: React.FC = () => {
  const [animatedValues, setAnimatedValues] = useState({
    totalWords: 0,
    wordsMastered: 0,
    averageAccuracy: 0,
    streakDays: 0,
  });

  // Animate numbers on mount
  useEffect(() => {
    const timer = setTimeout(() => {
      setAnimatedValues({
        totalWords: 2847,
        wordsMastered: 1923,
        averageAccuracy: 89,
        streakDays: 42,
      });
    }, 500);

    return () => clearTimeout(timer);
  }, []);


  const dashboardStats = [
    {
      label: 'Words Studied Today',
      value: 12,
      suffix: '',
      icon: <i className="fas fa-book text-[#FF7B54]"></i>,
      description: 'Today\'s learning activity',
    },
    {
      label: 'Current Streak',
      value: animatedValues.streakDays,
      suffix: ' days',
      icon: <i className="fas fa-fire text-[#FF7B54]"></i>,
      description: 'Daily study streak',
    },
    {
      label: 'Words Mastered',
      value: animatedValues.wordsMastered,
      suffix: '',
      icon: <i className="fas fa-trophy text-[#FFD700]"></i>,
      description: '80%+ mastery level',
    },
    {
      label: 'Today\'s Accuracy',
      value: animatedValues.averageAccuracy,
      suffix: '%',
      icon: <i className="fas fa-bullseye text-[#4CAF50]"></i>,
      description: 'Success rate today',
    },
  ];


  // Chart.js configurations matching mobile app theme
  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
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
      },
    },
    scales: {
      x: {
        grid: {
          display: false,
        },
        ticks: {
          color: '#636E72',
          font: {
            size: 12,
          },
        },
      },
      y: {
        grid: {
          color: '#F0F0F0',
        },
        ticks: {
          color: '#636E72',
          font: {
            size: 12,
          },
        },
      },
    },
  };

  // Learning Progress Chart Data (Line Chart)
  const learningProgressData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Words Studied',
        data: [15, 12, 18, 22, 16, 25, 20],
        borderColor: '#FF7B54',
        backgroundColor: 'rgba(255, 123, 84, 0.1)',
        pointBackgroundColor: '#FFFFFF',
        pointBorderColor: '#FF7B54',
        pointBorderWidth: 2,
        pointRadius: 6,
        tension: 0.4,
        fill: true,
      },
    ],
  };

  // Practice Time Chart Data (Bar Chart)
  const practiceTimeData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Practice Time (minutes)',
        data: [25, 18, 32, 28, 22, 45, 38],
        backgroundColor: 'rgba(156, 39, 176, 0.8)',
        borderColor: '#9C27B0',
        borderWidth: 1,
        borderRadius: 4,
      },
    ],
  };

  // Box Distribution Chart Data (Doughnut Chart)
  const boxDistributionData = {
    labels: ['Box 1 (New)', 'Box 2 (Learning)', 'Box 3 (Progress)', 'Box 4 (Familiar)', 'Box 5 (Known)', 'Box 6 (Strong)', 'Box 7 (Mastered)'],
    datasets: [
      {
        data: [45, 38, 29, 22, 18, 12, 8],
        backgroundColor: [
          '#EF4444', // red-500
          '#F97316', // orange-500
          '#EAB308', // yellow-500
          '#84CC16', // lime-500
          '#22C55E', // green-500
          '#10B981', // emerald-500
          '#059669', // green-600
        ],
        borderWidth: 0,
        cutout: '50%',
      },
    ],
  };


  // Word Mastery Chart Data (Polar Area Chart)
  const wordMasteryData = {
    labels: ['Bonjour', 'Hola', 'Ciao', 'Guten Tag', 'Kon\'nichiwa', 'Olá'],
    datasets: [
      {
        label: 'Mastery Level',
        data: [92, 88, 85, 79, 74, 68],
        backgroundColor: [
          'rgba(255, 123, 84, 0.8)',
          'rgba(76, 175, 80, 0.8)',
          'rgba(255, 215, 0, 0.8)',
          'rgba(156, 39, 176, 0.8)',
          'rgba(33, 150, 243, 0.8)',
          'rgba(255, 107, 61, 0.8)',
        ],
        borderColor: [
          '#FF7B54',
          '#4CAF50',
          '#FFD700',
          '#9C27B0',
          '#2196F3',
          '#FF6B3D',
        ],
        borderWidth: 2,
      },
    ],
  };

  // Historical Box Distribution Data (Stacked Bar Chart)
  const historicalBoxData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      { label: 'Box 1', data: [47, 46, 45, 44, 45, 44, 45], backgroundColor: '#EF4444', stack: 'Stack 0' },
      { label: 'Box 2', data: [40, 39, 38, 37, 38, 37, 38], backgroundColor: '#F97316', stack: 'Stack 0' },
      { label: 'Box 3', data: [31, 30, 29, 28, 29, 28, 29], backgroundColor: '#EAB308', stack: 'Stack 0' },
      { label: 'Box 4', data: [24, 23, 22, 21, 22, 21, 22], backgroundColor: '#84CC16', stack: 'Stack 0' },
      { label: 'Box 5', data: [19, 18, 17, 16, 18, 17, 18], backgroundColor: '#22C55E', stack: 'Stack 0' },
      { label: 'Box 6', data: [13, 12, 11, 10, 12, 11, 12], backgroundColor: '#10B981', stack: 'Stack 0' },
      { label: 'Box 7', data: [6, 7, 8, 9, 8, 9, 8], backgroundColor: '#059669', stack: 'Stack 0' },
    ],
  };

  // Top Words Data (not a chart, but a leaderboard)
  const topWords = [
    { rank: 1, word: 'Bonjour', mastery: 92, box: 7 },
    { rank: 2, word: 'Hola', mastery: 88, box: 6 },
    { rank: 3, word: 'Ciao', mastery: 85, box: 6 },
    { rank: 4, word: 'Guten Tag', mastery: 79, box: 5 },
    { rank: 5, word: 'Kon\'nichiwa', mastery: 74, box: 5 },
  ];

  return (
    <section id="analytics" className="py-16 bg-gradient-to-br from-[#FDF6E3] to-orange-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="text-center mb-12">
          <h2 className="text-3xl lg:text-4xl font-bold mb-3">
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">Advanced Analytics</span>
          </h2>
          <p className="text-lg text-[#636E72] max-w-2xl mx-auto">
            9 powerful analytics tools to track your vocabulary mastery and learning progress
          </p>
        </div>

        {/* Dashboard Stats */}
        <div className="bg-white rounded-2xl shadow-lg p-6 mb-8">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {dashboardStats.map((stat) => (
              <div key={stat.label} className="text-center">
                <div className="text-2xl mb-2">{stat.icon}</div>
                <div className="text-2xl font-bold text-[#2D3436] mb-1">
                  {stat.value.toLocaleString()}{stat.suffix}
                </div>
                <div className="text-sm text-[#636E72]">{stat.label}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Charts Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6 mb-8">
          
          {/* Learning Progress Chart */}
          <div className="bg-white rounded-2xl shadow-lg p-6">
            <h3 className="text-lg font-bold text-[#2D3436] mb-4">Learning Progress</h3>
            <div className="h-48">
              <Line 
                data={learningProgressData} 
                options={{
                  ...chartOptions,
                  plugins: {
                    ...chartOptions.plugins,
                    tooltip: {
                      ...chartOptions.plugins.tooltip,
                      callbacks: {
                        label: function(context: { parsed: { y: number } }) {
                          return `Words: ${context.parsed.y}`;
                        }
                      }
                    },
                  },
                  scales: {
                    ...chartOptions.scales,
                    y: {
                      ...chartOptions.scales.y,
                      beginAtZero: true,
                      max: 30,
                    },
                  },
                }} 
              />
            </div>
          </div>

          {/* Practice Time Chart */}
          <div className="bg-white rounded-2xl shadow-lg p-6">
            <h3 className="text-lg font-bold text-[#2D3436] mb-4">Practice Time</h3>
            <div className="h-48">
              <Bar 
                data={practiceTimeData} 
                options={{
                  ...chartOptions,
                  plugins: {
                    ...chartOptions.plugins,
                    tooltip: {
                      ...chartOptions.plugins.tooltip,
                      callbacks: {
                        label: function(context: { parsed: { y: number } }) {
                          return `${context.parsed.y} minutes`;
                        }
                      }
                    },
                  },
                  scales: {
                    ...chartOptions.scales,
                    y: {
                      ...chartOptions.scales.y,
                      beginAtZero: true,
                      max: 50,
                      ticks: {
                        ...chartOptions.scales.y.ticks,
                        callback: function(value: string | number) {
                          return value + 'm';
                        }
                      }
                    },
                  },
                }} 
              />
            </div>
          </div>

          {/* Box Distribution Chart */}
          <div className="bg-white rounded-2xl shadow-lg p-6">
            <h3 className="text-lg font-bold text-[#2D3436] mb-4">Box Distribution</h3>
            <div className="h-48 flex items-center justify-center">
              <Doughnut 
                data={boxDistributionData} 
                options={{
                  ...chartOptions,
                  plugins: {
                    ...chartOptions.plugins,
                    legend: { display: false },
                    tooltip: {
                      ...chartOptions.plugins.tooltip,
                      callbacks: {
                        label: function(context: { label?: string; parsed: number; dataset: { data: number[] } }) {
                          const label = context.label || '';
                          const value = context.parsed;
                          return `${label}: ${value} words`;
                        }
                      }
                    },
                  },
                  scales: undefined,
                }} 
              />
            </div>
          </div>

          {/* Word Mastery Chart */}
          <div className="bg-white rounded-2xl shadow-lg p-6">
            <h3 className="text-lg font-bold text-[#2D3436] mb-4">Word Mastery</h3>
            <div className="h-48 flex items-center justify-center">
              <PolarArea 
                data={wordMasteryData} 
                options={{
                  ...chartOptions,
                  plugins: {
                    ...chartOptions.plugins,
                    legend: { display: false },
                    tooltip: {
                      ...chartOptions.plugins.tooltip,
                      callbacks: {
                        label: function(context: { label?: string; parsed: { r: number } }) {
                          const label = context.label || '';
                          const value = context.parsed.r;
                          return `${label}: ${value}%`;
                        }
                      }
                    },
                  },
                  scales: {
                    r: {
                      beginAtZero: true,
                      max: 100,
                      ticks: { display: false },
                      grid: { color: '#F0F0F0' },
                      pointLabels: { display: false },
                    },
                  },
                }} 
              />
            </div>
          </div>

          {/* Quiz Performance Chart */}
          <div className="bg-white rounded-2xl shadow-lg p-6">
            <h3 className="text-lg font-bold text-[#2D3436] mb-4">Quiz Performance</h3>
            <div className="h-48">
              <Bar 
                data={{
                  labels: ['Guess', 'Meaning', 'Image', 'Audio'],
                  datasets: [{
                    data: [94, 88, 87, 85],
                    backgroundColor: 'rgba(255, 123, 84, 0.8)',
                    borderRadius: 4,
                  }]
                }}
                options={{
                  ...chartOptions,
                  plugins: {
                    ...chartOptions.plugins,
                    tooltip: {
                      ...chartOptions.plugins.tooltip,
                      callbacks: {
                        label: function(context: { parsed: { y: number } }) {
                          return `${context.parsed.y}% accuracy`;
                        }
                      }
                    },
                  },
                  scales: {
                    ...chartOptions.scales,
                    y: {
                      ...chartOptions.scales.y,
                      beginAtZero: true,
                      max: 100,
                      ticks: {
                        ...chartOptions.scales.y.ticks,
                        callback: function(value: string | number) {
                          return value + '%';
                        }
                      }
                    },
                  },
                }} 
              />
            </div>
          </div>

          {/* Historical Box Distribution Chart */}
          <div className="bg-white rounded-2xl shadow-lg p-6">
            <h3 className="text-lg font-bold text-[#2D3436] mb-4">Historical Progress</h3>
            <div className="h-48">
              <Bar 
                data={historicalBoxData} 
                options={{
                  ...chartOptions,
                  plugins: {
                    ...chartOptions.plugins,
                    legend: { display: false },
                  },
                  scales: {
                    ...chartOptions.scales,
                    x: { stacked: true },
                    y: { 
                      stacked: true,
                      beginAtZero: true,
                    },
                  },
                }} 
              />
            </div>
          </div>

        </div>

        {/* Top Words Leaderboard */}
        <div className="bg-white rounded-2xl shadow-lg p-6">
          <h3 className="text-xl font-bold text-[#2D3436] mb-6">Top Words Leaderboard</h3>
          <div className="space-y-3">
            {topWords.map((word) => (
              <div key={word.rank} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="flex items-center space-x-4">
                  <div className={`w-8 h-8 rounded-full flex items-center justify-center text-white font-bold ${
                    word.rank === 1 ? 'bg-[#FFD700]' : 'bg-gray-400'
                  }`}>
                    {word.rank}
                  </div>
                  <span className="font-semibold text-[#2D3436]">{word.word}</span>
                </div>
                <div className="flex items-center space-x-4">
                  <span className="bg-[#4CAF50] text-white px-3 py-1 rounded-full text-sm font-bold">
                    {word.mastery}%
                  </span>
                  <span className="bg-gray-200 text-[#636E72] px-3 py-1 rounded-full text-sm">
                    Box {word.box}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>


      </div>
    </section>
  );
};

export default AnalyticsSection;