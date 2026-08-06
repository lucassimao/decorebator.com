import { getApiBaseUrl } from "../baseUrl";

let mockChannel: string | null = null;
let mockReleaseChannel: string | null = null;

jest.mock("expo-updates", () => ({
  get channel() {
    return mockChannel;
  },
  get releaseChannel() {
    return mockReleaseChannel;
  },
}));

describe("getApiBaseUrl", () => {
  const originalApiUrl = process.env.EXPO_PUBLIC_API_URL;

  beforeEach(() => {
    mockChannel = null;
    mockReleaseChannel = null;
    delete process.env.EXPO_PUBLIC_API_URL;
  });

  afterAll(() => {
    if (originalApiUrl === undefined) {
      delete process.env.EXPO_PUBLIC_API_URL;
    } else {
      process.env.EXPO_PUBLIC_API_URL = originalApiUrl;
    }
  });

  it("uses the production API for a production channel without an override", () => {
    mockChannel = "production";

    expect(getApiBaseUrl()).toBe("https://api.decorebator.com");
  });

  it("prevents a production channel from using a local URL", () => {
    mockChannel = "production";
    process.env.EXPO_PUBLIC_API_URL = "http://localhost:3000";
    const consoleError = jest.spyOn(console, "error").mockImplementation();

    expect(getApiBaseUrl()).toBe("https://api.decorebator.com");
    expect(consoleError).toHaveBeenCalledTimes(1);

    consoleError.mockRestore();
  });

  it("does not mistake a remote URL containing local text for a local host", () => {
    mockChannel = "production";
    process.env.EXPO_PUBLIC_API_URL =
      "https://staging.example.com/api?redirect=localhost";

    expect(getApiBaseUrl()).toBe(
      "https://staging.example.com/api?redirect=localhost",
    );
  });

  it.each([
    [undefined, "must define EXPO_PUBLIC_API_URL"],
    ["https://api.decorebator.com", "must not use the production API"],
    ["https://api.decorebator.com.", "must not use the production API"],
    ["http://sandbox.example.com", "must use an HTTPS API URL"],
    ["http://10.0.2.2:3000", "must use an HTTPS API URL"],
    ["https://localhost.", "must use an HTTPS API URL"],
    ["https://[::1]", "must use an HTTPS API URL"],
    ["not-a-url", "must be an absolute URL"],
  ])("fails closed for an unsafe store-test API URL: %s", (value, message) => {
    mockChannel = "store-test";
    if (value !== undefined) process.env.EXPO_PUBLIC_API_URL = value;

    expect(() => getApiBaseUrl()).toThrow(message);
  });

  it("accepts an explicit nonproduction HTTPS API on store-test", () => {
    mockChannel = "store-test";
    process.env.EXPO_PUBLIC_API_URL = "https://iap-sandbox.decorebator.com";

    expect(getApiBaseUrl()).toBe("https://iap-sandbox.decorebator.com");
  });
});
