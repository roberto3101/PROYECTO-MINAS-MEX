import { test, expect } from "@playwright/test";
import {
  contextoAnonimo,
  comoAdministrador,
  exigirBaseDesechable,
  EMPRESA_PRINCIPAL,
  sello
} from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

test("adivinar contrasenas se bloquea tras varios intentos y lo dice con segundos de espera", async () => {
  const anonimo = await contextoAnonimo();
  const victima = sello("qa.fuerza");
  let bloqueado = false;
  let esperaSegundos = 0;

  for (let intento = 1; intento <= 8; intento += 1) {
    const respuesta = await anonimo.post("/sesiones", {
      data: { codigo_empresa: EMPRESA_PRINCIPAL, usuario: victima, contrasena: `intento-${intento}` }
    });
    if (respuesta.status() === 429) {
      bloqueado = true;
      esperaSegundos = (await respuesta.json()).reintentar_en_segundos ?? 0;
      break;
    }
    expect(
      respuesta.status(),
      "un intento fallido debe responder 401, sin revelar si el usuario existe"
    ).toBe(401);
  }

  expect(bloqueado, "tras 8 intentos fallidos seguidos nunca se bloqueo: fuerza bruta abierta").toBe(true);
  expect(esperaSegundos, "el bloqueo debe decir cuantos segundos esperar").toBeGreaterThan(0);
  await anonimo.dispose();
});

test("el mensaje de credenciales no revela si el usuario existe", async () => {
  const anonimo = await contextoAnonimo();
  const inexistente = await anonimo.post("/sesiones", {
    data: { codigo_empresa: EMPRESA_PRINCIPAL, usuario: sello("qa.nadie"), contrasena: "LoQueSea123" }
  });
  const existentePeroMal = await anonimo.post("/sesiones", {
    data: { codigo_empresa: EMPRESA_PRINCIPAL, usuario: "admin.mina", contrasena: "ClaveIncorrecta123" }
  });

  expect(inexistente.status()).toBe(401);
  expect(existentePeroMal.status()).toBe(401);
  expect(
    (await inexistente.json()).error,
    "los dos casos deben responder identico para no enumerar usuarios"
  ).toBe((await existentePeroMal.json()).error);
  await anonimo.dispose();
});

const PAYLOADS = [
  "' OR '1'='1",
  "'; DROP TABLE gobierno.usuario; --",
  "admin' --",
  "%' UNION SELECT contrasena_hash FROM gobierno.usuario --"
];

for (const payload of PAYLOADS) {
  test(`la busqueda resiste la inyeccion: ${payload.slice(0, 24)}`, async () => {
    const antes = await consultar<{ total: string }>(
      "SELECT count(*)::text AS total FROM gobierno.usuario"
    );

    const administrador = await comoAdministrador();
    const respuesta = await administrador.get(
      `/gobierno/usuarios?busqueda=${encodeURIComponent(payload)}&limite=20`
    );

    expect(
      respuesta.status(),
      "la busqueda debe tratar el texto como dato, no romperse ni ejecutarlo"
    ).toBe(200);

    const cuerpo = await respuesta.json();
    expect(
      cuerpo.Elementos ?? [],
      "una inyeccion no debe devolver filas que el filtro no pidio"
    ).toHaveLength(0);

    const texto = JSON.stringify(cuerpo);
    expect(texto, "la respuesta filtro material de contrasenas").not.toMatch(/\$2[aby]\$/);
    await administrador.dispose();

    const despues = await consultar<{ total: string }>(
      "SELECT count(*)::text AS total FROM gobierno.usuario"
    );
    expect(
      despues[0].total,
      "la inyeccion altero la cantidad de usuarios: hubo ejecucion de SQL"
    ).toBe(antes[0].total);
  });
}

test("el cuerpo de la peticion no acepta campos desconocidos", async () => {
  const administrador = await comoAdministrador();
  const respuesta = await administrador.post("/gobierno/usuarios", {
    data: {
      usuario: sello("qa.extra"),
      nombre: "QA EXTRA",
      contrasena: "ClaveValida123",
      es_superadmin: true,
      id_empresa: "11111111-1111-1111-1111-111111111111"
    }
  });

  expect(
    respuesta.status(),
    "aceptar campos desconocidos permite inyectar atributos como es_superadmin"
  ).toBe(400);
  expect((await respuesta.json()).error).toContain("campos desconocidos");
  await administrador.dispose();
});
