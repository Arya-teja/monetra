import { NextResponse } from "next/server";
import { auth } from "@/lib/auth";

// Halaman yang boleh diakses tanpa login. Selain ini, semua wajib login.
const publicPaths = ["/", "/login", "/register"];
const authPages = ["/login", "/register"];

export const proxy = auth((req) => {
  const isLoggedIn = !!req.auth;
  const { pathname } = req.nextUrl;

  const isPublic = publicPaths.includes(pathname);
  const isAuthPage = authPages.includes(pathname);

  // Belum login dan bukan halaman publik → redirect ke login
  if (!isLoggedIn && !isPublic) {
    return NextResponse.redirect(new URL("/login", req.url));
  }

  // Sudah login tapi buka halaman login/register → redirect ke dashboard
  if (isLoggedIn && isAuthPage) {
    return NextResponse.redirect(new URL("/dashboard", req.url));
  }

  return NextResponse.next();
});

export const config = {
  // Jalankan di semua path, kecuali API, aset Next.js, dan file statis (yang ada titiknya)
  matcher: ["/((?!api|_next/static|_next/image|.*\\..*).*)"],
};