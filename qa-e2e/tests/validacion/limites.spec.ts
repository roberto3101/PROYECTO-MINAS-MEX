import { test, expect, type APIRequestContext } from "@playwright/test";
import {
  comoAdministrador,
  comoSuperadmin,
  exigirBaseDesechable,
  sello,
  contrasenaValida
} from "../soporte/base";
import { consultar } from "../soporte/datos";

test.beforeAll(() => exigirBaseDesechable());

const LARGO = "A".repeat(5000);

async function primeraOpcion(contexto: APIRequestContext, ruta: string): Promise<string> {
  const datos = await (await contexto.get(ruta)).json();
  const lista = Array.isArray(datos) ? datos : (datos.Elementos ?? []);
  return lista[0]?.Identificador ?? "";
}

test("ningun texto libre admite 5000 caracteres", async () => {
  const administrador = await comoAdministrador();
  const idMina = await primeraOpcion(administrador, "/catalogos/minas?limite=5");
  const idDepartamento = await primeraOpcion(administrador, "/catalogos/departamentos");
  const idPuesto = await primeraOpcion(administrador, "/catalogos/puestos");
  const idActividad = await primeraOpcion(administrador, "/catalogos/actividades");
  const idTipo = await primeraOpcion(administrador, "/catalogos/tipos-de-equipo");
  const idModulo = await primeraOpcion(administrador, "/catalogos/modulos-de-trabajo");

  const casos = [
    {
      campo: "nombre de la mina",
      ruta: "/catalogos/minas",
      cuerpo: { nombre: LARGO, area: "Norte", niveles: "" }
    },
    {
      campo: "nombre completo del empleado",
      ruta: "/catalogos/empleados",
      cuerpo: {
        id_mina: idMina,
        numero_nomina: sello("N").toUpperCase(),
        nombre_completo: LARGO,
        id_departamento: idDepartamento,
        id_puesto: idPuesto,
        id_actividad: idActividad,
        centro_costo: "",
        gerente_a_cargo: "",
        grupo: ""
      }
    },
    {
      campo: "grupo del empleado",
      ruta: "/catalogos/empleados",
      cuerpo: {
        id_mina: idMina,
        numero_nomina: sello("G").toUpperCase(),
        nombre_completo: "QA GRUPO LARGO",
        id_departamento: idDepartamento,
        id_puesto: idPuesto,
        id_actividad: idActividad,
        centro_costo: "",
        gerente_a_cargo: "",
        grupo: LARGO
      }
    },
    {
      campo: "descripcion del equipo",
      ruta: "/catalogos/equipos",
      cuerpo: {
        id_mina: idMina,
        id_tipo_equipo: idTipo,
        id_modulo_trabajo: idModulo,
        codigo: sello("EQ").toUpperCase(),
        descripcion: LARGO,
        fabricante: "",
        modelo: "",
        numero_serie: "",
        anio_fabricacion: null,
        fecha_ingreso_mina: "",
        modelo_perforadora: "",
        capacidad_longitud: null
      }
    },
    {
      campo: "codigo del rol",
      ruta: "/gobierno/roles",
      cuerpo: { codigo: LARGO, descripcion: "QA" }
    },
    {
      campo: "nombre del usuario",
      ruta: "/gobierno/usuarios",
      cuerpo: {
        usuario: sello("qa.largo"),
        nombre: LARGO,
        correo: "",
        contrasena: contrasenaValida(),
        id_empleado: ""
      }
    }
  ];

  for (const caso of casos) {
    const respuesta = await administrador.post(caso.ruta, { data: caso.cuerpo });
    expect(
      respuesta.status(),
      `se acepto un texto de 5000 caracteres en ${caso.campo}`
    ).toBe(400);
    expect(
      (await respuesta.json()).error,
      "el rechazo debe decir que campo se paso de largo"
    ).toMatch(/no puede pasar de|caracteres/i);
  }
  await administrador.dispose();
});

test("la ficha de la empresa no admite datos imposibles", async () => {
  const administrador = await comoAdministrador();
  const base = {
    color_primario: "#D9A440",
    zona_horaria: "America/Mexico_City",
    moneda: "MXN",
    identificacion_fiscal: "",
    correo_contacto: "",
    telefono: ""
  };

  const casos = [
    { que: "una zona horaria que no existe", cuerpo: { ...base, zona_horaria: "Marte/Olimpo" } },
    { que: "una moneda inventada", cuerpo: { ...base, moneda: "XYZ" } },
    { que: "un telefono con letras", cuerpo: { ...base, telefono: "no-es-telefono!!" } },
    { que: "una identificacion fiscal larguisima", cuerpo: { ...base, identificacion_fiscal: LARGO } }
  ];

  for (const caso of casos) {
    const respuesta = await administrador.put("/gobierno/empresa", { data: caso.cuerpo });
    expect(respuesta.status(), `se acepto ${caso.que}`).toBe(400);
  }
  await administrador.dispose();
});

test("un equipo no puede ingresar a la mina en una fecha futura", async () => {
  const administrador = await comoAdministrador();
  const idMina = await primeraOpcion(administrador, "/catalogos/minas?limite=5");
  const idTipo = await primeraOpcion(administrador, "/catalogos/tipos-de-equipo");
  const idModulo = await primeraOpcion(administrador, "/catalogos/modulos-de-trabajo");
  const codigo = sello("EQFUT").toUpperCase();

  const respuesta = await administrador.post("/catalogos/equipos", {
    data: {
      id_mina: idMina,
      id_tipo_equipo: idTipo,
      id_modulo_trabajo: idModulo,
      codigo,
      descripcion: "",
      fabricante: "",
      modelo: "",
      numero_serie: "",
      anio_fabricacion: null,
      fecha_ingreso_mina: "2099-01-01",
      modelo_perforadora: "",
      capacidad_longitud: null
    }
  });

  expect(respuesta.status(), "se acepto un equipo que ingresa en el año 2099").toBe(400);

  const creados = await consultar<{ total: string }>(
    "SELECT count(*)::text AS total FROM catalogos.equipo WHERE codigo = $1",
    [codigo]
  );
  expect(Number(creados[0].total), "el equipo con fecha futura quedo guardado").toBe(0);
  await administrador.dispose();
});

test("la razon social de una empresa nueva tiene limite", async () => {
  const superadmin = await comoSuperadmin();
  const respuesta = await superadmin.post("/plataforma/empresas", {
    data: {
      codigo: sello("QA").toUpperCase().slice(0, 12),
      razon_social: LARGO,
      identificacion_fiscal: "",
      correo_contacto: "",
      telefono: "",
      zona_horaria: "America/Mexico_City",
      moneda: "MXN",
      color_primario: "",
      admin: {
        usuario: sello("qa.rz"),
        nombre: "QA",
        correo: "",
        contrasena: contrasenaValida()
      }
    }
  });
  expect(respuesta.status(), "se acepto una razon social de 5000 caracteres").toBe(400);
  await superadmin.dispose();
});
