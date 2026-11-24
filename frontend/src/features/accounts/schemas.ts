import { z } from "zod";
import { accountTypes } from "@/features/accounts/types";

const accountFields = {
  name: z
    .string()
    .trim()
    .min(1, "Nama akun harus diisi")
    .max(50, "Nama akun maksimal 50 karakter"),
  type: z.enum(accountTypes, "Tipe akun tidak valid"),
  institutionName: z
    .string()
    .trim()
    .max(50, "Nama institusi maksimal 50 karakter"),
  accountNumber: z
    .string()
    .trim()
    .max(50, "Nomor akun maksimal 50 karakter"),
};

export const createAccountSchema = z.object({
  ...accountFields,
  balance: z.coerce
    .number()
    .int("Saldo awal harus bilangan bulat")
    .nonnegative("Saldo awal tidak boleh negatif"),
});

// The balance is not editable: after creation it only changes through transactions.
export const updateAccountSchema = z.object(accountFields);
