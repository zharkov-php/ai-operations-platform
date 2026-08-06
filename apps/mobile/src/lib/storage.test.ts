import * as SecureStore from "expo-secure-store";
import { secureTokenStorage } from "./storage";
jest.mock("expo-secure-store", () => ({
  getItemAsync: jest.fn(),
  setItemAsync: jest.fn(),
  deleteItemAsync: jest.fn(),
  WHEN_UNLOCKED_THIS_DEVICE_ONLY: "device",
}));
describe("secure token storage", () => {
  it("reads, writes and clears the mobile session", async () => {
    (SecureStore.getItemAsync as jest.Mock).mockResolvedValue("session");
    expect(await secureTokenStorage.get()).toBe("session");
    await secureTokenStorage.set("next");
    expect(SecureStore.setItemAsync).toHaveBeenCalledWith(
      "ai_execution_session",
      "next",
    );
    await secureTokenStorage.clear();
    expect(SecureStore.deleteItemAsync).toHaveBeenCalledWith(
      "ai_execution_session",
    );
  });
});
