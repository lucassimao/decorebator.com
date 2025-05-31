import UpgradePromptDialog from "@/components/dashboard/UpgradePromptDialog";
import React, {
  createContext,
  useContext,
  useState,
  ReactNode,
  useCallback,
  useMemo,
} from "react";

interface ContextType {
  show: () => void;
}

const Context = createContext<ContextType | undefined>(undefined);

export const useUpgradePromptDialog = () => {
  const context = useContext(Context);
  if (!context) {
    throw new Error(
      "useUpgradePromptDialog must be used within UpgradePromptDialogProvider",
    );
  }
  return context;
};

export const UpgradePromptDialogProvider: React.FC<{ children: ReactNode }> = ({
  children,
}) => {
  const [visible, setVisible] = useState<boolean>(false);

  const show = useCallback(() => setVisible(true), []);

  const hide = useCallback(() => setVisible(false), []);

  // now this object identity won’t change unless `show` changes
  const contextValue = useMemo(() => ({ show }), [show]);

  return (
    <Context.Provider value={contextValue}>
      {children}
      {visible && <UpgradePromptDialog visible onDismiss={hide} />}
    </Context.Provider>
  );
};
