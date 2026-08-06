import { render, screen } from "@testing-library/react-native";
import Overview from "./index";

describe("overview foundation", () => {
  it("renders the product name", () => {
    render(<Overview />);
    expect(screen.getByText("AI Execution Advisor")).toBeTruthy();
  });
});
