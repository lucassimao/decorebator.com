export interface BurstErrorData {
  error: string;
  error_code: "BURST_WARNING" | "ACCOUNT_SUSPENDED_BURST";
  suspended_until?: string;
  retry_after?: number;
}

export class BurstViolationError extends Error {
  public errorCode: "BURST_WARNING" | "ACCOUNT_SUSPENDED_BURST";
  public suspendedUntil?: Date;
  public retryAfter?: number;

  constructor(data: BurstErrorData) {
    super(data.error);
    this.name = "BurstViolationError";
    this.errorCode = data.error_code;
    
    if (data.suspended_until) {
      this.suspendedUntil = new Date(data.suspended_until);
    }
    
    this.retryAfter = data.retry_after;
  }
  
  isWarning(): boolean {
    return this.errorCode === "BURST_WARNING";
  }
  
  isSuspended(): boolean {
    return this.errorCode === "ACCOUNT_SUSPENDED_BURST";
  }
}