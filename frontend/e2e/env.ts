// Settings shared by playwright.config.ts, the global setup, and the tests.
export const API_PORT = 8090;
export const WEB_PORT = 3100;
export const E2E_API_URL = `http://localhost:${API_PORT}`;
export const E2E_DATABASE_NAME = "monetra_e2e";
export const E2E_DATABASE_URL = `postgres://monetra:monetra@localhost:5434/${E2E_DATABASE_NAME}?sslmode=disable`;
export const E2E_JWT_SECRET = "e2e-only-secret-0123456789abcdefghij";
