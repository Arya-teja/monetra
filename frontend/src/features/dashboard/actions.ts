"use server";

import { apiRequest, type ApiResult } from "@/lib/api";
import { getAccessToken } from "@/lib/session";
import type { DashboardSummary, MonthTotal } from "@/features/dashboard/types";

// Summary of the current month (Asia/Jakarta, decided by the API).
export async function getDashboardSummary(): Promise<ApiResult<DashboardSummary>> {
  return apiRequest<DashboardSummary>("/dashboard/summary", {
    method: "GET",
    token: await getAccessToken(),
  });
}

// Income and expense of the last 6 months, oldest first.
export async function getTrend(): Promise<ApiResult<MonthTotal[]>> {
  return apiRequest<MonthTotal[]>("/dashboard/trend?months=6", {
    method: "GET",
    token: await getAccessToken(),
  });
}
