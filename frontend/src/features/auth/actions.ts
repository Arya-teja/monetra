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
    switch (result.error.code) {
      case "CONFLICT":
        return { error: "Email sudah digunakan" };
      case "VALIDATION_ERROR":
        return { error: "Data registrasi tidak valid. Periksa kembali isiannya." };
      default:
        return { error: "Server tidak dapat dihubungi. Coba lagi beberapa saat lagi." };
    }
  }

  return { success: true };
}
