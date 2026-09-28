import { getBankAccount } from "@/features/bank-accounts/actions";
import { AddAccountDialog } from "@/features/bank-accounts/components/add-account-dialog";
import { AccountRowActions } from "@/features/bank-accounts/components/account-row-actions";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

function formatRupiah(amount: number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
}

function maskAccountNumber(accountNumber: string) {
  if (accountNumber.length <= 4) return accountNumber;
  return "•••• " + accountNumber.slice(-4);
}

export default async function AccountsPage() {
  const accounts = await getBankAccount();

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading text-2xl font-semibold">Akun Bank</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Kelola akun bank, e-wallet, dan cash kamu di sini.
          </p>
        </div>
        <AddAccountDialog />
      </div>

      <div className="mt-6 rounded-lg border">
        {accounts.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-16 text-center">
            <p className="text-sm text-muted-foreground">
              Kamu belum punya akun bank.
            </p>
            <p className="text-xs text-muted-foreground">
              Tambahkan akun pertama kamu untuk mulai mencatat transaksi.
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nama Akun</TableHead>
                <TableHead>Bank</TableHead>
                <TableHead>Nomor Akun</TableHead>
                <TableHead>Tipe</TableHead>
                <TableHead className="text-right">Saldo</TableHead>
                <TableHead className="w-10"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {accounts.map((account) => (
                <TableRow key={account.id}>
                  <TableCell className="font-medium">{account.name}</TableCell>
                  <TableCell>{account.bankName}</TableCell>
                  <TableCell className="font-mono text-sm">
                    {maskAccountNumber(account.accountNumber)}
                  </TableCell>
                  <TableCell>
                    {account.type === "BANK" ? "Bank" : "E-Wallet"}
                  </TableCell>
                  <TableCell className="text-right font-mono">
                    {formatRupiah(account.balance)}
                  </TableCell>
                  <TableCell>
                    <AccountRowActions account={account} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  );
}
