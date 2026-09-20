import { test, expect, type APIRequestContext } from "@playwright/test";
import { comoAdministrador, exigirBaseDesechable, sello } from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

async function primeraOpcion(contexto: APIRequestContext, ruta: string): Promise<string> {
  const opciones = await (await contexto.get(ruta)).json();
  const lista = Array.isArray(opciones) ? opciones : (opciones.Elementos ?? []);
  expect(lista.length, `la semilla no trae opciones en ${ruta}`).toBeGreaterThan(0);
  return lista[0].Identificador;
}

test("crear una mina la deja activa y visible en el listado", async () => {
  const administrador = await comoAdministrador();
  const nombre = `Mina QA ${sello("")}`;

  const creada = await administrador.post("/catalogos/minas", {
    data: { nombre, area: "Norte", niveles: "1000-1200" }
  });
  expect(creada.status()).toBe(201);
  const { id } = await creada.json();

  const filas = await consultar<{ nombre: string; area: string; estado: string }>(
    "SELECT nombre, area, estado FROM catalogos.mina WHERE id = $1",
    [id]
  );
  expect(filas[0].nombre).toBe(nombre);
  expect(filas[0].estado, "una mina nueva debe nacer activa").toBe("ACTIVA");

  const listado = await (await administrador.get(`/catalogos/minas?busqueda=${encodeURIComponent(nombre)}&limite=20`)).json();
  expect(
    (listado.Elementos ?? []).map((m: any) => m.Identificador),
    "la mina recien creada no aparece en el listado"
  ).toContain(id);
  await administrador.dispose();
});

test("no se puede repetir el nombre de una mina", async () => {
  const administrador = await comoAdministrador();
  const nombre = `Mina Repetida ${sello("")}`;
  const datos = { nombre, area: "Sur", niveles: "" };

  expect((await administrador.post("/catalogos/minas", { data: datos })).status()).toBe(201);
  const repetida = await administrador.post("/catalogos/minas", { data: datos });
  expect(repetida.status(), "se permitio repetir el nombre de la mina").toBeGreaterThanOrEqual(400);

  const cuantas = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM catalogos.mina WHERE nombre = $1 AND eliminado_en IS NULL",
    [nombre]
  );
  expect(Number(cuantas[0].total), "quedaron dos minas con el mismo nombre").toBe(1);
  await administrador.dispose();
});

test("una mina se puede desactivar, y un estado inventado se rechaza", async () => {
  const administrador = await comoAdministrador();
  const creada = await administrador.post("/catalogos/minas", {
    data: { nombre: `Mina Baja ${sello("")}`, area: "Este", niveles: "" }
  });
  const { id } = await creada.json();

  const invalido = await administrador.patch(`/catalogos/minas/${id}/estado`, {
    data: { estado: "DORMIDA" }
  });
  expect(invalido.status(), "se acepto un estado que no existe").toBe(400);

  const sigueActiva = await consultar<{ estado: string }>(
    "SELECT estado FROM catalogos.mina WHERE id = $1",
    [id]
  );
  expect(sigueActiva[0].estado, "el estado cambio pese al rechazo").toBe("ACTIVA");

  const baja = await administrador.patch(`/catalogos/minas/${id}/estado`, {
    data: { estado: "INACTIVA" }
  });
  expect(baja.status()).toBe(200);

  const despues = await consultar<{ estado: string }>(
    "SELECT estado FROM catalogos.mina WHERE id = $1",
    [id]
  );
  expect(despues[0].estado).toBe("INACTIVA");
  await administrador.dispose();
});

test("contratar un empleado lo deja ligado a su mina y a su puesto", async () => {
  const administrador = await comoAdministrador();
  const idMina = await primeraOpcion(administrador, "/catalogos/minas?limite=5");
  const idDepartamento = await primeraOpcion(administrador, "/catalogos/departamentos");
  const idPuesto = await primeraOpcion(administrador, "/catalogos/puestos");
  const idActividad = await primeraOpcion(administrador, "/catalogos/actividades");
  const nomina = sello("N").toUpperCase();

  const creado = await administrador.post("/catalogos/empleados", {
    data: {
      id_mina: idMina,
      numero_nomina: nomina,
      nombre_completo: "QA EMPLEADO DE PRUEBA",
      id_departamento: idDepartamento,
      id_puesto: idPuesto,
      id_actividad: idActividad,
      centro_costo: "CC-QA",
      gerente_a_cargo: "QA GERENTE",
      grupo: "A"
    }
  });
  expect(creado.status(), `no se pudo contratar: ${await creado.text()}`).toBe(201);
  const { id } = await creado.json();

  const filas = await consultar<{ numero_nomina: string; id_mina: string; estado: string }>(
    "SELECT numero_nomina, id_mina::text, estado FROM catalogos.empleado WHERE id = $1",
    [id]
  );
  expect(filas[0].numero_nomina).toBe(nomina);
  expect(filas[0].id_mina, "el empleado quedo en otra mina").toBe(idMina);
  expect(filas[0].estado, "un empleado nuevo debe nacer activo").toBe("ACTIVO");

  const repetido = await administrador.post("/catalogos/empleados", {
    data: {
      id_mina: idMina,
      numero_nomina: nomina,
      nombre_completo: "QA EMPLEADO CLONADO",
      id_departamento: idDepartamento,
      id_puesto: idPuesto,
      id_actividad: idActividad,
      centro_costo: "",
      gerente_a_cargo: "",
      grupo: ""
    }
  });
  expect(repetido.status(), "se permitio repetir el numero de nomina").toBeGreaterThanOrEqual(400);
  await administrador.dispose();
});

