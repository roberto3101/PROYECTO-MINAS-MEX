import { test, expect } from "@playwright/test";
import {
  comoAdministrador,
  contextoAnonimo,
  iniciarSesionDeEmpresa,
  exigirBaseDesechable,
  EMPRESA_PRINCIPAL,
  sello,
  contrasenaValida
} from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

test("crear un usuario lo deja realmente en la base, con su empresa y su contrasena cifrada", async () => {
  const administrador = await comoAdministrador();
  const nombre = sello("qa.alta");
  const clave = contrasenaValida();

  const respuesta = await administrador.post("/gobierno/usuarios", {
    data: { usuario: nombre, nombre: "QA ALTA", correo: "qa.alta@minera.mx", contrasena: clave, id_empleado: "" }
  });
  expect(respuesta.status()).toBe(201);
  const { id } = await respuesta.json();

  const filas = await consultar<{
    usuario: string;
    nombre: string;
    correo: string | null;
    estado: string;
    contrasena_hash: string | null;
    codigo: string;
  }>(
    `SELECT u.usuario, u.nombre, u.correo, u.estado, u.contrasena_hash, e.codigo
     FROM gobierno.usuario u JOIN gobierno.empresa e ON e.id = u.id_empresa
     WHERE u.id = $1`,
    [id]
  );

  expect(filas, "el usuario no aparecio en la base").toHaveLength(1);
  expect(filas[0].usuario).toBe(nombre);
  expect(filas[0].estado, "un usuario nuevo debe nacer activo").toBe("ACTIVO");
  expect(filas[0].codigo, "el usuario quedo en otra empresa").toBe(EMPRESA_PRINCIPAL);
  expect(filas[0].contrasena_hash, "la contrasena no quedo cifrada con bcrypt").toMatch(/^\$2[aby]\$/);
  expect(filas[0].contrasena_hash, "la contrasena quedo en claro").not.toContain(clave);
  await administrador.dispose();
});

test("el mismo nombre de usuario no se puede repetir dentro de la empresa", async () => {
  const administrador = await comoAdministrador();
  const nombre = sello("qa.dup");
  const datos = {
    usuario: nombre,
    nombre: "QA DUPLICADO",
    correo: "",
    contrasena: contrasenaValida(),
    id_empleado: ""
  };

  expect((await administrador.post("/gobierno/usuarios", { data: datos })).status()).toBe(201);
  const repetido = await administrador.post("/gobierno/usuarios", { data: datos });

  expect(repetido.status(), "se permitio duplicar el nombre de usuario").toBe(409);

  const cuantos = await consultar<{ total: string }>(
    `SELECT count(*)::text AS total FROM gobierno.usuario WHERE usuario = $1 AND eliminado_en IS NULL`,
    [nombre]
  );
  expect(Number(cuantos[0].total), "quedaron dos usuarios con el mismo nombre").toBe(1);
  await administrador.dispose();
});

test("un usuario recien creado puede iniciar sesion con su contrasena", async () => {
  const administrador = await comoAdministrador();
  const nombre = sello("qa.entra");
  const clave = contrasenaValida();
  expect(
    (
      await administrador.post("/gobierno/usuarios", {
        data: { usuario: nombre, nombre: "QA ENTRA", correo: "", contrasena: clave, id_empleado: "" }
      })
    ).status()
  ).toBe(201);
  await administrador.dispose();

  const sesion = await iniciarSesionDeEmpresa(nombre, clave);
  expect(sesion.token, "no se emitio token para el usuario nuevo").toBeTruthy();
  expect(sesion.permisos, "un usuario sin roles no debe traer permisos").toHaveLength(0);
});

test("desactivar a un usuario le cierra la puerta de inmediato", async () => {
  const administrador = await comoAdministrador();
  const nombre = sello("qa.baja");
  const clave = contrasenaValida();
  const creado = await administrador.post("/gobierno/usuarios", {
    data: { usuario: nombre, nombre: "QA BAJA", correo: "", contrasena: clave, id_empleado: "" }
  });
  const { id } = await creado.json();

  await iniciarSesionDeEmpresa(nombre, clave);

  const baja = await administrador.patch(`/gobierno/usuarios/${id}/estado`, {
    data: { estado: "INACTIVO" }
  });
  expect(baja.status(), "no se pudo desactivar al usuario").toBe(200);

  const estado = await consultar<{ estado: string }>(
    "SELECT estado FROM gobierno.usuario WHERE id = $1",
    [id]
  );
  expect(estado[0].estado).toBe("INACTIVO");

  const anonimo = await contextoAnonimo();
  const intento = await anonimo.post("/sesiones", {
    data: { codigo_empresa: EMPRESA_PRINCIPAL, usuario: nombre, contrasena: clave }
  });
  expect(intento.status(), "un usuario desactivado siguio pudiendo entrar").toBe(401);
  await anonimo.dispose();

  const reactivado = await administrador.patch(`/gobierno/usuarios/${id}/estado`, {
    data: { estado: "INACTIVO" }
  });
  expect(reactivado.status(), "desactivar dos veces debe avisar del conflicto").toBe(409);
  await administrador.dispose();
});

test("editar un usuario cambia sus datos y no toca su contrasena", async () => {
  const administrador = await comoAdministrador();
  const nombre = sello("qa.edita");
  const clave = contrasenaValida();
  const creado = await administrador.post("/gobierno/usuarios", {
    data: { usuario: nombre, nombre: "NOMBRE ORIGINAL", correo: "", contrasena: clave, id_empleado: "" }
  });
  const { id } = await creado.json();
  const antes = await consultar<{ contrasena_hash: string }>(
    "SELECT contrasena_hash FROM gobierno.usuario WHERE id = $1",
    [id]
  );

  const editado = await administrador.put(`/gobierno/usuarios/${id}`, {
    data: { nombre: "NOMBRE CORREGIDO", correo: "corregido@minera.mx" }
  });
  expect(editado.status()).toBe(200);

  const despues = await consultar<{ nombre: string; correo: string; contrasena_hash: string }>(
    "SELECT nombre, correo, contrasena_hash FROM gobierno.usuario WHERE id = $1",
    [id]
  );
  expect(despues[0].nombre).toBe("NOMBRE CORREGIDO");
  expect(despues[0].correo).toBe("corregido@minera.mx");
  expect(
    despues[0].contrasena_hash,
    "editar el perfil no debe alterar la contrasena"
  ).toBe(antes[0].contrasena_hash);

  const sesion = await iniciarSesionDeEmpresa(nombre, clave);
  expect(sesion.token, "tras editar el perfil el usuario ya no pudo entrar").toBeTruthy();
  await administrador.dispose();
});

test("pedir un usuario inexistente responde 404 y no 500", async () => {
  const administrador = await comoAdministrador();
  const respuesta = await administrador.get("/gobierno/usuarios/00000000-0000-0000-0000-000000000000");
  expect(respuesta.status()).toBe(404);
  await administrador.dispose();
});
