import { StyleSheet } from "react-native";
import { Portal, Snackbar, Text, useTheme } from "react-native-paper";

export type SnackBarProps = {
  message: string;
  type: "success" | "error" 
  onDismiss: () => void;
  duration?: number;
};
export default function Component({
  message,
  type,
  onDismiss,
  duration = 2000,
}: SnackBarProps) {

  const theme = useTheme()

  const backgroundColor =
    type === 'error'
      ? theme.colors.error
      : styles.success.backgroundColor


  return (
    <Portal>
      <Snackbar
        visible={true}
        duration={duration}
        onDismiss={onDismiss}
        // style={[styles.snackbar, styles[type]]}
              style={{ backgroundColor }}
                    theme={{ colors: { onSurface: '#fff' } }} 
        action={{
          label: "Hide",
           onPress: onDismiss,
        labelStyle: { color: '#fff' },
        }}
      >
        {message}
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
});
