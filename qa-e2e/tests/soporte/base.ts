import { request, type APIRequestContext } from "@playwright/test";

export const API = process.env.URL_API ?? "http://localhost:8099";
export const NOMBRE_BASE = process.env.BASE_QA ?? "mina_qa";

export const EMPRESA_PRINCIPAL = "MIN";
export const EMPRESA_AJENA = "MIN2";
export const ID_EMPRESA_PRINCIPAL = "11111111-1111-1111-1111-111111111111";

export const ADMINISTRADOR = { usuario: "admin.mina", contrasena: "Mina#2026" };
export const SUPERADMIN = { usuario: "plataforma", contrasena: "Plataforma#2026" };

export function exigirBaseDesechable(): void {
  if (!/qa/i.test(NOMBRE_BASE)) {
    throw new Error(
      `Esta suite ESCRIBE. La base configurada es "${NOMBRE_BASE}" y no parece desechable. ` +
        `Apunta BASE_QA/CADENA_QA a la base de pruebas (mina_qa) antes de correr.`
    );
  }
}

export async function contextoAnonimo(): Promise<APIRequestContext> {
  return request.newContext({ baseURL: API });
}

export async function contextoConToken(token: string): Promise<APIRequestContext> {
  return request.newContext({
    baseURL: API,
    extraHTTPHeaders: { Authorization: `Bearer ${token}` }
  });
}

export type SesionDeEmpresa = { token: string; permisos: string[] };

export async function iniciarSesionDeEmpresa(
  usuario = ADMINISTRADOR.usuario,
  contrasena = ADMINISTRADOR.contrasena,
  codigoEmpresa = EMPRESA_PRINCIPAL
): Promise<SesionDeEmpresa> {
  const anonimo = await contextoAnonimo();
  const respuesta = await anonimo.post("/sesiones", {
    data: { codigo_empresa: codigoEmpresa, usuario, contrasena }
  });
  const cuerpo = await respuesta.json();
  await anonimo.dispose();
  if (respuesta.status() !== 200) {
    throw new Error(`No se pudo iniciar sesion como ${usuario}: ${respuesta.status()} ${JSON.stringify(cuerpo)}`);
  }
  return { token: cuerpo.token, permisos: cuerpo.permisos ?? [] };
}

export async function iniciarSesionDePlataforma(): Promise<string> {
  const anonimo = await contextoAnonimo();
  const respuesta = await anonimo.post("/plataforma/sesiones", { data: SUPERADMIN });
  const cuerpo = await respuesta.json();
  await anonimo.dispose();
  if (respuesta.status() !== 200) {
    throw new Error(`No se pudo iniciar sesion de plataforma: ${respuesta.status()} ${JSON.stringify(cuerpo)}`);
  }
  return cuerpo.token;
}

export async function comoAdministrador(): Promise<APIRequestContext> {
  const sesion = await iniciarSesionDeEmpresa();
  return contextoConToken(sesion.token);
}

export async function comoSuperadmin(): Promise<APIRequestContext> {
  return contextoConToken(await iniciarSesionDePlataforma());
}

let contador = 0;
export function sello(prefijo: string): string {
  contador += 1;
  const corrida = process.env.SELLO_CORRIDA ?? String(Date.now()).slice(-6);
  return `${prefijo}${corrida}${String(contador).padStart(2, "0")}`;
}

export function contrasenaValida(): string {
  return `Clave${sello("Qa")}9`;
}
