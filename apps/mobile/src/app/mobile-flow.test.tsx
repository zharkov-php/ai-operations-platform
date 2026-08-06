import {
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react-native";
import { OverviewView } from "./(app)/index";
import { AlertCard } from "./(app)/alerts";
import { RecommendationActions } from "./(app)/recommendations/[id]";
import { ExperimentActions } from "./(app)/experiments/[id]";
import { State } from "../components/ui";
import SignIn from "./sign-in";
import { NotificationList } from "./(app)/notifications";

const mockSignIn = jest.fn();
jest.mock("../lib/auth", () => ({
  useAuth: () => ({ signIn: mockSignIn }),
}));
jest.mock("expo-router", () => {
  // Jest requires dependencies used by a mock factory to be loaded in the factory.
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const { Text } = require("react-native");
  return { Link: ({ children }: any) => <Text>{children}</Text> };
});
describe("mobile operational journey", () => {
  it("submits mobile credentials through the authentication flow", async () => {
    mockSignIn.mockResolvedValue(undefined);
    render(<SignIn />);
    fireEvent.changeText(screen.getByLabelText("Email"), "owner@example.test");
    fireEvent.changeText(screen.getByLabelText("Password"), "correct-password");
    fireEvent.press(screen.getByText("Sign in"));
    await waitFor(() =>
      expect(mockSignIn).toHaveBeenCalledWith(
        "owner@example.test",
        "correct-password",
      ),
    );
  });
  it("shows overview and navigation", () => {
    render(
      <OverviewView
        data={{
          current_month_cost: "42.10",
          projected_month_cost: "84.20",
          verified_savings: "7.00",
          currency: "USD",
        }}
      />,
    );
    expect(screen.getByText("42.10 USD")).toBeTruthy();
    expect(screen.getByText("Budget alerts")).toBeTruthy();
  });
  it("acknowledges an alert", () => {
    const fn = jest.fn();
    render(
      <AlertCard
        alert={{
          severity: "high",
          threshold_type: "percentage",
          threshold_value: "90",
          status: "open",
        }}
        onAcknowledge={fn}
      />,
    );
    fireEvent.press(screen.getByText("Acknowledge"));
    expect(fn).toHaveBeenCalled();
  });
  it("enforces recommendation permissions and accepts for testing", () => {
    const accept = jest.fn();
    const { rerender } = render(
      <RecommendationActions
        allowed={false}
        onAccept={accept}
        onReject={jest.fn()}
      />,
    );
    expect(screen.getByText(/cannot manage/)).toBeTruthy();
    rerender(
      <RecommendationActions allowed onAccept={accept} onReject={jest.fn()} />,
    );
    fireEvent.press(screen.getByText("Accept for evaluation"));
    expect(accept).toHaveBeenCalled();
  });
  it("pauses and rolls back a running experiment", () => {
    const pause = jest.fn(),
      rollback = jest.fn();
    render(
      <ExperimentActions
        allowed
        status="running"
        onPause={pause}
        onRollback={rollback}
      />,
    );
    fireEvent.press(screen.getByText("Pause experiment"));
    fireEvent.press(screen.getByText("Rollback"));
    expect(pause).toHaveBeenCalled();
    expect(rollback).toHaveBeenCalled();
  });
  it("renders an offline-safe error state", () => {
    render(<State error={new Error("offline")} loading={false} />);
    expect(screen.getByText(/Check your connection/)).toBeTruthy();
  });
  it("shows a safe in-app notification summary", () => {
    render(
      <NotificationList
        items={[
          {
            id: "one",
            title: "Experiment rolled back",
            body: "A guardrail was exceeded.",
            status: "delivered",
          },
        ]}
      />,
    );
    expect(screen.getByText("Experiment rolled back")).toBeTruthy();
    expect(screen.queryByText(/Bearer/)).toBeNull();
  });
});
