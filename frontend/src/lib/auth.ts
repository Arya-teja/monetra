import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import { apiRequest } from "@/lib/api";

type LoginResponse = {
  accessToken: string;
  expiresAt: string;
  user: { id: string; name: string; email: string };
};

export const { handlers, signIn, signOut, auth } = NextAuth({
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
          return null;
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
      // End the session together with the Go token, so a rolling session
      // never outlives the token it depends on.
      if (Date.now() >= token.accessTokenExpires) {
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
