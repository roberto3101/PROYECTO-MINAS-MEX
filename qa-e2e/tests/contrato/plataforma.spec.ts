import { test, expect } from "@playwright/test";
import {
  comoSuperadmin,
  contextoAnonimo,
  iniciarSesionDeEmpresa,
  exigirBaseDesechable,
  sello,
  contrasenaValida
} from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

function empresaNueva() {
  return {
    codigo: sello("QA").toUpperCase().slice(0, 12),
    usuario: sello("qa.jefe"),
    contrasena: contrasenaValida()
  };
}

test("aprovisionar una empresa la deja lista para operar: admin, rol y catalogos", async () => {
  const superadmin = await comoSuperadmin();
  const nueva = empresaNueva();

  const creada = await superadmin.post("/plataforma/empresas", {
    data: {
      codigo: nueva.codigo,
      razon_social: "Minera QA de Prueba SA",
      identificacion_fiscal: "QAX010101AAA",
      correo_contacto: "contacto@qa.mx",
      telefono: "5550001122",
      zona_horaria: "America/Mexico_City",
      moneda: "MXN",
      color_primario: "#D9A440",
      admin: {
        usuario: nueva.usuario,
        nombre: "QA ADMINISTRADOR",
        correo: "jefe@qa.mx",
        contrasena: nueva.contrasena
      }
    }
  });
  expect(creada.status(), `no se pudo aprovisionar: ${await creada.text()}`).toBe(201);
  const { id, id_admin } = await creada.json();

  const empresa = await consultar<{ codigo: string; estado: string; moneda_defecto: string }>(
    "SELECT codigo, estado, moneda_defecto FROM gobierno.empresa WHERE id = $1",
    [id]
  );
  expect(empresa[0].codigo).toBe(nueva.codigo);
  expect(empresa[0].estado, "una empresa nueva debe nacer activa").toBe("ACTIVA");

  const rolAdministrador = await consultar<{ codigo: string; es_sistema: boolean }>(
    "SELECT codigo, es_sistema FROM gobierno.rol WHERE id_empresa = $1",
    [id]
  );
  expect(
    rolAdministrador.map((r) => r.codigo),
    "al crear la empresa solo debe sembrarse el rol Administrador"
  ).toEqual(["ADMIN_EMPRESA"]);
  expect(rolAdministrador[0].es_sistema).toBe(true);

  const asignacion = await consultar<{ id_mina: string | null }>(
    `SELECT ur.id_mina::text FROM gobierno.usuario_rol ur
     WHERE ur.id_usuario = $1 AND ur.eliminado_en IS NULL`,
    [id_admin]
  );
  expect(asignacion, "el administrador inicial no quedo con su rol").toHaveLength(1);
  expect(
    asignacion[0].id_mina,
    "el administrador inicial debe nacer con alcance a toda la empresa"
  ).toBeNull();

  const catalogos = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM catalogos.departamento WHERE id_empresa = $1",
    [id]
  );
  expect(Number(catalogos[0].total), "no se sembraron los catalogos basicos").toBeGreaterThan(0);
  await superadmin.dispose();
});

test("el administrador de la empresa nueva entra y manda en su empresa", async () => {
  const superadmin = await comoSuperadmin();
  const nueva = empresaNueva();
  const creada = await superadmin.post("/plataforma/empresas", {
    data: {
      codigo: nueva.codigo,
      razon_social: "Minera QA Segunda SA",
      identificacion_fiscal: "",
      correo_contacto: "",
      telefono: "",
      zona_horaria: "America/Mexico_City",
      moneda: "USD",
      color_primario: "",
      admin: { usuario: nueva.usuario, nombre: "QA JEFE", correo: "", contrasena: nueva.contrasena }
    }
  });
  expect(creada.status()).toBe(201);
  await superadmin.dispose();

  const sesion = await iniciarSesionDeEmpresa(nueva.usuario, nueva.contrasena, nueva.codigo);
  const permisos = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.permiso WHERE eliminado_en IS NULL AND estado = 'ACTIVO'"
  );
  expect(
    sesion.permisos.length,
    "el administrador inicial debe traer todos los permisos del catalogo"
  ).toBe(Number(permisos[0].total));
});

