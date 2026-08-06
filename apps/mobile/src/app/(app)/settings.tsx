import { Action, Body, Card, Screen } from "../../components/ui";
import { useAuth } from "../../lib/auth";
export default function Settings() {
  const { signOut } = useAuth();
  return (
    <Screen eyebrow="Account" title="Settings">
      <Card>
        <Body>
          Access and refresh tokens are kept in the device secure store. Signing
          out removes them.
        </Body>
        <Action danger label="Sign out" onPress={() => void signOut()} />
      </Card>
    </Screen>
  );
}
