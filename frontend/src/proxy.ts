import { NextResponse } from "next/server";
import { auth } from "@/lib/auth";

// Pages reachable without logging in. Every other page requires a session.
const publicPaths = ["/", "/login", "/register"];
const authPages = ["/login", "/register"];

export const proxy = auth((req) => {
  const isLoggedIn = !!req.auth;
  const { pathname } = req.nextUrl;

  const isPublic = publicPaths.includes(pathname);
  const isAuthPage = authPages.includes(pathname);

  // Not logged in and not a public page → send to login
  if (!isLoggedIn && !isPublic) {
    return NextResponse.redirect(new URL("/login", req.url));
  }

  // Already logged in but opening login/register → send to the dashboard
  if (isLoggedIn && isAuthPage) {
    return NextResponse.redirect(new URL("/dashboard", req.url));
  }

  return NextResponse.next();
});

export const config = {
  // Run on every path except the API, Next.js assets, PWA icons (they must load
  // before login), and static files (paths containing a dot)
  matcher: ["/((?!api|_next/static|_next/image|icon|apple-icon|.*\\..*).*)"],
};