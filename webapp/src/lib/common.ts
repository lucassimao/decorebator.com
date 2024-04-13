type ApplicationErrorCode =
  | "AuthenticationError"
  | "SignUpError"
  | "ValidationError";

export class ApplicationError extends Error {
  readonly code?: ApplicationErrorCode;

  constructor(message: string, code?: ApplicationErrorCode) {
    super(message);
    this.code = code;
  }
}

export class ValidationError extends Error {
  // readonly validationErrors: Record<string, string>;

  constructor(readonly validationErrors: Record<string, string>) {
    super("Validation error");
  }
}
