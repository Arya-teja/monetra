"use client";

import { useState, type SubmitEvent } from "react";
import { useRouter } from "next/navigation";
import { registerUser } from "@/features/auth/actions";
import { AuthCard } from "@/features/auth/components/auth-card";
import { Button } from "@/components/ui/button";
import { FormError } from "@/components/form-error";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function RegisterPage() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // onSubmit, not a form action: React resets a form after its action runs, which
  // would clear what the user typed whenever the submit fails.
  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    setLoading(true);
    setError(null);
    const result = await registerUser(formData);
    if (result.error) {
      setError(result.error);
      setLoading(false);
      return;
    }
    router.push("/login");
    router.refresh();
  }

  return (
    <AuthCard
      description="Daftar dan mulai lihat uangmu dengan cara yang lebih sederhana."
      footer={{ text: "Sudah punya akun?", linkLabel: "Masuk di sini", href: "/login" }}
    >
      <form onSubmit={handleSubmit} className="flex flex-col gap-5">
        <div className="flex flex-col gap-2">
          <Label htmlFor="name">Nama</Label>
          <Input
            type="text"
            name="name"
            id="name"
            placeholder="Nama lengkap"
            required
            className="h-11 bg-background"
          />
        </div>
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
          {loading ? "Loading..." : "Register"}
        </Button>
      </form>
    </AuthCard>
  );
}
