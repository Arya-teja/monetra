export function formatRupiah(amount: number) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
}

// Formats a calendar date ("YYYY-MM-DD"). It is read and printed as UTC so the day never shifts.
export function formatDate(date: string) {
  return new Intl.DateTimeFormat("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(date));
}

// Formats a month ("YYYY-MM") such as "November 2026".
export function formatMonth(month: string) {
  return new Intl.DateTimeFormat("id-ID", {
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(`${month}-01`));
}

// Formats a month ("YYYY-MM") as a short label such as "Nov".
export function formatMonthShort(month: string) {
  return new Intl.DateTimeFormat("id-ID", {
    month: "short",
    timeZone: "UTC",
  }).format(new Date(`${month}-01`));
}

// Today's date in Asia/Jakarta as "YYYY-MM-DD".
export function todayInJakarta() {
  return new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Jakarta" }).format(new Date());
}
