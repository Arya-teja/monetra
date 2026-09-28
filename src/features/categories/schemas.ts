import { z } from "zod";

export const categorySchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Nama kategori harus diisi")
    .max(50, "Nama kategori harus kurang dari 50 karakter"),
  type: z.enum(["INCOME", "EXPENSE"]),
});

export type Category = z.infer<typeof categorySchema>;
