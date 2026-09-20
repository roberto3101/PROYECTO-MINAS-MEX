import { test, expect } from "@playwright/test";
import {
  comoAdministrador,
  contextoConToken,
  iniciarSesionDeEmpresa,
  exigirBaseDesechable,
  sello,
  contrasenaValida
} from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

async function prepararUsuarioConAccesoAUnaMina(): Promise<{
  usuario: string;
  clave: string;
  identificadorMina: string;
  nombreMina: string;
}> {
  const administrador = await comoAdministrador();
  const minas = await (await administrador.get("/catalogos/minas?limite=50")).json();
  expect(minas.Elementos.length, "la semilla debe traer al menos dos minas").toBeGreaterThan(1);
  const elegida = minas.Elementos[0];

  const usuario = sello("qa.unamina");
  const clave = contrasenaValida();
  const creado = await administrador.post("/gobierno/usuarios", {
    data: { usuario, nombre: "QA UNA MINA", correo: "", contrasena: clave, id_empleado: "" }
  });
  expect(creado.status()).toBe(201);
  const { id } = await creado.json();

  const rol = await administrador.post("/gobierno/roles", {
    data: { codigo: sello("QA_MINA"), descripcion: "Rol de prueba con alcance de mina" }
  });
  const { id: identificadorRol } = await rol.json();
  await administrador.post(`/gobierno/roles/${identificadorRol}/permisos`, {
    data: { permiso: "catalogos.ver" }
  });

  const asignado = await administrador.post("/gobierno/asignaciones", {
    data: { id_usuario: id, id_rol: identificadorRol, id_mina: elegida.Identificador }
  });
  expect(asignado.status(), "no se pudo asignar el rol con alcance de mina").toBe(201);
  await administrador.dispose();

  return {
    usuario,
    clave,
    identificadorMina: elegida.Identificador,
    nombreMina: elegida.Nombre
  };
}

test("el administrador tiene alcance global y ve todas las minas de su empresa", async () => {
  const administrador = await comoAdministrador();
  const acceso = await (await administrador.get("/gobierno/sesion/minas")).json();

  expect(acceso.EsGlobal, "el administrador debe tener alcance global").toBe(true);

  const enBase = await consultar<{ total: string }>(
    `SELECT count(*)::text AS total FROM catalogos.mina m
     JOIN gobierno.empresa e ON e.id = m.id_empresa
     WHERE e.codigo = 'MIN' AND m.eliminado_en IS NULL AND m.estado = 'ACTIVA'`
  );
  expect(
    acceso.Minas.length,
    "el endpoint no devolvio todas las minas activas de la empresa"
  ).toBe(Number(enBase[0].total));
  await administrador.dispose();
});

test("un usuario con acceso a una sola mina no ve las demas", async () => {
  const preparado = await prepararUsuarioConAccesoAUnaMina();
  const sesion = await iniciarSesionDeEmpresa(preparado.usuario, preparado.clave);
  const restringido = await contextoConToken(sesion.token);

  const acceso = await (await restringido.get("/gobierno/sesion/minas")).json();
  expect(acceso.EsGlobal, "este usuario no deberia tener alcance global").toBe(false);
  expect(acceso.Minas, "deberia ver exactamente una mina").toHaveLength(1);
  expect(acceso.Minas[0].Identificador).toBe(preparado.identificadorMina);

  const minas = await (await restringido.get("/catalogos/minas?limite=50")).json();
  expect(
    minas.Elementos.map((m: any) => m.Identificador),
    "el listado de minas debe quedar acotado a su alcance"
  ).toEqual([preparado.identificadorMina]);
  await restringido.dispose();
});

test("un usuario con una mina no puede ver empleados de otra mina ni forzando el filtro", async () => {
  const preparado = await prepararUsuarioConAccesoAUnaMina();
  const sesion = await iniciarSesionDeEmpresa(preparado.usuario, preparado.clave);
  const restringido = await contextoConToken(sesion.token);

  const otras = await consultar<{ id: string }>(
    `SELECT m.id FROM catalogos.mina m
     JOIN gobierno.empresa e ON e.id = m.id_empresa
     WHERE e.codigo = 'MIN' AND m.id <> $1 AND m.eliminado_en IS NULL LIMIT 1`,
    [preparado.identificadorMina]
  );
  test.skip(otras.length === 0, "la empresa solo tiene una mina");

  const forzado = await restringido.get(`/catalogos/empleados?mina=${otras[0].id}&limite=50`);
  expect(forzado.status()).toBe(200);
  const cuerpo = await forzado.json();
  expect(
    cuerpo.Elementos ?? [],
    "pidiendo explicitamente otra mina devolvio datos: el alcance se puede burlar"
  ).toHaveLength(0);

  const propios = await restringido.get(
    `/catalogos/empleados?mina=${preparado.identificadorMina}&limite=50`
  );
  expect(propios.status(), "su propia mina si debe responder").toBe(200);
  await restringido.dispose();
});

