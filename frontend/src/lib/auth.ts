import NextAuth, { CredentialsSignin } from "next-auth";
import Credentials from "next-auth/providers/credentials";
import { apiRequest } from "@/lib/api";
import { API_UNAVAILABLE } from "@/lib/auth-errors";

// Lets the login page tell "wrong password" apart from "the Go API could not be reached".
class ApiUnavailableError extends CredentialsSignin {
  code = API_UNAVAILABLE;
}

type LoginResponse = {
  accessToken: string;
  expiresAt: string;
  user: { id: string; name: string; email: string };
};

export const { handlers, signOut, auth } = NextAuth({
  session: {
    strategy: "jwt",
    maxAge: 60 * 60 * 24 * 7, // 7 days, same as the Go access token
    updateAge: 60 * 60 * 24, // 24 hours
  },
  pages: {
    signIn: "/login",
  },

  providers: [
    Credentials({
      name: "Credentials",
      credentials: {
        email: { label: "Email", type: "text" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        const result = await apiRequest<LoginResponse>("/auth/login", {
          method: "POST",
          body: { email: credentials?.email, password: credentials?.password },
        });
        if (!result.ok) {
          // Wrong credentials (401) or empty fields (400): a normal failed login.
          if (result.error.code === "UNAUTHORIZED" || result.error.code === "VALIDATION_ERROR") {
            return null;
          }
          throw new ApiUnavailableError();
        }

        const { accessToken, expiresAt, user } = result.data;
        return {
          id: user.id,
          name: user.name,
          email: user.email,
          accessToken,
          accessTokenExpiresAt: expiresAt,
        };
      },
    }),
  ],

  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.id = user.id as string;
        token.accessToken = user.accessToken;
        token.accessTokenExpires = Date.parse(user.accessTokenExpiresAt);
      }
      // A session is only valid with an unexpired Go token. Ending it here (instead of
      // in each page) keeps proxy.ts and the pages in agreement: sessions created before
      // the Go API existed, or whose token expired, are dropped everywhere at once.
      // Otherwise proxy.ts would treat them as logged in while pages send them to /login.
      if (
        !token.accessToken ||
        !Number.isFinite(token.accessTokenExpires) ||
        Date.now() >= token.accessTokenExpires
      ) {
        return null;
      }
      return token;
    },
    async session({ session, token }) {
      session.user.id = token.id;
      session.accessToken = token.accessToken;
      return session;
    },
  },
});