test("no se puede crear dos empresas con el mismo codigo", async () => {
  const superadmin = await comoSuperadmin();
  const nueva = empresaNueva();
  const cuerpo = {
    codigo: nueva.codigo,
    razon_social: "Minera QA Duplicada SA",
    identificacion_fiscal: "",
    correo_contacto: "",
    telefono: "",
    zona_horaria: "America/Mexico_City",
    moneda: "MXN",
    color_primario: "",
    admin: { usuario: nueva.usuario, nombre: "QA UNO", correo: "", contrasena: nueva.contrasena }
  };

  expect((await superadmin.post("/plataforma/empresas", { data: cuerpo })).status()).toBe(201);

  const repetida = await superadmin.post("/plataforma/empresas", {
    data: { ...cuerpo, admin: { ...cuerpo.admin, usuario: sello("qa.otro") } }
  });
  expect(repetida.status(), "se permitio repetir el codigo de empresa").toBe(409);

  const cuantas = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM gobierno.empresa WHERE codigo = $1",
    [nueva.codigo]
  );
  expect(Number(cuantas[0].total), "quedaron dos empresas con el mismo codigo").toBe(1);
  await superadmin.dispose();
});

test("suspender una empresa deja a su gente fuera, y reactivarla los devuelve", async () => {
  const superadmin = await comoSuperadmin();
  const nueva = empresaNueva();
  const creada = await superadmin.post("/plataforma/empresas", {
    data: {
      codigo: nueva.codigo,
      razon_social: "Minera QA Suspendible SA",
      identificacion_fiscal: "",
      correo_contacto: "",
      telefono: "",
      zona_horaria: "America/Mexico_City",
      moneda: "MXN",
      color_primario: "",
      admin: { usuario: nueva.usuario, nombre: "QA SUSPENSO", correo: "", contrasena: nueva.contrasena }
    }
  });
  const { id } = await creada.json();

  await iniciarSesionDeEmpresa(nueva.usuario, nueva.contrasena, nueva.codigo);

  const suspendida = await superadmin.patch(`/plataforma/empresas/${id}/estado`, {
    data: { estado: "INACTIVA" }
  });
  expect(suspendida.status()).toBe(200);

  const anonimo = await contextoAnonimo();
  const bloqueado = await anonimo.post("/sesiones", {
    data: { codigo_empresa: nueva.codigo, usuario: nueva.usuario, contrasena: nueva.contrasena }
  });
  expect(
    bloqueado.status(),
    "con la empresa suspendida su administrador siguio entrando"
  ).toBe(401);
  await anonimo.dispose();

  const reactivada = await superadmin.patch(`/plataforma/empresas/${id}/estado`, {
    data: { estado: "ACTIVA" }
  });
  expect(reactivada.status()).toBe(200);

  const sesion = await iniciarSesionDeEmpresa(nueva.usuario, nueva.contrasena, nueva.codigo);
  expect(sesion.token, "tras reactivar la empresa su administrador no pudo volver a entrar").toBeTruthy();
  await superadmin.dispose();
});

test("el detalle de empresa que ve la plataforma cuadra con la base", async () => {
  const superadmin = await comoSuperadmin();
  const empresas = await (await superadmin.get("/plataforma/empresas?limite=50")).json();
  const principal = empresas.Elementos.find((e: any) => e.Codigo === "MIN");
  expect(principal, "no se encontro la empresa MIN").toBeTruthy();

  const detalle = await (await superadmin.get(`/plataforma/empresas/${principal.Identificador}`)).json();

  const reales = await consultar<{ usuarios: string; minas: string }>(
    `SELECT
       (SELECT count(*) FROM gobierno.usuario u WHERE u.id_empresa = $1 AND u.eliminado_en IS NULL)::text AS usuarios,
       (SELECT count(*) FROM catalogos.mina m WHERE m.id_empresa = $1 AND m.eliminado_en IS NULL)::text AS minas`,
    [principal.Identificador]
  );
  expect(String(detalle.TotalUsuarios), "el total de usuarios no cuadra con la base").toBe(reales[0].usuarios);
  expect(String(detalle.TotalMinas), "el total de minas no cuadra con la base").toBe(reales[0].minas);
  await superadmin.dispose();
});

test("una empresa inexistente responde 404 en el panel de plataforma", async () => {
  const superadmin = await comoSuperadmin();
  const respuesta = await superadmin.get("/plataforma/empresas/00000000-0000-0000-0000-000000000000");
  expect(respuesta.status()).toBe(404);
  await superadmin.dispose();
});
