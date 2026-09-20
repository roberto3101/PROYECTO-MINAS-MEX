import { test, expect } from "@playwright/test";
import {
  contextoAnonimo,
  contextoConToken,
  iniciarSesionDeEmpresa,
  iniciarSesionDePlataforma,
  exigirBaseDesechable
} from "../soporte/base";

test.beforeAll(() => exigirBaseDesechable());

test("sin token la API no entrega nada", async () => {
  const anonimo = await contextoAnonimo();
  const respuesta = await anonimo.get("/gobierno/usuarios");
  expect(respuesta.status()).toBe(401);
  expect((await respuesta.json()).error).toBe("no autorizado");
  await anonimo.dispose();
});

test("un token inventado o con la firma alterada no abre nada", async () => {
  const basura = await contextoConToken("esto.no.es-un-token");
  expect((await basura.get("/gobierno/usuarios")).status()).toBe(401);
  await basura.dispose();

  const { token } = await iniciarSesionDeEmpresa();
  const partes = token.split(".");
  const firmaCambiada = `${partes[0]}.${partes[1]}.${partes[2].slice(0, -3)}aaa`;
  const alterado = await contextoConToken(firmaCambiada);
  expect(
    (await alterado.get("/gobierno/usuarios")).status(),
    "se acepto un token con la firma cambiada"
  ).toBe(401);
  await alterado.dispose();
});

test("alterar el contenido del token para darse permisos no funciona", async () => {
  const { token } = await iniciarSesionDeEmpresa();
  const [cabecera, cuerpo, firma] = token.split(".");
  const datos = JSON.parse(Buffer.from(cuerpo, "base64url").toString());
  datos.per = ["usuarios.ver", "usuarios.crear", "roles.asignar"];
  datos.alm = true;
  const cuerpoFalso = Buffer.from(JSON.stringify(datos)).toString("base64url").replace(/=+$/, "");

  const falsificado = await contextoConToken(`${cabecera}.${cuerpoFalso}.${firma}`);
  expect(
    (await falsificado.get("/gobierno/usuarios")).status(),
    "se acepto un token con el cuerpo modificado: la firma no se esta verificando"
  ).toBe(401);
  await falsificado.dispose();
});

test("el token de una empresa no sirve en plataforma, ni el de plataforma en una empresa", async () => {
  const { token } = await iniciarSesionDeEmpresa();
  const deEmpresa = await contextoConToken(token);
  expect(
    (await deEmpresa.get("/plataforma/empresas")).status(),
    "un administrador de empresa entro al panel de plataforma"
  ).toBe(401);
  await deEmpresa.dispose();

  const dePlataforma = await contextoConToken(await iniciarSesionDePlataforma());
  expect(
    (await dePlataforma.get("/gobierno/usuarios")).status(),
    "el superadmin uso su token dentro del ambito de una empresa"
  ).toBe(401);
  await dePlataforma.dispose();
});

test("las respuestas traen las cabeceras que protegen al navegador", async () => {
  const anonimo = await contextoAnonimo();
  const cabeceras = (await anonimo.get("/salud")).headers();
  expect(cabeceras["x-content-type-options"]).toBe("nosniff");
  expect(cabeceras["x-frame-options"]).toBe("DENY");
  expect(cabeceras["referrer-policy"]).toBe("no-referrer");
  expect(cabeceras["content-security-policy"], "falta la politica de contenido").toContain(
    "default-src 'self'"
  );
  await anonimo.dispose();
});

test("la contrasena nunca viaja de vuelta en las respuestas", async () => {
  const { token } = await iniciarSesionDeEmpresa();
  const administrador = await contextoConToken(token);
  const listado = await (await administrador.get("/gobierno/usuarios?limite=50")).text();
  expect(listado.toLowerCase(), "el listado de usuarios expone material de contrasena").not.toMatch(
    /contrasena_hash|\$2a\$|\$2b\$/
  );
  await administrador.dispose();
});
