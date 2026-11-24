import { getAccounts } from "@/features/accounts/actions";
import { AddAccountDialog } from "@/features/accounts/components/add-account-dialog";
import { AccountRowActions } from "@/features/accounts/components/account-row-actions";
import { accountTypeLabels } from "@/features/accounts/types";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { formatRupiah } from "@/lib/format";

function maskAccountNumber(accountNumber: string) {
  if (accountNumber.length <= 4) return accountNumber;
  return "•••• " + accountNumber.slice(-4);
}

export default async function AccountsPage() {
  const result = await getAccounts();

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading text-2xl font-semibold">Akun</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Kelola akun bank, e-wallet, dan cash kamu di sini.
          </p>
        </div>
        <AddAccountDialog />
      </div>

      <div className="mt-6 rounded-lg border">
        {!result.ok ? (
          <p className="px-4 py-16 text-center text-sm text-destructive">
            Gagal memuat akun. Silakan muat ulang halaman.
          </p>
        ) : result.data.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-16 text-center">
            <p className="text-sm text-muted-foreground">
              Kamu belum punya akun.
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
                <TableHead>Institusi</TableHead>
                <TableHead>Nomor Akun</TableHead>
                <TableHead>Tipe</TableHead>
                <TableHead className="text-right">Saldo</TableHead>
                <TableHead className="w-32"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {result.data.map((account) => (
                <TableRow
                  key={account.id}
                  className={account.isArchived ? "text-muted-foreground" : undefined}
                >
                  <TableCell className="font-medium">
                    {account.name}
                    {account.isArchived && (
                      <span className="ml-2 rounded-full bg-muted px-2 py-0.5 text-xs font-normal">
                        Diarsipkan
                      </span>
                    )}
                  </TableCell>
                  <TableCell>{account.institutionName ?? "—"}</TableCell>
                  <TableCell className="font-mono text-sm">
                    {account.accountNumber
                      ? maskAccountNumber(account.accountNumber)
                      : "—"}
                  </TableCell>
                  <TableCell>{accountTypeLabels[account.type]}</TableCell>
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
