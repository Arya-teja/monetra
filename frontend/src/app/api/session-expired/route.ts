import { signOut } from "@/lib/auth";

// Reached when the Go API rejects the token of a logged-in session (see lib/api.ts):
// drop the session cookie so proxy.ts and the pages agree the user is logged out.
export async function GET() {
  await signOut({ redirectTo: "/login" });
}
