import { prepareProfileImage } from "./index.web";

describe("profile image processor web fallback", () => {
  it("fails closed before attempting an unbounded browser decode", async () => {
    await expect(prepareProfileImage()).rejects.toThrow(
      "Profile image uploads are available in the iOS and Android apps.",
    );
  });
});
