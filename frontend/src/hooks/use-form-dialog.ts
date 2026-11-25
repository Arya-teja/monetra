"use client";

import { useState, type SubmitEvent } from "react";

type SubmitResult = { error: string } | { success: true };

// State of a dialog holding one form: open/closed, the submit error, and whether a
// submit is running. The dialog closes when submit succeeds; closing it clears the error.
export function useFormDialog(submit: (formData: FormData) => Promise<SubmitResult>) {
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  function onOpenChange(value: boolean) {
    setOpen(value);
    if (!value) setError(null);
  }

  // Used as onSubmit, not a form action: React resets a form after its action runs,
  // which would clear what the user typed whenever the submit fails.
  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    setLoading(true);
    setError(null);
    const result = await submit(formData);
    setLoading(false);
    if ("error" in result) {
      setError(result.error);
      return;
    }
    setOpen(false);
  }

  return { open, onOpenChange, error, loading, handleSubmit };
}
