import { defineConfig } from "@playwright/test";

const visible = !!process.env.VER_NAVEGADOR;
const pausa = Number(process.env.PAUSA) || (visible ? 700 : 0);

export default defineConfig({
  testDir: "./tests",
  timeout: visible ? 180000 : 60000,
  expect: { timeout: 10000 },
  fullyParallel: false,
  workers: 1,
  retries: 0,
  globalTeardown: "./cierre.ts",
  reporter: [
    ["list"],
    ["json", { outputFile: "resultados.json" }],
    ["./reporte/informe.ts"]
  ],
  use: {
    baseURL: process.env.URL_FRONT ?? "http://localhost:8099",
    headless: !visible,
    launchOptions: { slowMo: pausa, args: visible ? ["--start-maximized"] : [] },
    navigationTimeout: 30000,
    actionTimeout: 12000,
    screenshot: "only-on-failure"
  }
});
