import { StyleSheet, Text, View } from "react-native";

export default function Overview() {
  return (
    <View style={styles.container}>
      <Text style={styles.eyebrow}>Operations companion</Text>
      <Text accessibilityRole="header" style={styles.title}>AI Execution Advisor</Text>
      <Text style={styles.body}>Authentication, alerts, and recommendations will be added as their API capabilities become available.</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, justifyContent: "center", padding: 32, backgroundColor: "#07111f" },
  eyebrow: { color: "#67e8c6", fontSize: 14, fontWeight: "700", textTransform: "uppercase" },
  title: { color: "#f4f7fb", fontSize: 40, fontWeight: "700", marginVertical: 16 },
  body: { color: "#a9bad0", fontSize: 18, lineHeight: 28 },
});
