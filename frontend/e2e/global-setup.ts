import { execSync } from "node:child_process";
import { E2E_DATABASE_NAME } from "./env";

function compose(args: string): string {
  return execSync(`docker compose -f ../docker-compose.yml ${args}`, { stdio: "pipe" }).toString().trim();
}

// Prepares an empty, migrated monetra_e2e database before the tests run.
export default function globalSetup() {
  const exists = compose(
    `exec -T db psql -U monetra -tAc "SELECT 1 FROM pg_database WHERE datname = '${E2E_DATABASE_NAME}'"`,
  );
  if (exists !== "1") {
    compose(`exec -T db createdb -U monetra ${E2E_DATABASE_NAME}`);
  }
  compose(
    `run --rm migrate -path /migrations -database "postgres://monetra:monetra@db:5432/${E2E_DATABASE_NAME}?sslmode=disable" up`,
  );
  compose(`exec -T db psql -U monetra -d ${E2E_DATABASE_NAME} -c "TRUNCATE users CASCADE"`);
}
