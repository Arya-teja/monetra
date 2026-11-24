import { z } from "zod";
import { cashflowTypes } from "@/features/categories/types";

export const transactionSchema = z.object({
  type: z.enum(cashflowTypes, "Pilih pemasukan atau pengeluaran"),
  accountId: z.string().min(1, "Akun harus dipilih"),
  categoryId: z.string().min(1, "Kategori harus dipilih"),
  amount: z.coerce
    .number()
    .int("Nominal harus bilangan bulat")
    .positive("Nominal harus lebih dari 0"),
  occurredAt: z.iso.date("Tanggal tidak valid"),
  note: z.string().trim().max(200, "Catatan maksimal 200 karakter"),
});
