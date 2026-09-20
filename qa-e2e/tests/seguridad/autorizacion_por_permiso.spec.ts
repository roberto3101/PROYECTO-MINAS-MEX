import { test, expect, type APIRequestContext } from "@playwright/test";
import {
  comoAdministrador,
  contextoConToken,
  iniciarSesionDeEmpresa,
  exigirBaseDesechable,
  sello,
  contrasenaValida
} from "../soporte/base";


const PUERTAS = [
  { metodo: "get", ruta: "/gobierno/usuarios", permiso: "usuarios.ver" },
  { metodo: "post", ruta: "/gobierno/usuarios", permiso: "usuarios.crear" },
  { metodo: "get", ruta: "/gobierno/roles", permiso: "roles.ver" },
  { metodo: "get", ruta: "/gobierno/permisos", permiso: "roles.ver" },
  { metodo: "post", ruta: "/gobierno/roles", permiso: "roles.crear" },
  { metodo: "post", ruta: "/gobierno/asignaciones", permiso: "roles.asignar" },
  { metodo: "get", ruta: "/catalogos/minas", permiso: "catalogos.ver" },
  { metodo: "post", ruta: "/catalogos/minas", permiso: "catalogos.editar" },
  { metodo: "get", ruta: "/catalogos/empleados", permiso: "catalogos.ver" },
  { metodo: "get", ruta: "/catalogos/equipos", permiso: "catalogos.ver" },
  { metodo: "put", ruta: "/gobierno/empresa", permiso: "empresa.configurar" }
] as const;

let sinPermisos: APIRequestContext;
let credenciales: { usuario: string; contrasena: string };

test.beforeAll(async () => {
  exigirBaseDesechable();
  const administrador = await comoAdministrador();
  credenciales = { usuario: sello("qa.sinpermiso"), contrasena: contrasenaValida() };
  const creado = await administrador.post("/gobierno/usuarios", {
    data: {
      usuario: credenciales.usuario,
      nombre: "QA SIN PERMISOS",
      correo: "",
      contrasena: credenciales.contrasena,
      id_empleado: ""
    }
  });
  if (creado.status() !== 201) {
    throw new Error(`No se pudo preparar el usuario sin permisos: ${await creado.text()}`);
  }
  await administrador.dispose();

  const sesion = await iniciarSesionDeEmpresa(credenciales.usuario, credenciales.contrasena);
  if (sesion.permisos.length !== 0) {
    throw new Error(`El usuario nuevo llego con permisos: ${sesion.permisos.join(", ")}`);
  }
  sinPermisos = await contextoConToken(sesion.token);
});

test.afterAll(async () => {
  await sinPermisos?.dispose();
});

for (const puerta of PUERTAS) {
  test(`un usuario sin permisos recibe 403 en ${puerta.metodo.toUpperCase()} ${puerta.ruta}`, async () => {
    const respuesta = await (sinPermisos as any)[puerta.metodo](puerta.ruta, { data: {} });

    expect(
      respuesta.status(),
      `${puerta.ruta} dejo pasar a un usuario sin el permiso ${puerta.permiso}`
    ).toBe(403);

    expect(
      (await respuesta.json()).error,
      "el rechazo debe nombrar el permiso que falta, no un error generico"
    ).toBe(`permiso insuficiente: ${puerta.permiso}`);
  });
}

test("el mismo usuario, una vez que recibe el rol, si entra", async () => {
  const administrador = await comoAdministrador();
  const usuarios = await (
    await administrador.get(`/gobierno/usuarios?busqueda=${credenciales.usuario}&limite=5`)
  ).json();
  const objetivo = usuarios.Elementos.find((u: any) => u.NombreCorto === credenciales.usuario);
  expect(objetivo, "no se encontro el usuario de prueba").toBeTruthy();

  const codigoRol = sello("QA_LECTOR");
  const rolCreado = await administrador.post("/gobierno/roles", {
    data: { codigo: codigoRol, descripcion: "Rol de prueba solo lectura" }
  });
  expect(rolCreado.status(), "no se pudo crear el rol de prueba").toBe(201);
  const { id: identificadorRol } = await rolCreado.json();

  const concedido = await administrador.post(`/gobierno/roles/${identificadorRol}/permisos`, {
    data: { permiso: "usuarios.ver" }
  });
  expect(concedido.status(), "no se pudo conceder el permiso").toBe(200);

  const asignado = await administrador.post("/gobierno/asignaciones", {
    data: { id_usuario: objetivo.Identificador, id_rol: identificadorRol, id_mina: "" }
  });
  expect(asignado.status(), "no se pudo asignar el rol").toBe(201);
  await administrador.dispose();

  const sesion = await iniciarSesionDeEmpresa(credenciales.usuario, credenciales.contrasena);
  expect(sesion.permisos, "el permiso concedido no llego a la sesion").toContain("usuarios.ver");

  const conPermiso = await contextoConToken(sesion.token);
  expect(
    (await conPermiso.get("/gobierno/usuarios")).status(),
    "con el permiso concedido deberia poder listar usuarios"
  ).toBe(200);
  expect(
    (await conPermiso.post("/gobierno/usuarios", { data: {} })).status(),
    "usuarios.ver no debe habilitar usuarios.crear"
  ).toBe(403);
  await conPermiso.dispose();
});
