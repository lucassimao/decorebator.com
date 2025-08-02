'use client';

import React, { useState, useEffect } from 'react';
import { Quiz, getRandomQuizSet, getQuizTypeDisplayName } from '@/lib/quiz-data';
import SmartDownloadButton from '@/components/common/SmartDownloadButton';

interface QuizDemoModalProps {
  isOpen: boolean;
  onClose: () => void;
  demoQuizzes: Quiz[];
}

type ErrorType = 'wrong_meaning' | 'image_not_match' | 'image_not_loading' | 'wrong_example' | 'sound_not_playing';

interface ErrorReportModalProps {
  isOpen: boolean;
  onClose: () => void;
  onReportError: (errorType: ErrorType) => void;
  isLoading: boolean;
}

const ErrorReportModal: React.FC<ErrorReportModalProps> = ({ isOpen, onClose, onReportError, isLoading }) => {
  if (!isOpen) return null;

  const errorOptions = [
    { type: 'wrong_meaning' as ErrorType, icon: 'fa-question-circle', label: 'Wrong meaning or definition' },
    { type: 'image_not_match' as ErrorType, icon: 'fa-image', label: 'Image doesn\'t match the word' },
    { type: 'image_not_loading' as ErrorType, icon: 'fa-exclamation-triangle', label: 'Image not loading' },
    { type: 'wrong_example' as ErrorType, icon: 'fa-quote-left', label: 'Wrong example sentence' },
    { type: 'sound_not_playing' as ErrorType, icon: 'fa-volume-mute', label: 'Audio not playing' },
  ];

  return (
    <div className="fixed inset-0 z-[80] flex items-center justify-center">
      <div className="absolute inset-0 bg-black bg-opacity-50" onClick={onClose} />
      <div className="relative bg-white rounded-2xl shadow-2xl w-full max-w-md mx-4 p-6">
        <div className="text-center mb-6">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <i className="fas fa-exclamation-triangle text-red-500 text-xl"></i>
          </div>
          <h3 className="text-xl font-bold text-[#2D3436] mb-2">Report an Issue</h3>
          <p className="text-[#636E72]">What seems to be wrong with this question?</p>
        </div>

        <div className="space-y-3 mb-6">
          {errorOptions.map((option) => (
            <button
              key={option.type}
              onClick={() => onReportError(option.type)}
              disabled={isLoading}
              className="w-full flex items-center space-x-3 p-4 bg-gray-50 hover:bg-gray-100 rounded-xl transition-colors duration-200 text-left disabled:opacity-50"
            >
              <i className={`fas ${option.icon} text-[#636E72] text-lg`}></i>
              <span className="text-[#2D3436]">{option.label}</span>
            </button>
          ))}
        </div>

        <div className="flex justify-center">
          <button
            onClick={onClose}
            disabled={isLoading}
            className="px-6 py-2 text-[#636E72] hover:text-[#2D3436] transition-colors duration-200 disabled:opacity-50"
          >
            Cancel
          </button>
        </div>

        {isLoading && (
          <div className="absolute inset-0 bg-white bg-opacity-80 rounded-2xl flex items-center justify-center">
            <div className="flex items-center space-x-3">
              <i className="fas fa-spinner fa-spin text-[#FF7B54] text-xl"></i>
              <span className="text-[#2D3436]">Reporting issue...</span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

const QuizDemoModal: React.FC<QuizDemoModalProps> = ({ isOpen, onClose, demoQuizzes }) => {
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null);
  const [showFeedback, setShowFeedback] = useState(false);
  const [quizSet, setQuizSet] = useState<Quiz[]>([]);
  const [score, setScore] = useState(0);
  const [isCompleted, setIsCompleted] = useState(false);
  const [isPlayingAudio, setIsPlayingAudio] = useState(false);
  const [fastMode, setFastMode] = useState(false);
  const [showErrorReport, setShowErrorReport] = useState(false);
  const [isReporting, setIsReporting] = useState(false);
  const [userInput, setUserInput] = useState('');
  const [isWriteMode, setIsWriteMode] = useState(false);

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
      setUserInput('');
      setIsWriteMode(randomQuizzes[0]?.type === 'WRITE_WORD_FROM_DEFINITION');
      setShowErrorReport(false);
      setIsReporting(false);
    }
  }, [isOpen, demoQuizzes]);

  const currentQuiz = quizSet[currentQuestionIndex];

  // Update write mode when question changes
  useEffect(() => {
    if (currentQuiz) {
      setIsWriteMode(currentQuiz.type === 'WRITE_WORD_FROM_DEFINITION');
    }
  }, [currentQuiz]);

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

  const handleAnswerSelect = (answerIndex: number) => {
    if (showFeedback) return; // Prevent multiple selections
    
    setSelectedAnswer(answerIndex);
    setShowFeedback(true);
    
    // Check if correct
    const isCorrect = answerIndex === currentQuiz.answerIndex;
    if (isCorrect) {
      setScore(score + 1);
    }

    // Auto-advance in fast mode
    if (fastMode) {
      setTimeout(() => {
        handleNext();
      }, 600);
    }
  };

  const handleWriteAnswer = () => {
    if (!userInput.trim()) return;
    
    setShowFeedback(true);
    
    // Check if correct (case insensitive)
    const isCorrect = userInput.toLowerCase().trim() === currentQuiz.options[currentQuiz.answerIndex]?.toLowerCase();
    if (isCorrect) {
      setScore(score + 1);
      setSelectedAnswer(currentQuiz.answerIndex);
    } else {
      setSelectedAnswer(-1); // Mark as wrong
    }

    // Auto-advance in fast mode
    if (fastMode) {
      setTimeout(() => {
        handleNext();
      }, 600);
    }
  };

  const handleSkipQuestion = () => {
    setShowFeedback(true);
    setSelectedAnswer(-1); // Mark as skipped

    // Auto-advance in fast mode
    if (fastMode) {
      setTimeout(() => {
        handleNext();
      }, 600);
    }
  };

  const handleReportError = (errorType: ErrorType) => {
    setIsReporting(true);
    
    // Log the error type for demo purposes
    console.log('Demo: Reported error type:', errorType);
    
    // Simulate API call
    setTimeout(() => {
      setIsReporting(false);
      setShowErrorReport(false);
      
      // Show success message and advance to next question
      alert('Thank you for your feedback! This will help us improve the content.');
      handleNext();
    }, 1000);
  };

  const toggleFastMode = () => {
    setFastMode(!fastMode);
    
    // If enabling fast mode and feedback is shown, auto-advance
    if (!fastMode && showFeedback) {
      setTimeout(() => {
        handleNext();
      }, 600);
    }
  };

  const handleNext = () => {
    if (currentQuestionIndex < quizSet.length - 1) {
      setCurrentQuestionIndex(currentQuestionIndex + 1);
      setSelectedAnswer(null);
      setShowFeedback(false);
      setUserInput('');
      setIsWriteMode(currentQuiz?.type === 'WRITE_WORD_FROM_DEFINITION');
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
    setUserInput('');
    setIsWriteMode(randomQuizzes[0]?.type === 'WRITE_WORD_FROM_DEFINITION');
    setShowErrorReport(false);
    setIsReporting(false);
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
                  Question {currentQuestionIndex + 1} of {quizSet.length} • Score: {score}/{currentQuestionIndex + (showFeedback ? 1 : 0)}
                </p>
              )}
            </div>
          </div>
          
          <div className="flex items-center space-x-2">
            {!isCompleted && currentQuiz && (
              <button
                onClick={() => setShowErrorReport(true)}
                className="w-10 h-10 rounded-full hover:bg-red-50 flex items-center justify-center transition-colors duration-200"
                title="Report an issue"
              >
                <i className="fas fa-flag text-red-500 text-sm"></i>
              </button>
            )}
            <button
              onClick={onClose}
              className="w-10 h-10 rounded-full hover:bg-gray-100 flex items-center justify-center transition-colors duration-200"
            >
              <i className="fas fa-times text-[#636E72] text-lg"></i>
            </button>
          </div>
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

        {/* Fast Mode Toggle */}
        {!isCompleted && currentQuiz && (
          <div className="flex items-center justify-center py-4 border-b border-gray-100">
            <div className="flex items-center space-x-3">
              <span className="text-sm font-medium text-[#2D3436]">Fast Mode</span>
              <button
                onClick={toggleFastMode}
                className={`relative w-12 h-6 rounded-full transition-colors duration-300 ${
                  fastMode ? 'bg-[#FF7B54]' : 'bg-gray-300'
                }`}
              >
                <div
                  className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform duration-300 ${
                    fastMode ? 'translate-x-7' : 'translate-x-1'
                  }`}
                />
              </button>
              <span className="text-xs text-[#636E72]">
                {fastMode ? 'Auto-advance enabled' : 'Manual progression'}
              </span>
            </div>
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
                
                <SmartDownloadButton 
                  className="bg-gray-100 text-[#2D3436] px-6 py-3 rounded-xl font-semibold hover:bg-gray-200 transition-all duration-300 flex items-center justify-center space-x-2"
                  size="medium"
                  onClick={onClose}
                >
                  <i className="fas fa-download text-sm"></i>
                  <span>Download App</span>
                </SmartDownloadButton>
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

              {/* Answer Options or Write Input */}
              {isWriteMode ? (
                <div className="space-y-4">
                  <div className="space-y-3">
                    <input
                      type="text"
                      value={userInput}
                      onChange={(e) => setUserInput(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && !showFeedback && handleWriteAnswer()}
                      placeholder="Type your answer here..."
                      disabled={showFeedback}
                      className="w-full p-4 border-2 border-gray-200 rounded-xl text-lg focus:border-[#FF7B54] focus:outline-none transition-colors duration-200 disabled:bg-gray-50"
                    />
                    {showFeedback && (
                      <div className="p-4 rounded-xl border-2 border-green-500 bg-green-50">
                        <div className="flex items-center justify-between">
                          <span className="text-green-700 font-semibold">
                            Correct answer: {currentQuiz.options[currentQuiz.answerIndex]}
                          </span>
                          <i className="fas fa-check-circle text-green-500 text-xl"></i>
                        </div>
                      </div>
                    )}
                  </div>
                  
                  {!showFeedback && (
                    <div className="flex gap-3">
                      <button
                        onClick={handleWriteAnswer}
                        disabled={!userInput.trim()}
                        className="flex-1 bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-3 rounded-xl font-semibold hover:shadow-lg transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Submit Answer
                      </button>
                      <button
                        onClick={handleSkipQuestion}
                        className="px-6 py-3 border-2 border-gray-300 text-gray-600 rounded-xl font-semibold hover:border-gray-400 transition-colors duration-200"
                      >
                        Skip
                      </button>
                    </div>
                  )}
                </div>
              ) : (
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
              )}

              {/* Action Buttons */}
              <div className="flex justify-center pt-6">
                {showFeedback ? (
                  // Next button always visible after answering
                  <button
                    onClick={handleNext}
                    className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-3 rounded-xl font-semibold hover:shadow-lg transition-all duration-300 flex items-center space-x-2"
                  >
                    <span>
                      {currentQuestionIndex < quizSet.length - 1 ? 'Next Question' : 'Complete Quiz'}
                    </span>
                    <i className="fas fa-arrow-right text-sm"></i>
                  </button>
                ) : !isWriteMode ? (
                  // Skip button for multiple choice questions
                  <button
                    onClick={handleSkipQuestion}
                    className="px-6 py-3 border-2 border-gray-300 text-gray-600 rounded-xl font-semibold hover:border-gray-400 transition-colors duration-200"
                  >
                    Skip Question
                  </button>
                ) : null}
              </div>
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

        {/* Error Report Modal */}
        <ErrorReportModal
          isOpen={showErrorReport}
          onClose={() => setShowErrorReport(false)}
          onReportError={handleReportError}
          isLoading={isReporting}
        />
      </div>
    </div>
  );
};

export default QuizDemoModal;