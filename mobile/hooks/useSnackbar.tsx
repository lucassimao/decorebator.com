import SnackBar from "@/components/SnackBar";
import React, {
  createContext,
  useContext,
  useState,
  ReactNode,
  useCallback,
  useMemo,
} from "react";

interface SnackbarContextType {
  show: (message: string, type: "success" | "error", duration?: number) => void;
}

const SnackbarContext = createContext<SnackbarContextType | undefined>(
  undefined,
);

export const useSnackbar = () => {
  const context = useContext(SnackbarContext);
  if (!context) {
    throw new Error("useSnackbar must be used within SnackbarProvider");
  }
  return context;
};

interface SnackbarState {
  visible: boolean;
  message: string;
  type: "success" | "error";
  duration: number;
}

export const SnackbarProvider: React.FC<{ children: ReactNode }> = ({
  children,
}) => {
  const [snackbar, setSnackbar] = useState<SnackbarState>({
    visible: false,
    message: "",
    type: "success",
    duration: 2000,
  });

  const show = useCallback(
    (
      message: string,
      type: "success" | "error" = "success",
      duration = 2000,
    ) => {
      setSnackbar({ visible: true, message, type, duration });
    },
    [],
  );

  const hide = useCallback(() => {
    setSnackbar((prev) => ({ ...prev, visible: false }));
  }, []);

  // now this object identity won’t change unless `show` changes
  const contextValue = useMemo(() => ({ show }), [show]);

  return (
    <SnackbarContext.Provider value={contextValue}>
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
