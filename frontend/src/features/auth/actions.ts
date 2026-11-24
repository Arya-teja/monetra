"use server";

import { apiRequest } from "@/lib/api";
import { registerSchema } from "@/features/auth/schemas";

export async function registerUser(data: FormData) {
  const parsed = registerSchema.safeParse({
    name: data.get("name"),
    email: data.get("email"),
    password: data.get("password"),
  });

  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  const result = await apiRequest("/auth/register", {
    method: "POST",
    body: parsed.data,
  });

  if (!result.ok) {
    return {
      error:
        result.error.code === "CONFLICT"
          ? "Email sudah digunakan"
          : "Registrasi gagal. Silakan coba lagi.",
    };
  }

  return { success: true };
}
