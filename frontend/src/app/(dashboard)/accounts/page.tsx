import { getAccounts } from "@/features/accounts/actions";
import { AddAccountDialog } from "@/features/accounts/components/add-account-dialog";
import { AccountRowActions } from "@/features/accounts/components/account-row-actions";
import { accountTypeLabels } from "@/features/accounts/types";
import { formatRupiah } from "@/lib/format";
import { cn } from "@/lib/utils";

function maskAccountNumber(accountNumber: string) {
  if (accountNumber.length <= 4) return accountNumber;
  return "•••• " + accountNumber.slice(-4);
}

export default async function AccountsPage() {
  const result = await getAccounts();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="font-heading text-2xl font-semibold">Akun</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Kelola akun bank, e-wallet, dan cash kamu di sini.
          </p>
        </div>
        <AddAccountDialog />
      </div>

      {!result.ok ? (
        <p className="rounded-lg border px-4 py-16 text-center text-sm text-destructive">
          Gagal memuat akun. Silakan muat ulang halaman.
        </p>
      ) : result.data.length === 0 ? (
        <div className="rounded-lg border px-4 py-16 text-center">
          <p className="text-sm text-muted-foreground">Kamu belum punya akun.</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Tambahkan akun pertama kamu untuk mulai mencatat transaksi.
          </p>
        </div>
      ) : (
        <ul className="divide-y rounded-lg border">
          {result.data.map((account) => (
            <li
              key={account.id}
              className={cn(
                "flex items-center gap-3 px-4 py-3",
                account.isArchived && "text-muted-foreground",
              )}
            >
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">
                  {account.name}
                  {account.isArchived && (
                    <span className="ml-2 rounded-full bg-muted px-2 py-0.5 text-xs font-normal">
                      Diarsipkan
                    </span>
                  )}
                </p>
                <p className="truncate text-sm text-muted-foreground">
                  {[
                    accountTypeLabels[account.type],
                    account.institutionName,
                    account.accountNumber && maskAccountNumber(account.accountNumber),
                  ]
                    .filter(Boolean)
                    .join(" · ")}
                </p>
              </div>
              <p
                className={cn(
                  "font-semibold tabular-nums",
                  account.balance < 0 && "text-expense",
                )}
              >
                {formatRupiah(account.balance)}
                {account.balance < 0 && <span className="sr-only"> (saldo minus)</span>}
              </p>
              <AccountRowActions account={account} />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
