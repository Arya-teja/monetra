"use server";

import { auth } from "@/lib/auth";
import { prisma } from "@/lib/prisma";
import { categorySchema } from "@/features/categories/schemas";
import { revalidatePath } from "next/cache";

export async function getCategories() {
  const session = await auth();
  if (!session?.user?.id) {
    return [];
  }

  return prisma.category.findMany({
    where: {
      userId: session.user.id,
    },
    orderBy: [{ type: "asc" }, { name: "asc" }],
  });
}

export async function createCategory(data: FormData) {
  const session = await auth();
  if (!session?.user?.id) {
    throw new Error("Kamu belum login. Silakan login terlebih dahulu.");
  }

  // Ambil data mentah dari FormData
  const raw = {
    name: data.get("name"),
    type: data.get("type"),
  };

  const parsed = categorySchema.safeParse(raw);

  if (!parsed.success) {
    return {
      error: parsed.error.issues[0].message,
    };
  }

  // Cek apakah kategori dengan nama dan tipe yang sama sudah ada untuk user ini
  const duplicateCategory = await prisma.category.findFirst({
    where: {
      userId: session.user.id,
      type: parsed.data.type,
      name: { equals: parsed.data.name, mode: "insensitive" },
    },
  });

  if (duplicateCategory) {
    return {
      error: "Kategori dengan nama dan tipe yang sama sudah ada.",
    };
  }

  // Simpan data ke database menggunakan Prisma
  await prisma.category.create({
    data: {
      ...parsed.data,
      userId: session.user.id,
    },
  });

  // Revalidasi path untuk memperbarui cache
  revalidatePath("/categories");

  return { success: true };
}

export async function updateCategory(id: string, data: FormData) {
  const session = await auth();
  if (!session?.user?.id) {
    throw new Error("Kamu belum login. Silakan login terlebih dahulu.");
  }

  // Cek apakah kategori dengan ID yang diberikan sudah add atau belom
  const existingCategory = await prisma.category.findUnique({
    where: {
      id,
    },
  });

  if (!existingCategory || existingCategory.userId !== session.user.id) {
    return { error: "Kategori tidak ditemukan" };
  }

  // Ambil data mentah dari FormData
  const raw = {
    name: data.get("name"),
    type: data.get("type"),
  };

  // Validasi data menggunakan Zod
  const parsed = categorySchema.safeParse(raw);

  if (!parsed.success) {
    return {
      error: parsed.error.issues[0].message,
    };
  }

  await prisma.category.update({
    where: {
      id,
    },
    data: {
      ...parsed.data,
    },
  });

  // Revalidasi path untuk memperbarui cache
  revalidatePath("/categories");

  return { success: true };
}

export async function deleteCategory(id: string) {
  const session = await auth();
    if (!session?.user?.id) {
    throw new Error("Kamu belum login. Silakan login terlebih dahulu.");
    }

    // Cek apakah kategori dengan ID yang diberikan sudah ada atau belom
    const existingCategory = await prisma.category.findUnique({
        where: {
            id,
        },
    });

    if (!existingCategory || existingCategory.userId !== session.user.id) {
        return { error: "Kategori tidak ditemukan" };
    }

    await prisma.category.delete({
        where: {
            id,
        },
    });

    // Revalidasi path untuk memperbarui cache
    revalidatePath("/categories");

    return { success: true };
}