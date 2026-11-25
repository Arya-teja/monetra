"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

// Rupiah amounts are BIGINT; 15 digits stay within JavaScript's safe integer range.
const MAX_DIGITS = 15;

type Props = Omit<
  React.ComponentProps<typeof Input>,
  "type" | "name" | "value" | "defaultValue" | "onChange"
> & {
  /** Form field name; the submitted value is plain digits, e.g. "1500000". */
  name: string;
  defaultValue?: number;
};

// A money input that shows thousand separators while typing ("1.500.000") and
// submits the plain number through a hidden field.
export function RupiahInput({ name, defaultValue, className, ...props }: Props) {
  const [digits, setDigits] = useState(defaultValue ? String(defaultValue) : "");

  return (
    <div className="relative">
      <span
        aria-hidden="true"
        className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted-foreground"
      >
        Rp
      </span>
      <Input
        {...props}
        type="text"
        inputMode="numeric"
        autoComplete="off"
        value={digits === "" ? "" : new Intl.NumberFormat("id-ID").format(Number(digits))}
        onChange={(e) =>
          setDigits(
            e.target.value
              .replace(/\D/g, "")
              .replace(/^0+(?=\d)/, "")
              .slice(0, MAX_DIGITS),
          )
        }
        className={cn("pl-10 tabular-nums", className)}
      />
      <input type="hidden" name={name} value={digits} />
    </div>
  );
}
