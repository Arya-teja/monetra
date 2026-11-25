"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { signIn } from "next-auth/react";
import { AuthCard } from "@/features/auth/components/auth-card";
import { Button } from "@/components/ui/button";
import { FormError } from "@/components/form-error";
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
    const result = await signIn("credentials", {
      email: formData.get("email") as string,
      password: formData.get("password") as string,
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
    <AuthCard
      description="Login dan lihat uangmu dengan cara yang lebih sederhana."
      footer={{ text: "Belum punya akun?", linkLabel: "Daftar di sini", href: "/register" }}
    >
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

        <FormError message={error} />

        <Button type="submit" disabled={loading} className="mt-1 h-11 w-full text-sm">
          {loading ? "Loading..." : "Login"}
        </Button>
      </form>
    </AuthCard>
  );
}
