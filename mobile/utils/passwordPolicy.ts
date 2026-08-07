export const PASSWORD_MIN_CODE_POINTS = 8;
export const PASSWORD_MAX_UTF8_BYTES = 72;

export function passwordCodePointLength(password: string): number {
  return Array.from(password).length;
}

export function passwordUtf8ByteLength(password: string): number {
  return new TextEncoder().encode(password).length;
}

export function isPasswordTooShort(password: string): boolean {
  return passwordCodePointLength(password) < PASSWORD_MIN_CODE_POINTS;
}

export function isPasswordTooLong(password: string): boolean {
  return passwordUtf8ByteLength(password) > PASSWORD_MAX_UTF8_BYTES;
}
