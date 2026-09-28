"use server";

import { auth } from "@/lib/auth";
import { prisma } from "@/lib/prisma";
import { bankAccountSchema } from "@/features/bank-accounts/schemas";
import { revalidatePath } from "next/cache";
import type { BankAccountView } from "@/features/bank-accounts/types/types";

export async function getBankAccount(): Promise<BankAccountView[]> {
  const session = await auth();
  if (!session?.user?.id) {
    return [];
  }

  // Ambil akun bank dari database menggunakan Prisma
  const accounts = await prisma.bankAccount.findMany({
    where: {
      userId: session.user.id,
    },
  });

  // Konversi balance dari Prisma Decimal ke number
  return accounts.map((account) => ({
    ...account,
    balance: account.balance.toNumber(),
  }));
}

export async function createBankAccount(data: FormData) {
  const session = await auth();
  if (!session?.user?.id) {
    throw new Error("Kamu belum login. Silakan login terlebih dahulu.");
  }

  // Ambil data mentah dari FormData
  const raw = {
    name: data.get("name"),
    bankName: data.get("bankName"),
    accountNumber: data.get("accountNumber"),
    type: data.get("type"),
    balance: data.get("balance"),
  };

  // Validasi data menggunakan Zod
  const parsed = bankAccountSchema.safeParse(raw);

  if (!parsed.success) {
    return {
      error: parsed.error.issues[0].message,
    };
  }

  // Simpan data ke database menggunakan Prisma
  await prisma.bankAccount.create({
    data: {
      ...parsed.data,
      userId: session.user.id,
    },
  });

  revalidatePath("/accounts");
  return { success: true };
}

export async function updateBankAccount(id: string, data: FormData) {
  const session = await auth();
  if (!session?.user?.id) {
    throw new Error("Kamu belum login. Silakan login terlebih dahulu.");
  }

  const existingAccount = await prisma.bankAccount.findUnique({
    where: {
      id,
    },
  });

  if (!existingAccount || existingAccount.userId !== session.user.id) {
    return { error: "Akun tidak ditemukan " };
  }

  const raw = {
    name: data.get("name"),
    bankName: data.get("bankName"),
    accountNumber: data.get("accountNumber"),
    type: data.get("type"),
    balance: data.get("balance"),
  };

  const parsed = bankAccountSchema.safeParse(raw);

  if (!parsed.success) {
    return { error: parsed.error.issues[0].message };
  }

  await prisma.bankAccount.update({
    where: { id },
    data: {
      ...parsed.data,
    },
  });
  revalidatePath("/accounts");
  return { success: true };
}

export async function deleteBankAccount(id: string) {
  const session = await auth();
  if (!session?.user?.id) {
    throw new Error("Kamu belum login. Silakan login terlebih dahulu.");
  }

  const existingAccount = await prisma.bankAccount.findUnique({
    where: {
      id,
    },
  });

  if (!existingAccount || existingAccount.userId !== session.user.id) {
    return { error: "Akun tidak ditemukan " };
  }

  await prisma.bankAccount.delete({
    where: { id },
  });
  revalidatePath("/accounts");
  return { success: true };
}
