
import SnackBar from '@/components/SnackBar';
import React, { createContext, useContext, useState, ReactNode } from 'react';

interface SnackbarContextType {
  show: (message: string, type: 'success' | 'error', duration?: number) => void;
}

const SnackbarContext = createContext<SnackbarContextType | undefined>(undefined);

export const useSnackbar = () => {
  const context = useContext(SnackbarContext);
  if (!context) {
    throw new Error('useSnackbar must be used within SnackbarProvider');
  }
  return context;
};

interface SnackbarState {
  visible: boolean;
  message: string;
  type: 'success' | 'error';
  duration: number;
}

export const SnackbarProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [snackbar, setSnackbar] = useState<SnackbarState>({
    visible: false,
    message: '',
    type: 'success',
    duration: 2000,
  });

  const show = (message: string, type: 'success' | 'error' = 'success', duration = 2000) => {
    setSnackbar({
      visible: true,
      message,
      type,
      duration,
    });
  };

  const hide = () => {
    setSnackbar(prev => ({ ...prev, visible: false }));
  };

  return (
    <SnackbarContext.Provider value={{ show }}>
      {children}
      <SnackBar
        visible={snackbar.visible}
        message={snackbar.message}
        type={snackbar.type}
        duration={snackbar.duration}
        onDismiss={hide}
      />
    </SnackbarContext.Provider>
  );
};