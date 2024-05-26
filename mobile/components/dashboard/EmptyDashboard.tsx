import * as usersApi from "@/api/users";
import * as React from "react";
import { Image, StyleSheet, View } from "react-native";
import { Button, Text, useTheme } from "react-native-paper";
import NewWordlistDialog from "./NewWordlistDialog";

type Props = {
  title: string;
  subtitle: string;
  onWordlistCreated: () => void;
};

export default function EmptyDashboard({
  title,
  subtitle,
  onWordlistCreated,
}: Props) {
  const theme = useTheme();
  const user = usersApi.getUserInfo();
  const [visible, setVisible] = React.useState(false);

  const onDialogDismissed = (success?: boolean) => {
    setVisible(false);
    if (success) {
      onWordlistCreated();
    }
  };

  return (
    <View style={{ flex: 1, backgroundColor: "#f5f5f5" }}>
      <View style={styles.imageContainer}>
        <Image
          source={require("@/assets/images/empty-dashboard.png")}
          style={styles.image}
          resizeMode="contain"
        />
      </View>
      <View style={styles.overlay}>
        {visible && <NewWordlistDialog onDismiss={onDialogDismissed} />}

        <Text style={styles.title}>{title}</Text>

        <View style={styles.bottom}>
          <Text style={styles.subtitle}>{subtitle}</Text>

          <Button
            style={styles.button}
            labelStyle={{ fontSize: 20 }}
            rippleColor={theme.colors.inversePrimary}
            onPress={() => setVisible(true)}
            mode="contained"
          >
            Create new wordlist
          </Button>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  imageContainer: {
    ...StyleSheet.absoluteFillObject,
    top: 0,
    //  alignItems: "center",
    // justifyContent: "center",
    backgroundColor: "transparent",
  },
  image: {
    width: "100%",
    // height: "70%",
  },
  overlay: {
    ...StyleSheet.absoluteFillObject,
    alignItems: "center",
    justifyContent: "space-between",
  },
  title: {
    fontSize: 30,
    fontWeight: "bold",
    marginTop: 50,
    textAlign: "center",
  },
  bottom: {
    alignItems: "center",
    marginBottom: 100,
  },
  subtitle: {
    fontSize: 20,
    textAlign: "center",
  },
  button: {
    width: "100%",
    marginTop: 20,
    padding: 10,
  },
});