test("revocar la asignacion deja al usuario sin acceso a ninguna mina", async () => {
  const preparado = await prepararUsuarioConAccesoAUnaMina();
  const administrador = await comoAdministrador();

  const usuarios = await (
    await administrador.get(`/gobierno/usuarios?busqueda=${preparado.usuario}&limite=5`)
  ).json();
  const objetivo = usuarios.Elementos.find((u: any) => u.NombreCorto === preparado.usuario);
  const asignaciones = await (
    await administrador.get(`/gobierno/usuarios/${objetivo.Identificador}/asignaciones`)
  ).json();
  expect(asignaciones.length, "el usuario deberia tener una asignacion").toBeGreaterThan(0);
  expect(
    asignaciones[0].AlcanceMina,
    "la asignacion debe registrar la mina de alcance"
  ).toBe(preparado.identificadorMina);

  const revocado = await administrador.delete(`/gobierno/asignaciones/${asignaciones[0].Identificador}`);
  expect(revocado.status(), "no se pudo revocar la asignacion").toBe(200);
  await administrador.dispose();

  const sesion = await iniciarSesionDeEmpresa(preparado.usuario, preparado.clave);
  expect(sesion.permisos, "tras revocar el rol no debe quedar ningun permiso").toHaveLength(0);

  const restringido = await contextoConToken(sesion.token);
  const acceso = await (await restringido.get("/gobierno/sesion/minas")).json();
  expect(acceso.EsGlobal).toBe(false);
  expect(acceso.Minas ?? [], "tras revocar no debe quedar acceso a ninguna mina").toHaveLength(0);
  await restringido.dispose();
});

test("el superadmin puede dar acceso a una mina a un usuario de cualquier empresa", async () => {
  const { comoSuperadmin } = await import("../soporte/base");
  const superadmin = await comoSuperadmin();

  const empresas = await (await superadmin.get("/plataforma/empresas?limite=20")).json();
  const empresa = empresas.Elementos.find((e: any) => e.Codigo === "MIN");
  expect(empresa, "no se encontro la empresa MIN desde plataforma").toBeTruthy();

  const usuarios = await (
    await superadmin.get(`/plataforma/empresas/${empresa.Identificador}/usuarios?limite=20`)
  ).json();
  const roles = await (
    await superadmin.get(`/plataforma/empresas/${empresa.Identificador}/roles`)
  ).json();
  const minas = await (
    await superadmin.get(`/plataforma/empresas/${empresa.Identificador}/minas?limite=20`)
  ).json();

  expect(usuarios.Elementos.length, "plataforma no vio usuarios de la empresa").toBeGreaterThan(0);
  expect(roles.length, "plataforma no vio roles de la empresa").toBeGreaterThan(0);
  expect(minas.Elementos.length, "plataforma no vio minas de la empresa").toBeGreaterThan(0);

  const objetivo = usuarios.Elementos[0];
  const asignado = await superadmin.post(`/plataforma/empresas/${empresa.Identificador}/asignaciones`, {
    data: {
      id_usuario: objetivo.Identificador,
      id_rol: roles[0].Identificador,
      id_mina: minas.Elementos[0].Identificador
    }
  });
  expect(asignado.status(), "el superadmin no pudo asignar acceso a una mina").toBe(201);
  const { id: identificadorAsignacion } = await asignado.json();

  const enBase = await consultar<{ id_mina: string }>(
    "SELECT id_mina::text FROM gobierno.usuario_rol WHERE id = $1",
    [identificadorAsignacion]
  );
  expect(
    enBase[0]?.id_mina,
    "la asignacion no quedo con la mina que eligio el superadmin"
  ).toBe(minas.Elementos[0].Identificador);

  const revocado = await superadmin.delete(
    `/plataforma/empresas/${empresa.Identificador}/asignaciones/${identificadorAsignacion}`
  );
  expect(revocado.status(), "el superadmin no pudo revocar lo que asigno").toBe(200);
  await superadmin.dispose();
});
