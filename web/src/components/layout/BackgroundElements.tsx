import React from 'react';

const BackgroundElements: React.FC = () => {
  return (
    <div className="fixed inset-0 overflow-hidden pointer-events-none">
      <div className="absolute top-20 left-10 w-64 h-64 bg-orange-300 rounded-full opacity-20 blur-3xl float-animation"></div>
      <div className="absolute bottom-20 right-10 w-96 h-96 bg-yellow-300 rounded-full opacity-20 blur-3xl float-animation" style={{animationDelay: '3s'}}></div>
      <div className="absolute top-1/2 left-1/2 w-80 h-80 bg-amber-300 rounded-full opacity-10 blur-3xl pulse-glow"></div>
    </div>
  );
};

export default BackgroundElements;