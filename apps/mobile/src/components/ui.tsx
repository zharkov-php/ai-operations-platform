import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { Link } from "expo-router";
export function Screen({
  title,
  eyebrow,
  children,
}: {
  title: string;
  eyebrow: string;
  children: React.ReactNode;
}) {
  return (
    <ScrollView contentContainerStyle={s.screen}>
      <Text style={s.eyebrow}>{eyebrow}</Text>
      <Text accessibilityRole="header" style={s.title}>
        {title}
      </Text>
      {children}
    </ScrollView>
  );
}
export function Card({ children }: { children: React.ReactNode }) {
  return <View style={s.card}>{children}</View>;
}
export function Body({ children }: { children: React.ReactNode }) {
  return <Text style={s.body}>{children}</Text>;
}
export function Action({
  label,
  onPress,
  disabled = false,
  danger = false,
}: {
  label: string;
  onPress(): void;
  disabled?: boolean;
  danger?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      disabled={disabled}
      onPress={onPress}
      style={[s.button, danger && s.danger, disabled && s.disabled]}
    >
      <Text style={s.buttonText}>{label}</Text>
    </Pressable>
  );
}
export function NavLink({
  href,
  children,
}: {
  href: any;
  children: React.ReactNode;
}) {
  return (
    <Link href={href} style={s.link}>
      {children}
    </Link>
  );
}
export function State({
  loading,
  error,
  empty,
}: {
  loading: boolean;
  error?: unknown;
  empty?: string;
}) {
  if (loading)
    return <ActivityIndicator accessibilityLabel="Loading" color="#65e6bd" />;
  if (error)
    return (
      <Card>
        <Text style={s.error}>
          Unable to load. Check your connection and try again.
        </Text>
      </Card>
    );
  if (empty)
    return (
      <Card>
        <Body>{empty}</Body>
      </Card>
    );
  return null;
}
export const s = StyleSheet.create({
  screen: { backgroundColor: "#07111f", flexGrow: 1, padding: 24, gap: 16 },
  eyebrow: { color: "#65e6bd", fontWeight: "800", textTransform: "uppercase" },
  title: { color: "#f5f8fc", fontSize: 34, fontWeight: "800" },
  card: {
    backgroundColor: "#0d1b2e",
    borderColor: "#29425f",
    borderRadius: 16,
    borderWidth: 1,
    gap: 10,
    padding: 18,
  },
  body: { color: "#abc0d8", fontSize: 16, lineHeight: 23 },
  button: { backgroundColor: "#65e6bd", borderRadius: 24, padding: 13 },
  danger: { backgroundColor: "#c7686f" },
  disabled: { opacity: 0.45 },
  buttonText: { color: "#07111f", fontWeight: "800", textAlign: "center" },
  link: {
    color: "#65e6bd",
    fontSize: 17,
    fontWeight: "700",
    paddingVertical: 8,
  },
  error: { color: "#ffb5b5", fontSize: 16 },
  input: {
    backgroundColor: "#0d1b2e",
    borderColor: "#47647f",
    borderRadius: 10,
    borderWidth: 1,
    color: "#f5f8fc",
    padding: 13,
  },
});
