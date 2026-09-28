import { z } from "zod";

export const bankAccountSchema = z.object({
  name: z.string().min(1, "Nama akun harus diisi"),
  bankName: z.string().min(1, "Nama bank harus diisi"),
  accountNumber: z.string().min(1, "Nomor akun harus diisi"),
  type: z.enum(["BANK", "EWALLET"]),                                                                     
  balance: z.coerce.number().nonnegative("Saldo tidak boleh negatif"),
});

export type BankAccount = z.infer<typeof bankAccountSchema>;
