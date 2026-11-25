"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { signIn } from "next-auth/react";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { API_UNAVAILABLE } from "@/lib/auth-errors";

export default function LoginPage() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(formData: FormData) {
    setLoading(true);
    setError(null);
    const email = formData.get("email") as string;
    const password = formData.get("password") as string;

    // panggil fungsi signIn dari next-auth untuk melakukan login
    const result = await signIn("credentials", {
      email,
      password,
      redirect: false,
    });
    if (result?.error) {
      setError(
        result.code === API_UNAVAILABLE
          ? "Server tidak dapat dihubungi. Coba lagi beberapa saat lagi."
          : "Email atau password salah",
      );
      setLoading(false);
      return;
    }
    router.push("/dashboard");
    router.refresh();
  }

  return (
<main className="min-h-screen bg-background px-4 py-5 text-foreground sm:px-6 lg:px-10 lg:py-8">
      <div className="mx-auto grid min-h-[calc(100vh-2.5rem)] max-w-6xl overflow-hidden rounded-2xl border bg-background shadow-lg lg:grid-cols-[1.05fr_0.95fr]">

      {/* Bagian kiri halaman login */}
        <section className="relative flex flex-col justify-between overflow-hidden bg-primary p-7 text-primary-foreground lg:min-h-105 sm:p-10 lg:p-12">
          <div className="absolute -right-24 -top-24 size-72 rounded-full border border-primary-foreground/20" />
          <div className="absolute -right-8 top-8 size-40 rounded-full border border-primary-foreground/15" />

          <div className="relative z-10">
            <div className="flex items-center gap-3">
              <div className="grid size-9 place-items-center rounded-lg bg-background text-sm font-black text-primary">
                M
              </div>
              <span className="font-heading text-lg font-semibold tracking-tight">
                monetra
              </span>
            </div>
            <div className="mt-8 max-w-md lg:mt-24">
              <p className="mb-4 text-xs font-semibold uppercase tracking-[0.22em] text-primary-foreground">
                Your money, in focus
              </p>
              <h1 className="font-heading text-3xl font-semibold leading-[1.05] tracking-tight lg:text-5xl">
                Make every rupiah tell a better story.
              </h1>
              <p className="mt-5 max-w-sm text-sm leading-6 text-primary-foreground/90">
                Satu tempat untuk memahami kebiasaan belanja, merencanakan
                tujuan, dan menjaga uang tetap bergerak ke arah yang benar.
              </p>
            </div>
          </div>
        </section>

        {/* Bagian kanan halaman login */}
        <section className="flex items-center bg-background px-6 py-10 sm:px-12 lg:px-16">
          <Card className="w-full border-0 bg-transparent py-0 ring-0">
            <CardHeader className="px-0 pb-7">
              <p className="mb-3 text-xs font-semibold uppercase tracking-[0.18em] text-primary">
                Start with clarity
              </p>
              <CardTitle className="font-heading text-3xl font-semibold tracking-tight sm:text-4xl">
                Buat ruang untuk tujuanmu.
              </CardTitle>
              <CardDescription className="mt-2 max-w-sm text-sm leading-6">
                Login dan lihat uangmu dengan cara yang lebih sederhana.
              </CardDescription>
            </CardHeader>
            <CardContent className="px-0">
              <form action={handleSubmit} className="flex flex-col gap-5">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    type="email"
                    name="email"
                    id="email"
                    placeholder="kamu@email.com"
                    required
                    className="h-11 bg-background"
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    type="password"
                    name="password"
                    id="password"
                    placeholder="Minimal 8 karakter"
                    required
                    className="h-11 bg-background"
                  />
                </div>

                {error && (
                  <p className="border border-destructive/20 bg-destructive/10 px-3 py-2 text-sm text-destructive">
                    {error}
                  </p>
                )}

                <Button
                  type="submit"
                  disabled={loading}
                  className="mt-1 h-11 w-full text-sm"
                >
                  {loading ? "Loading..." : "Login"}
                </Button>
              </form>
            </CardContent>
            <CardFooter className="mt-7 justify-center border-t px-0 pt-6 gap-x-0.5 text-sm text-muted-foreground">
              Belum punya akun?{" "}
              <Link
                href="/register"
                className="font-semibold text-primary hover:underline"
              >
                Daftar di sini
              </Link>
            </CardFooter>
          </Card>
        </section>
      </div>
    </main>

  )
}
