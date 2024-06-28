import { StyleSheet } from "react-native";
import { Portal, Snackbar, Text } from "react-native-paper";

export type SnackBarProps = {
  message: string;
  type: "success" | "error"|'info';
  onDismiss: () => void;
};
export default function Component({ message, type, onDismiss }: SnackBarProps) {
  return (
    <Portal>
      <Snackbar
        visible={true}
        duration={2000}
        onDismiss={onDismiss}
        style={[
          styles.snackbar,
          styles[type]
        ]}
        action={{
          label: "Hide",
        }}
      >
        <Text style={styles.snackbarText}>{message}</Text>
      </Snackbar>
    </Portal>
  );
}

const styles = StyleSheet.create({
  snackbar: {
    borderRadius: 4,
  },
  success: {
    backgroundColor: "#4caf50", // Green color for success
  },
  error: {
    backgroundColor: "#f44336", // Red color for error
  },
  snackbarText: {
    color: "#ffffff", // White text color
  },
  info:{

  }
});
