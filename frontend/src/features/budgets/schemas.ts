import { z } from "zod";

export const budgetAmountSchema = z.coerce
  .number()
  .int("Nominal budget harus bilangan bulat")
  .positive("Nominal budget harus lebih dari 0");

export const createBudgetSchema = z.object({
  categoryId: z.string().min(1, "Kategori harus dipilih"),
  month: z.string().regex(/^\d{4}-\d{2}$/, "Bulan tidak valid"),
  amount: budgetAmountSchema,
});
