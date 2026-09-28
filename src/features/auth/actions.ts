"use server";

import bcrypt from "bcryptjs";
import { prisma } from "@/lib/prisma";
import { z } from "zod";

const registerSchema = z.object({
  name: z.string().min(2, "Nama minimal 2 karakter"),
  email: z.email("Email tidak valid"),
  password: z.string().min(6, "Password minimal 6 karakter"),
});

export async function registerUser(data: FormData) {
  const rawData = {
    name: data.get("name"),
    email: data.get("email"),
    password: data.get("password"),
  };

  //cek apakah data registeruser valid sesuai kebutuhan registerSchema
  const parsed = registerSchema.safeParse(rawData);

  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  //cek apakah email sudah digunakan
  const { name, email, password } = parsed.data;

  const existingUser = await prisma.user.findUnique({
    where: { email },
  });
  if (existingUser) {
    return { error: "Email sudah digunakan" };
  }

  //cek apakah password sudah di hash
  const hashedPassword = await bcrypt.hash(password, 10);

  await prisma.user.create({
    data: {
      name,
      email,
      password: hashedPassword,
    },
  });

  return { success: true };
}
