import { z } from "zod";
import { cashflowTypes } from "@/features/categories/types";

export const categoryNameSchema = z
  .string()
  .trim()
  .min(1, "Nama kategori harus diisi")
  .max(50, "Nama kategori maksimal 50 karakter");

export const createCategorySchema = z.object({
  name: categoryNameSchema,
  type: z.enum(cashflowTypes, "Tipe kategori tidak valid"),
});
