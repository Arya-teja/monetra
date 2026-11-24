import { redirect } from "next/navigation";
import { auth } from "@/lib/auth";

// Returns the Go API token of the logged-in user, or sends them to the login page.
export async function getAccessToken(): Promise<string> {
  const session = await auth();
  if (!session?.accessToken) {
    redirect("/login");
  }
  return session.accessToken;
}
