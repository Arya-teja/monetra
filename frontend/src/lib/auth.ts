import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import { apiRequest } from "@/lib/api";

type LoginResponse = {
  accessToken: string;
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

        const { accessToken, user } = result.data;
        return { id: user.id, name: user.name, email: user.email, accessToken };
      },
    }),
  ],

  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.id = user.id as string;
        token.accessToken = user.accessToken;
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
