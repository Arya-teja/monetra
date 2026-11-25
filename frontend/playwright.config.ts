import { defineConfig, devices } from "@playwright/test";
import { API_PORT, E2E_API_URL, E2E_DATABASE_URL, E2E_JWT_SECRET, WEB_PORT } from "./e2e/env";

// End-to-end tests run against their own API (port 8090) and database (monetra_e2e),
// so the dev servers on 3000/8080 and the real database are never touched.
// Requires `docker compose up -d` (for PostgreSQL) and Go on the PATH.
export default defineConfig({
  testDir: "./e2e",
  globalSetup: "./e2e/global-setup.ts",
  fullyParallel: true,
  // More parallel browsers crash Playwright's WebKit on Windows ("Page crashed",
  // exit code 0xC0000409); with 2 workers the suite is stable.
  workers: 2,
  reporter: [["list"]],
  use: {
    baseURL: `http://localhost:${WEB_PORT}`,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    { name: "iphone", use: { ...devices["iPhone 13"] } },
    { name: "desktop", use: { ...devices["Desktop Chrome"] } },
  ],
  webServer: [
    {
      command: "go run ./cmd/api",
      cwd: "../backend",
      // Readiness without the database: web servers start before globalSetup creates
      // monetra_e2e, and this route answers 401 (accepted as "up") without touching it.
      url: `${E2E_API_URL}/api/v1/auth/me`,
      env: {
        ...process.env,
        DATABASE_URL: E2E_DATABASE_URL,
        PORT: String(API_PORT),
        JWT_SECRET: E2E_JWT_SECRET,
        CORS_ORIGIN: `http://localhost:${WEB_PORT}`,
      },
      timeout: 120_000,
    },
    {
      command: `npm run build && npx next start -p ${WEB_PORT}`,
      url: `http://localhost:${WEB_PORT}/login`,
      env: { ...process.env, API_URL: E2E_API_URL, AUTH_TRUST_HOST: "true" },
      timeout: 300_000,
    },
  ],
});
