import { useState } from "react";
import { Text, TextInput } from "react-native";
import { Action, Body, Screen, s } from "../components/ui";
import { useAuth } from "../lib/auth";
export default function SignIn() {
  const { signIn } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  return (
    <Screen eyebrow="Secure operations companion" title="AI Execution Advisor">
      <Body>
        Sign in with your organization account. Mobile credentials are stored in
        the device secure store.
      </Body>
      <TextInput
        accessibilityLabel="Email"
        autoCapitalize="none"
        keyboardType="email-address"
        onChangeText={setEmail}
        style={s.input}
        value={email}
      />
      <TextInput
        accessibilityLabel="Password"
        onChangeText={setPassword}
        secureTextEntry
        style={s.input}
        value={password}
      />
      {error ? <Text style={s.error}>{error}</Text> : null}
      <Action
        disabled={pending || !email || !password}
        label={pending ? "Signing in…" : "Sign in"}
        onPress={() => {
          setPending(true);
          setError("");
          signIn(email, password)
            .catch(() =>
              setError(
                "Sign-in failed. Check your credentials and connection.",
              ),
            )
            .finally(() => setPending(false));
        }}
      />
    </Screen>
  );
}