test("no se puede contratar a un empleado en una mina que no existe", async () => {
  const administrador = await comoAdministrador();
  const idDepartamento = await primeraOpcion(administrador, "/catalogos/departamentos");
  const idPuesto = await primeraOpcion(administrador, "/catalogos/puestos");
  const idActividad = await primeraOpcion(administrador, "/catalogos/actividades");

  const respuesta = await administrador.post("/catalogos/empleados", {
    data: {
      id_mina: "00000000-0000-0000-0000-000000000000",
      numero_nomina: sello("N").toUpperCase(),
      nombre_completo: "QA SIN MINA",
      id_departamento: idDepartamento,
      id_puesto: idPuesto,
      id_actividad: idActividad,
      centro_costo: "",
      gerente_a_cargo: "",
      grupo: ""
    }
  });
  expect(
    respuesta.status(),
    "se acepto un empleado apuntando a una mina inexistente"
  ).toBeGreaterThanOrEqual(400);
  await administrador.dispose();
});

test("dar de alta un equipo lo deja operativo y se puede mandar a mantenimiento", async () => {
  const administrador = await comoAdministrador();
  const idMina = await primeraOpcion(administrador, "/catalogos/minas?limite=5");
  const idTipo = await primeraOpcion(administrador, "/catalogos/tipos-de-equipo");
  const idModulo = await primeraOpcion(administrador, "/catalogos/modulos-de-trabajo");
  const codigo = sello("EQ").toUpperCase();

  const creado = await administrador.post("/catalogos/equipos", {
    data: {
      id_mina: idMina,
      id_tipo_equipo: idTipo,
      id_modulo_trabajo: idModulo,
      codigo,
      descripcion: "Equipo de prueba QA",
      fabricante: "QA MOTORS",
      modelo: "QA-1",
      numero_serie: "",
      anio_fabricacion: null,
      fecha_ingreso_mina: "",
      modelo_perforadora: "",
      capacidad_longitud: null
    }
  });
  expect(creado.status(), `no se pudo dar de alta el equipo: ${await creado.text()}`).toBe(201);
  const { id } = await creado.json();

  const filas = await consultar<{ codigo: string; estado: string; id_mina: string }>(
    "SELECT codigo, estado, id_mina::text FROM catalogos.equipo WHERE id = $1",
    [id]
  );
  expect(filas[0].codigo).toBe(codigo);
  expect(filas[0].estado, "un equipo nuevo debe nacer operativo").toBe("OPERATIVO");
  expect(filas[0].id_mina).toBe(idMina);

  const invalido = await administrador.patch(`/catalogos/equipos/${id}/estado`, {
    data: { estado: "DESCOMPUESTO" }
  });
  expect(invalido.status(), "se acepto un estado de equipo inventado").toBe(400);

  const mantenimiento = await administrador.patch(`/catalogos/equipos/${id}/estado`, {
    data: { estado: "MANTENIMIENTO" }
  });
  expect(mantenimiento.status()).toBe(200);

  const despues = await consultar<{ estado: string }>(
    "SELECT estado FROM catalogos.equipo WHERE id = $1",
    [id]
  );
  expect(despues[0].estado).toBe("MANTENIMIENTO");
  await administrador.dispose();
});

test("el listado pagina de verdad: el cursor trae la pagina siguiente sin repetir", async () => {
  const administrador = await comoAdministrador();
  const primera = await (await administrador.get("/catalogos/minas?limite=1")).json();
  expect(primera.Elementos.length, "se pidio una sola mina por pagina").toBe(1);
  expect(primera.SiguienteCursor, "con mas minas disponibles debe venir un cursor").not.toBe("");

  const segunda = await (
    await administrador.get(`/catalogos/minas?limite=1&cursor=${encodeURIComponent(primera.SiguienteCursor)}`)
  ).json();
  expect(segunda.Elementos.length).toBe(1);
  expect(
    segunda.Elementos[0].Identificador,
    "la segunda pagina repitio el elemento de la primera"
  ).not.toBe(primera.Elementos[0].Identificador);
  await administrador.dispose();
});

test("el detalle de una mina inexistente responde 404", async () => {
  const administrador = await comoAdministrador();
  const respuesta = await administrador.get("/catalogos/minas/00000000-0000-0000-0000-000000000000");
  expect(respuesta.status()).toBe(404);
  await administrador.dispose();
});
