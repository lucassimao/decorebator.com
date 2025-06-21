'use client';

import React, { useState, useEffect } from 'react';
import { Quiz, getRandomQuizSet, getQuizTypeDisplayName } from '@/lib/quiz-data';

interface QuizDemoModalProps {
  isOpen: boolean;
  onClose: () => void;
  demoQuizzes: Quiz[];
}

const QuizDemoModal: React.FC<QuizDemoModalProps> = ({ isOpen, onClose, demoQuizzes }) => {
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null);
  const [showFeedback, setShowFeedback] = useState(false);
  const [quizSet, setQuizSet] = useState<Quiz[]>([]);
  const [score, setScore] = useState(0);
  const [isCompleted, setIsCompleted] = useState(false);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);

  // Initialize quiz set when modal opens
  useEffect(() => {
    if (isOpen && demoQuizzes.length > 0) {
      const randomQuizzes = getRandomQuizSet(demoQuizzes, 5);
      setQuizSet(randomQuizzes);
      setCurrentQuestionIndex(0);
      setSelectedAnswer(null);
      setShowFeedback(false);
      setScore(0);
      setIsCompleted(false);
    }
  }, [isOpen, demoQuizzes]);

  // Cleanup audio when modal closes or question changes
  useEffect(() => {
    if (!isOpen) {
      // Stop any playing audio when modal closes
      if ('speechSynthesis' in window) {
        speechSynthesis.cancel();
      }
      setIsPlayingAudio(false);
    }
  }, [isOpen]);

  // Stop audio when navigating to different questions
  useEffect(() => {
    if ('speechSynthesis' in window) {
      speechSynthesis.cancel();
    }
    setIsPlayingAudio(false);
  }, [currentQuestionIndex]);

  const currentQuiz = quizSet[currentQuestionIndex];

  const handleAnswerSelect = (answerIndex: number) => {
    if (showFeedback) return; // Prevent multiple selections
    
    setSelectedAnswer(answerIndex);
    setShowFeedback(true);
    
    // Check if correct
    const isCorrect = answerIndex === currentQuiz.answerIndex;
    if (isCorrect) {
      setScore(score + 1);
    }
  };

  const handleNext = () => {
    if (currentQuestionIndex < quizSet.length - 1) {
      setCurrentQuestionIndex(currentQuestionIndex + 1);
      setSelectedAnswer(null);
      setShowFeedback(false);
    } else {
      setIsCompleted(true);
    }
  };

  const handleRestart = () => {
    const randomQuizzes = getRandomQuizSet(demoQuizzes, 5);
    setQuizSet(randomQuizzes);
    setCurrentQuestionIndex(0);
    setSelectedAnswer(null);
    setShowFeedback(false);
    setScore(0);
    setIsCompleted(false);
  };

  // Audio playing functionality using Web Speech API
  const playAudio = (text: string) => {
    if (isPlayingAudio) return;
    
    setIsPlayingAudio(true);
    
    // Use Web Speech API for text-to-speech
    if ('speechSynthesis' in window) {
      const utterance = new SpeechSynthesisUtterance(text);
      utterance.rate = 0.8; // Slightly slower for learning
      utterance.volume = 0.8;
      
      utterance.onend = () => {
        setIsPlayingAudio(false);
      };
      
      utterance.onerror = () => {
        setIsPlayingAudio(false);
      };
      
      speechSynthesis.speak(utterance);
    } else {
      // Fallback for browsers without speech synthesis
      console.log('Playing audio for:', text);
      setTimeout(() => {
        setIsPlayingAudio(false);
      }, 2000);
    }
  };

  const getQuizTitle = () => {
    if (!currentQuiz) return '';
    
    switch (currentQuiz.type) {
      case 'GUESS_MEANING':
        return 'What does this word mean?';
      case 'WORD_FROM_MEANING':
        return 'Which word matches this definition?';
      case 'WORD_FROM_IMAGE':
        return 'Which word does this image represent?';
      case 'COMPLETE_SENTENCE':
        return 'Complete the sentence:';
      case 'WRITE_WORD_FROM_DEFINITION':
        return 'What word matches this definition?';
      case 'WORD_FROM_AUDIO':
        return 'Which word did you hear?';
      case 'WORD_FROM_EXAMPLE_AUDIO':
        return 'Complete the sentence:';
      default:
        return 'Choose the correct answer:';
    }
  };

  const getOptionStyle = (index: number) => {
    let baseClasses = "flex items-center justify-between w-full p-4 rounded-xl border-2 transition-all duration-300 text-left ";
    
    if (!showFeedback) {
      baseClasses += "border-gray-200 bg-gray-50 hover:border-[#FF7B54] hover:bg-orange-50 ";
    } else {
      if (index === currentQuiz.answerIndex) {
        baseClasses += "border-green-500 bg-green-50 ";
      } else if (index === selectedAnswer && selectedAnswer !== currentQuiz.answerIndex) {
        baseClasses += "border-red-500 bg-red-50 ";
      } else {
        baseClasses += "border-gray-200 bg-gray-50 opacity-60 ";
      }
    }
    
    return baseClasses;
  };

  const getOptionTextStyle = (index: number) => {
    let baseClasses = "flex-1 text-lg ";
    
    if (!showFeedback) {
      baseClasses += "text-[#2D3436] ";
    } else {
      if (index === currentQuiz.answerIndex) {
        baseClasses += "text-green-700 font-semibold ";
      } else if (index === selectedAnswer && selectedAnswer !== currentQuiz.answerIndex) {
        baseClasses += "text-red-700 ";
      } else {
        baseClasses += "text-[#636E72] ";
      }
    }
    
    return baseClasses;
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black bg-opacity-50 backdrop-blur-sm"
        onClick={onClose}
      />
      
      {/* Modal Content */}
      <div className="relative bg-white rounded-3xl shadow-2xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-100">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-xl flex items-center justify-center">
              <i className="fas fa-brain text-white text-lg"></i>
            </div>
            <div>
              <h2 className="text-xl font-bold text-[#2D3436]">Quick Quiz Demo</h2>
              {!isCompleted && quizSet.length > 0 && (
                <p className="text-sm text-[#636E72]">
                  Question {currentQuestionIndex + 1} of {quizSet.length}
                </p>
              )}
            </div>
          </div>
          
          <button
            onClick={onClose}
            className="w-10 h-10 rounded-full hover:bg-gray-100 flex items-center justify-center transition-colors duration-200"
          >
            <i className="fas fa-times text-[#636E72] text-lg"></i>
          </button>
        </div>

        {/* Progress Bar */}
        {!isCompleted && quizSet.length > 0 && (
          <div className="h-2 bg-gray-100">
            <div 
              className="h-full bg-gradient-to-r from-[#FF7B54] to-orange-600 transition-all duration-500"
              style={{ width: `${((currentQuestionIndex + 1) / quizSet.length) * 100}%` }}
            />
          </div>
        )}

        {/* Quiz Content */}
        <div className="p-8">
          {isCompleted ? (
            // Completion Screen
            <div className="text-center space-y-6">
              <div className="w-20 h-20 bg-gradient-to-br from-green-400 to-green-600 rounded-full flex items-center justify-center mx-auto">
                <i className="fas fa-trophy text-white text-2xl"></i>
              </div>
              
              <div>
                <h3 className="text-2xl font-bold text-[#2D3436] mb-2">
                  Quiz Complete!
                </h3>
                <p className="text-lg text-[#636E72]">
                  You scored {score} out of {quizSet.length} questions
                </p>
              </div>

              <div className="flex flex-col sm:flex-row gap-3 justify-center">
                <button
                  onClick={handleRestart}
                  className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-3 rounded-xl font-semibold hover:shadow-lg transition-all duration-300 flex items-center justify-center space-x-2"
                >
                  <i className="fas fa-redo text-sm"></i>
                  <span>Try Again</span>
                </button>
                
                <a 
                  href="#" 
                  className="bg-gray-100 text-[#2D3436] px-6 py-3 rounded-xl font-semibold hover:bg-gray-200 transition-all duration-300 flex items-center justify-center space-x-2"
                >
                  <i className="fas fa-download text-sm"></i>
                  <span>Download App</span>
                </a>
              </div>
            </div>
          ) : currentQuiz ? (
            // Quiz Question
            <div className="space-y-6">
              {/* Quiz Type Badge */}
              <div className="flex justify-center">
                <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-orange-100 text-[#FF7B54]">
                  {getQuizTypeDisplayName(currentQuiz.type)}
                </span>
              </div>

              {/* Question Title */}
              <h3 className="text-xl font-semibold text-[#2D3436] text-center">
                {getQuizTitle()}
              </h3>

              {/* Word/Question Display */}
              <div className="text-center space-y-4">
                {currentQuiz.type === 'GUESS_MEANING' ? (
                  <div>
                    <div className="text-4xl font-bold text-[#FF7B54] mb-2">
                      {currentQuiz.value}
                    </div>
                    {currentQuiz.pronunciation && (
                      <div className="text-lg text-[#636E72] font-mono">
                        {currentQuiz.pronunciation}
                      </div>
                    )}
                  </div>
                ) : currentQuiz.type === 'WORD_FROM_AUDIO' ? (
                  // Only show play button for audio questions - no word displayed
                  <div className="bg-blue-50 rounded-xl p-8 text-center">
                    <i className="fas fa-headphones text-4xl text-[#FF7B54] mb-4"></i>
                    <p className="text-lg text-[#2D3436] mb-4">
                      Listen to the audio and select the word you hear
                    </p>
                    <button 
                      onClick={() => playAudio(currentQuiz.value)}
                      disabled={isPlayingAudio}
                      className={`mx-auto bg-gradient-to-br from-[#FF7B54] to-orange-600 w-20 h-20 rounded-full flex items-center justify-center hover:shadow-lg transition-all duration-300 hover:scale-105 ${
                        isPlayingAudio ? 'opacity-70 cursor-not-allowed' : ''
                      }`}
                    >
                      {isPlayingAudio ? (
                        <i className="fas fa-spinner fa-spin text-white text-2xl"></i>
                      ) : (
                        <i className="fas fa-play text-white text-2xl"></i>
                      )}
                    </button>
                  </div>
                ) : currentQuiz.type === 'WORD_FROM_IMAGE' ? (
                  <div className="bg-gray-100 rounded-xl p-8 text-center">
                    <i className="fas fa-image text-4xl text-[#636E72] mb-2"></i>
                    <p className="text-[#636E72] italic">
                      {currentQuiz.imageDescription || 'Image would be displayed here'}
                    </p>
                  </div>
                ) : currentQuiz.type === 'WORD_FROM_MEANING' || currentQuiz.type === 'WRITE_WORD_FROM_DEFINITION' ? (
                  <div className="bg-blue-50 rounded-xl p-6">
                    <p className="text-lg text-[#2D3436]">
                      {currentQuiz.value}
                    </p>
                  </div>
                ) : currentQuiz.type === 'WORD_FROM_EXAMPLE_AUDIO' ? (
                  // For example audio, show the sentence with blank and play button
                  <div className="bg-purple-50 rounded-xl p-6 text-center">
                    <p className="text-lg text-[#2D3436] mb-4">
                      {currentQuiz.value}
                    </p>
                    <button 
                      onClick={() => playAudio(currentQuiz.value)}
                      disabled={isPlayingAudio}
                      className={`mx-auto bg-gradient-to-br from-[#FF7B54] to-orange-600 w-16 h-16 rounded-full flex items-center justify-center hover:shadow-lg transition-all duration-300 hover:scale-105 ${
                        isPlayingAudio ? 'opacity-70 cursor-not-allowed' : ''
                      }`}
                    >
                      {isPlayingAudio ? (
                        <i className="fas fa-spinner fa-spin text-white text-lg"></i>
                      ) : (
                        <i className="fas fa-play text-white text-lg"></i>
                      )}
                    </button>
                  </div>
                ) : (
                  <div className="bg-purple-50 rounded-xl p-6">
                    <p className="text-lg text-[#2D3436]">
                      {currentQuiz.value}
                    </p>
                  </div>
                )}
              </div>

              {/* Answer Options */}
              <div className="space-y-3">
                {currentQuiz.options.map((option, index) => (
                  <button
                    key={index}
                    onClick={() => handleAnswerSelect(index)}
                    className={getOptionStyle(index)}
                    disabled={showFeedback}
                  >
                    <span className={getOptionTextStyle(index)}>
                      {option}
                    </span>
                    
                    {showFeedback && index === currentQuiz.answerIndex && (
                      <i className="fas fa-check-circle text-green-500 text-xl ml-3"></i>
                    )}
                    
                    {showFeedback && index === selectedAnswer && selectedAnswer !== currentQuiz.answerIndex && (
                      <i className="fas fa-times-circle text-red-500 text-xl ml-3"></i>
                    )}
                  </button>
                ))}
              </div>

              {/* Next Button */}
              {showFeedback && (
                <div className="flex justify-center pt-4">
                  <button
                    onClick={handleNext}
                    className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-3 rounded-xl font-semibold hover:shadow-lg transition-all duration-300 flex items-center space-x-2"
                  >
                    <span>
                      {currentQuestionIndex < quizSet.length - 1 ? 'Next Question' : 'Complete Quiz'}
                    </span>
                    <i className="fas fa-arrow-right text-sm"></i>
                  </button>
                </div>
              )}
            </div>
          ) : (
            // Loading or Error State
            <div className="text-center space-y-4">
              <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto">
                <i className="fas fa-exclamation-triangle text-white text-xl"></i>
              </div>
              <div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-2">
                  Demo Unavailable
                </h3>
                <p className="text-[#636E72] mb-4">
                  Quiz demo is temporarily unavailable
                </p>
                <button
                  onClick={onClose}
                  className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-3 rounded-xl font-semibold hover:shadow-lg transition-all duration-300"
                >
                  Close
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default QuizDemoModal;