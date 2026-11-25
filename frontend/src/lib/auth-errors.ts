// signIn() error codes that let the login page explain why a login failed.
// Kept apart from lib/auth.ts so the client-side login page can import them.

// The Go API could not be reached.
export const API_UNAVAILABLE = "api_unavailable";

// Too many failed attempts; the wait in minutes is part of the code because
// signIn() only passes a code string to the page, e.g. "too_many_attempts_12".
const TOO_MANY_ATTEMPTS = "too_many_attempts_";

// 0 minutes means the wait is unknown (the API did not send it).
export function tooManyAttemptsCode(retryAfterSeconds: number) {
  const minutes = Number.isFinite(retryAfterSeconds)
    ? Math.max(1, Math.ceil(retryAfterSeconds / 60))
    : 0;
  return TOO_MANY_ATTEMPTS + minutes;
}

// Returns the wait in minutes for a too-many-attempts code (0 = unknown), otherwise null.
export function retryAfterMinutes(code: string | undefined): number | null {
  if (!code?.startsWith(TOO_MANY_ATTEMPTS)) return null;
  const minutes = Number(code.slice(TOO_MANY_ATTEMPTS.length));
  return Number.isInteger(minutes) ? minutes : null;
}
