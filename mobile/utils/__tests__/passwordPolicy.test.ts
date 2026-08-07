import {
  isPasswordTooLong,
  isPasswordTooShort,
  passwordCodePointLength,
  passwordUtf8ByteLength,
} from "../passwordPolicy";

describe("password policy", () => {
  it("counts Unicode code points for the minimum", () => {
    expect(passwordCodePointLength("😀😀😀😀")).toBe(4);
    expect(isPasswordTooShort("😀😀😀😀")).toBe(true);
    expect(isPasswordTooShort("😀😀😀😀😀😀😀😀")).toBe(false);
  });

  it("counts UTF-8 bytes for the bcrypt ceiling", () => {
    expect(passwordUtf8ByteLength("界".repeat(24))).toBe(72);
    expect(isPasswordTooLong("界".repeat(24))).toBe(false);
    expect(isPasswordTooLong("界".repeat(25))).toBe(true);
  });
});
