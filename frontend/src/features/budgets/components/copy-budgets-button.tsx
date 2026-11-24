"use client";

import { useState } from "react";
import { toast } from "sonner";
import { copyBudgetsFromPreviousMonth } from "../actions";
import { Button } from "@/components/ui/button";
import { CopyIcon } from "@phosphor-icons/react";

export function CopyBudgetsButton({ month }: { month: string }) {
  const [loading, setLoading] = useState(false);

  async function handleCopy() {
    setLoading(true);
    const result = await copyBudgetsFromPreviousMonth(month);
    setLoading(false);

    if ("error" in result) {
      toast.error(result.error);
    } else if (result.copied === 0) {
      toast.info("Tidak ada budget baru yang disalin dari bulan lalu.");
    } else {
      toast.success(`${result.copied} budget disalin dari bulan lalu.`);
    }
  }

  return (
    <Button size="sm" variant="outline" onClick={handleCopy} disabled={loading}>
      <CopyIcon className="mr-1 size-4" />
      Salin dari bulan lalu
    </Button>
  );
}
