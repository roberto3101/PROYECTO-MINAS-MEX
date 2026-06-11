const CLAVE_SESION = "sesion.plataforma.minera";
const CLAVE_GRUPOS = "minas.nav.grupos.cerrados";
const CLAVE_RIEL = "minas.nav.riel";

const estado = {
  sesion: null,
  empresa: null,
  catalogoPermisos: [],
  vistaActual: "",
};

const $ = (selector) => document.querySelector(selector);
const elemento = (html) => {
  const plantilla = document.createElement("template");
  plantilla.innerHTML = html.trim();
  return plantilla.content.firstElementChild;
};
const escapar = (texto) =>
  String(texto ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
const formatearFecha = (iso) => {
  if (!iso) return "—";
  const fecha = new Date(iso);
  if (isNaN(fecha)) return iso;
  return fecha.toLocaleString("es-MX", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" });
};
const oVacio = (valor) => (valor === "" || valor == null ? "—" : valor);

async function pedir(ruta, opciones) {
  try {
    return await fetch(ruta, opciones);
  } catch {
    throw new Error("No hay conexión con el servidor. Verifica tu red e intenta de nuevo.");
  }
}

async function llamarApi(ruta, opciones = {}) {
  const cabeceras = { ...opciones.cabeceras };
  if (opciones.cuerpo !== undefined) cabeceras["Content-Type"] = "application/json";
  if (estado.sesion) cabeceras.Authorization = `Bearer ${estado.sesion.token}`;
  const respuesta = await pedir(ruta, {
    method: opciones.metodo ?? "GET",
    headers: cabeceras,
    body: opciones.cuerpo !== undefined ? JSON.stringify(opciones.cuerpo) : undefined,
  });
  return interpretarRespuesta(respuesta);
}

async function llamarApiArchivo(ruta, formulario) {
  const respuesta = await pedir(ruta, {
    method: "POST",
    headers: estado.sesion ? { Authorization: `Bearer ${estado.sesion.token}` } : {},
    body: formulario,
  });
  return interpretarRespuesta(respuesta);
}

async function interpretarRespuesta(respuesta) {
  if (respuesta.status === 401 && estado.sesion) {
    cerrarSesion();
    throw new Error("La sesión expiró. Vuelve a entrar.");
  }
  const texto = await respuesta.text();
  let datos = null;
  try { datos = texto ? JSON.parse(texto) : null; } catch { datos = null; }
  if (!respuesta.ok) {
    const error = new Error(datos?.error ?? `Error ${respuesta.status}`);
    error.estado = respuesta.status;
    error.datos = datos;
    throw error;
  }
  return datos;
}

const tienePermiso = (codigo) => estado.sesion?.permisos?.includes(codigo) ?? false;
const esPlataforma = () => estado.sesion?.ambito === "plataforma";

function avisar(mensaje, tipo = "exito") {
  const toast = elemento(`<div class="toast ${tipo === "exito" ? "" : tipo}">${escapar(mensaje)}</div>`);
  $("#toasts").appendChild(toast);
  setTimeout(() => {
    toast.classList.add("saliendo");
    setTimeout(() => toast.remove(), 320);
  }, 3600);
}

function confirmar(titulo, mensaje, textoAceptar = "Confirmar") {
  return new Promise((resolver) => {
    $("#titulo-confirmacion").textContent = titulo;
    $("#mensaje-confirmacion").textContent = mensaje;
    const aceptar = $("#boton-aceptar-confirmacion");
    aceptar.textContent = textoAceptar;
    $("#capa-confirmacion").hidden = false;
    const cerrar = (resultado) => {
      $("#capa-confirmacion").hidden = true;
      aceptar.onclick = null;
      $("#boton-cancelar-confirmacion").onclick = null;
      resolver(resultado);
    };
    aceptar.onclick = () => cerrar(true);
    $("#boton-cancelar-confirmacion").onclick = () => cerrar(false);
  });
}

function abrirModal(titulo, contenido) {
  $("#titulo-modal").textContent = titulo;
  const cuerpo = $("#cuerpo-modal");
  cuerpo.innerHTML = "";
  cuerpo.appendChild(contenido);
  $("#capa-modal").hidden = false;
  contenido.querySelector("input, select")?.focus();
}
function cerrarModal() { $("#capa-modal").hidden = true; }

function abrirPanel(titulo, contenido) {
  $("#titulo-panel").textContent = titulo;
  const cuerpo = $("#cuerpo-panel");
  cuerpo.innerHTML = "";
  cuerpo.appendChild(contenido);
  $("#fondo-panel").hidden = false;
  $("#panel-detalle").hidden = false;
}
function cerrarPanel() {
  $("#fondo-panel").hidden = true;
  $("#panel-detalle").hidden = true;
}

function campoHtml(campo) {
  const requerido = campo.opcional ? "" : "required";
  const etiqueta = `${escapar(campo.etiqueta)}${campo.opcional ? ' <small>(opcional)</small>' : ""}`;
  if (campo.tipo === "select") {
    return `<label class="campo"><span>${etiqueta}</span>
      <select name="${campo.nombre}" ${requerido}>
        ${campo.opcional || campo.placeholder ? `<option value="">${escapar(campo.placeholder ?? "Sin asignar")}</option>` : ""}
        ${(campo.opciones ?? []).map((opcion) =>
          `<option value="${escapar(opcion.valor)}">${escapar(opcion.texto)}</option>`).join("")}
      </select></label>`;
  }
  const tipo = campo.tipo ?? "text";
  const extras = [
    campo.maxlength ? `maxlength="${campo.maxlength}"` : 'maxlength="120"',
    campo.min != null ? `min="${campo.min}"` : "",
    campo.max != null ? `max="${campo.max}"` : "",
    campo.step ? `step="${campo.step}"` : "",
  ].join(" ");
  return `<label class="campo"><span>${etiqueta}</span>
    <input name="${campo.nombre}" type="${tipo}" placeholder="${escapar(campo.placeholder ?? "")}" ${extras} ${requerido}>
    ${campo.pista ? `<small class="pista">${escapar(campo.pista)}</small>` : ""}
    ${campo.requisitosContrasena ? `<ul class="requisitos" data-para="${campo.nombre}">
      <li data-regla="largo">Al menos 10 caracteres</li>
      <li data-regla="mayuscula">Una mayúscula</li>
      <li data-regla="minuscula">Una minúscula</li>
      <li data-regla="numero">Un número</li>
    </ul>` : ""}
  </label>`;
}

function formularioModal({ titulo, nota, partes, textoEnviar, alEnviar }) {
  const formulario = elemento(`<form novalidate>
    ${nota ? `<p class="nota-formulario">${escapar(nota)}</p>` : ""}
    ${partes.map((parte) => `
      ${parte.seccion ? `<p class="seccion-formulario">${escapar(parte.seccion)}</p>` : ""}
      ${parte.campos.map(campoHtml).join("")}
    `).join("")}
    <div class="pie-modal">
      <button type="button" class="boton-secundario" data-cancelar>Cancelar</button>
      <button type="submit" class="boton-primario">${escapar(textoEnviar)}</button>
    </div>
  </form>`);
  formulario.querySelector("[data-cancelar]").addEventListener("click", cerrarModal);
  formulario.querySelectorAll(".requisitos").forEach((lista) => {
    const entrada = formulario.querySelector(`[name="${lista.dataset.para}"]`);
    entrada.addEventListener("input", () => {
      const valor = entrada.value;
      lista.querySelector('[data-regla="largo"]').classList.toggle("cumplido", valor.length >= 10);
      lista.querySelector('[data-regla="mayuscula"]').classList.toggle("cumplido", /[A-ZÁÉÍÓÚÑ]/.test(valor));
      lista.querySelector('[data-regla="minuscula"]').classList.toggle("cumplido", /[a-záéíóúñ]/.test(valor));
      lista.querySelector('[data-regla="numero"]').classList.toggle("cumplido", /\d/.test(valor));
    });
  });
  formulario.addEventListener("submit", async (evento) => {
    evento.preventDefault();
    if (!formulario.reportValidity()) return;
    const datos = Object.fromEntries(new FormData(formulario).entries());
    const enviar = formulario.querySelector('[type="submit"]');
    enviar.disabled = true;
    enviar.classList.add("ocupado");
    try {
      await alEnviar(datos);
      cerrarModal();
    } catch (excepcion) {
      avisar(excepcion.message, "error");
    } finally {
      enviar.disabled = false;
      enviar.classList.remove("ocupado");
    }
  });
  abrirModal(titulo, formulario);
}

const ICONOS = {
  panel: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="7.5" height="7.5" rx="1.6"/><rect x="13.5" y="3" width="7.5" height="7.5" rx="1.6"/><rect x="3" y="13.5" width="7.5" height="7.5" rx="1.6"/><rect x="13.5" y="13.5" width="7.5" height="7.5" rx="1.6"/></svg>',
  usuarios: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="9" cy="8" r="3.4"/><path d="M3.5 20c.6-3.5 2.8-5.4 5.5-5.4S13.9 16.5 14.5 20"/><circle cx="17" cy="9" r="2.6"/><path d="M15.8 14.8c2.6.2 4.3 1.9 4.7 4.7"/></svg>',
  roles: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 3l7 3v5c0 4.6-3 8.4-7 10-4-1.6-7-5.4-7-10V6z"/><path d="M9.2 12.2l2 2 3.6-4"/></svg>',
  minas: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 20L10 6l4.5 8L17 10l4 10z"/><path d="M3 20h18"/></svg>',
  empleados: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 10h16M6 10V7a2 2 0 012-2h8a2 2 0 012 2v3M6 10l1 10h10l1-10"/><circle cx="12" cy="14.5" r="1.4" fill="currentColor" stroke="none"/></svg>',
  equipos: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 16h3l2-6h6l2 4h4v4h-2"/><circle cx="7.5" cy="18" r="1.8"/><circle cx="16.5" cy="18" r="1.8"/><path d="M9.3 18h5.4"/></svg>',
  empresa: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="3.2"/><path d="M12 2.8v2.6M12 18.6v2.6M2.8 12h2.6M18.6 12h2.6M5.5 5.5l1.9 1.9M16.6 16.6l1.9 1.9M18.5 5.5l-1.9 1.9M7.4 16.6l-1.9 1.9"/></svg>',
  empresas: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 21h18M5 21V7l7-4 7 4v14"/><path d="M9 9h2M13 9h2M9 13h2M13 13h2M9 17h2M13 17h2"/></svg>',
  chevron: '<svg class="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>',
  buscar: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><line x1="21" y1="21" x2="16.5" y2="16.5"/></svg>',
};

const VISTAS = {
  panel: { titulo: "Panel", subtitulo: "Tu sesión y accesos rápidos", icono: "panel", render: vistaPanel },
  usuarios: { titulo: "Usuarios", subtitulo: "Cuentas de acceso de tu empresa", icono: "usuarios", permiso: "usuarios.ver", render: vistaUsuarios },
  roles: { titulo: "Roles y permisos", subtitulo: "La matriz de lo que cada quien puede hacer", icono: "roles", permiso: "roles.ver", render: vistaRoles },
  minas: { titulo: "Minas", subtitulo: "Unidades mineras de la empresa", icono: "minas", permiso: "catalogos.ver", render: vistaMinas },
  empleados: { titulo: "Empleados", subtitulo: "Personal operativo por mina", icono: "empleados", permiso: "catalogos.ver", render: vistaEmpleados },
  equipos: { titulo: "Equipos", subtitulo: "Flota: camiones, scoops, tiro, banda…", icono: "equipos", permiso: "catalogos.ver", render: vistaEquipos },
  empresa: { titulo: "Mi empresa", subtitulo: "Identidad, contacto y logotipo", icono: "empresa", render: vistaEmpresa },
  empresas: { titulo: "Empresas", subtitulo: "Alta y administración de los tenants de la plataforma", icono: "empresas", render: vistaEmpresas },
};

const GRUPOS_EMPRESA = [
  { vistas: ["panel"] },
  { titulo: "Gobierno", vistas: ["usuarios", "roles"] },
  { titulo: "Catálogos", vistas: ["minas", "empleados", "equipos"] },
  { titulo: "Configuración", vistas: ["empresa"] },
];
const GRUPOS_PLATAFORMA = [{ vistas: ["empresas"] }];

function gruposCerrados() {
  try { return new Set(JSON.parse(localStorage.getItem(CLAVE_GRUPOS) ?? "[]")); } catch { return new Set(); }
}

function construirNavegacion() {
  const navegacion = $("#navegacion");
  navegacion.innerHTML = "";
  const cerrados = gruposCerrados();
  const grupos = esPlataforma() ? GRUPOS_PLATAFORMA : GRUPOS_EMPRESA;
  for (const grupo of grupos) {
    const vistasVisibles = grupo.vistas.filter((clave) => {
      const vista = VISTAS[clave];
      return !vista.permiso || tienePermiso(vista.permiso);
    });
    if (!vistasVisibles.length) continue;
    const enlaces = vistasVisibles.map((clave) => {
      const vista = VISTAS[clave];
      return `<button class="enlace-nav" data-vista="${clave}" title="${escapar(vista.titulo)}">${ICONOS[vista.icono]}<span>${escapar(vista.titulo)}</span></button>`;
    }).join("");
    if (!grupo.titulo) {
      navegacion.insertAdjacentHTML("beforeend", enlaces);
      continue;
    }
    const bloque = elemento(`<div class="grupo-nav ${cerrados.has(grupo.titulo) ? "cerrado" : ""}" data-grupo="${escapar(grupo.titulo)}">
      <button class="cabecera-grupo" type="button"><span>${escapar(grupo.titulo)}</span>${ICONOS.chevron}</button>
      <div class="items-grupo"><div>${enlaces}</div></div>
    </div>`);
    bloque.querySelector(".cabecera-grupo").addEventListener("click", () => {
      bloque.classList.toggle("cerrado");
      const actuales = gruposCerrados();
      bloque.classList.contains("cerrado") ? actuales.add(grupo.titulo) : actuales.delete(grupo.titulo);
      localStorage.setItem(CLAVE_GRUPOS, JSON.stringify([...actuales]));
    });
    navegacion.appendChild(bloque);
  }
  navegacion.querySelectorAll(".enlace-nav").forEach((enlace) =>
    enlace.addEventListener("click", () => mostrarVista(enlace.dataset.vista)));
}

function mostrarVista(clave) {
  estado.vistaActual = clave;
  const vista = VISTAS[clave];
  $("#titulo-vista").textContent = vista.titulo;
  $("#subtitulo-vista").textContent = vista.subtitulo;
  document.querySelectorAll(".enlace-nav").forEach((enlace) =>
    enlace.classList.toggle("activo", enlace.dataset.vista === clave));
  $("#acciones-vista").innerHTML = "";
  cerrarPanel();
  vista.render();
}

function botonDeAccion(texto, permiso, manejador) {
  if (permiso && !tienePermiso(permiso)) return;
  const boton = elemento(`<button class="boton-primario">＋ ${escapar(texto)}</button>`);
  boton.addEventListener("click", manejador);
  $("#acciones-vista").appendChild(boton);
}

function pintarEsqueleto(destino, columnas = 4, filas = 5) {
  destino.innerHTML = `<div class="tarjeta sin-relleno"><table class="tabla"><tbody>
    ${Array.from({ length: filas }, () =>
      `<tr class="fila-esqueleto">${Array.from({ length: columnas }, () => '<td><div class="esqueleto"></div></td>').join("")}</tr>`).join("")}
  </tbody></table></div>`;
}

function insigniaDeEstado(valor) {
  const color = { ACTIVO: "verde", ACTIVA: "verde", OPERATIVO: "verde", MANTENIMIENTO: "ambar", INACTIVO: "rojo", INACTIVA: "rojo", BAJA: "rojo" }[valor] ?? "gris";
  return `<span class="insignia ${color}">${escapar(valor)}</span>`;
}

function listadoPaginado({ ruta, columnas, estados, placeholderBusqueda, alClicFila, vacio }) {
  const vista = $("#vista");
  vista.innerHTML = "";
  const barra = elemento(`<div class="barra-herramientas">
    <span class="caja-busqueda">${ICONOS.buscar}<input type="search" placeholder="${escapar(placeholderBusqueda)}" maxlength="80"></span>
    <select class="filtro-estado">${estados.map((opcion) => `<option value="${opcion.valor}">${escapar(opcion.texto)}</option>`).join("")}</select>
  </div>`);
  const contenedor = elemento(`<div></div>`);
  vista.append(barra, contenedor);

  const filtro = { busqueda: "", estado: estados[0].valor, cursor: "" };
  let elementos = [];
  let temporizador = null;

  async function cargar(agregar = false) {
    if (!agregar) {
      filtro.cursor = "";
      elementos = [];
      pintarEsqueleto(contenedor, columnas.length);
    }
    try {
      const parametros = new URLSearchParams({ busqueda: filtro.busqueda, estado: filtro.estado, cursor: filtro.cursor, limite: "20" });
      const datos = await llamarApi(`${ruta}?${parametros}`);
      elementos = elementos.concat(datos.Elementos ?? []);
      filtro.cursor = datos.SiguienteCursor ?? "";
      pintar();
    } catch (excepcion) {
      avisar(excepcion.message, "error");
      contenedor.innerHTML = "";
    }
  }

  function pintar() {
    if (!elementos.length) {
      contenedor.innerHTML = `<div class="estado-vacio"><strong>${escapar(filtro.busqueda ? `Sin resultados para “${filtro.busqueda}”` : vacio.titulo)}</strong>${escapar(filtro.busqueda ? "Prueba con otro término de búsqueda." : vacio.detalle)}</div>`;
      return;
    }
    contenedor.innerHTML = `<div class="tarjeta sin-relleno"><table class="tabla">
      <thead><tr>${columnas.map((columna) => `<th>${escapar(columna.titulo)}</th>`).join("")}</tr></thead>
      <tbody>${elementos.map((fila, indice) =>
        `<tr data-indice="${indice}">${columnas.map((columna) => `<td>${columna.celda(fila)}</td>`).join("")}</tr>`).join("")}
      </tbody></table></div>
      ${filtro.cursor ? '<div class="pie-lista"><button class="boton-cargar">Cargar más</button></div>' : ""}`;
    contenedor.querySelectorAll("tbody tr").forEach((tr) =>
      tr.addEventListener("click", () => alClicFila(elementos[Number(tr.dataset.indice)], () => cargar(false))));
    contenedor.querySelector(".boton-cargar")?.addEventListener("click", () => cargar(true));
  }

  barra.querySelector("input").addEventListener("input", (evento) => {
    clearTimeout(temporizador);
    temporizador = setTimeout(() => {
      filtro.busqueda = evento.target.value.trim();
      cargar(false);
    }, 350);
  });
  barra.querySelector("select").addEventListener("change", (evento) => {
    filtro.estado = evento.target.value;
    cargar(false);
  });

  cargar(false);
  return { recargar: () => cargar(false) };
}

function fichaHtml(pares) {
  return `<dl class="ficha">${pares.map(([etiqueta, valor]) =>
    `<dt>${escapar(etiqueta)}</dt><dd>${valor}</dd>`).join("")}</dl>`;
}

function aplicarMarca(empresa) {
  estado.empresa = empresa;
  if (empresa.ColorPrimario) document.documentElement.style.setProperty("--acento", empresa.ColorPrimario);
  $("#nombre-identidad").textContent = empresa.RazonSocial;
  $("#detalle-identidad").textContent = `${empresa.Codigo} · ${empresa.Moneda}`;
  const logo = $("#logo-empresa");
  const generico = $("#logo-generico");
  if (empresa.LogoUrl) {
    logo.src = empresa.LogoUrl;
    logo.hidden = false;
    generico.hidden = true;
    logo.onerror = () => { logo.hidden = true; generico.hidden = false; };
  } else {
    logo.hidden = true;
    generico.hidden = false;
  }
}

function vistaPanel() {
  const accesos = GRUPOS_EMPRESA.flatMap((grupo) => grupo.vistas)
    .filter((clave) => clave !== "panel")
    .filter((clave) => !VISTAS[clave].permiso || tienePermiso(VISTAS[clave].permiso));
  const porModulo = {};
  for (const permiso of estado.sesion.permisos ?? []) {
    const modulo = permiso.split(".")[0];
    (porModulo[modulo] ??= []).push(permiso);
  }
  $("#vista").innerHTML = `
    <div class="rejilla-metricas">
      ${accesos.map((clave) => `
        <div class="metrica" data-vista="${clave}">
          ${ICONOS[VISTAS[clave].icono]}
          <div class="valor">→</div>
          <div class="nombre">${escapar(VISTAS[clave].titulo)}</div>
        </div>`).join("")}
    </div>
    <div class="tarjeta">
      <h3>Tus permisos vigentes (${estado.sesion.permisos?.length ?? 0})</h3>
      ${Object.entries(porModulo).map(([modulo, permisos]) => `
        <div style="margin-bottom:10px">
          <small style="text-transform:uppercase;letter-spacing:.1em">${escapar(modulo)}</small><br>
          ${permisos.map((permiso) => `<span class="chip acento">${escapar(permiso)}</span>`).join("")}
        </div>`).join("")}
    </div>`;
  document.querySelectorAll(".metrica").forEach((tarjeta) =>
    tarjeta.addEventListener("click", () => mostrarVista(tarjeta.dataset.vista)));
}

function vistaUsuarios() {
  botonDeAccion("Nuevo usuario", "usuarios.crear", abrirAltaDeUsuario);
  listadoPaginado({
    ruta: "/gobierno/usuarios",
    placeholderBusqueda: "Buscar por usuario, nombre o correo…",
    estados: [
      { valor: "ACTIVO", texto: "Activos" },
      { valor: "INACTIVO", texto: "Inactivos" },
      { valor: "TODOS", texto: "Todos" },
    ],
    columnas: [
      { titulo: "Usuario", celda: (usuario) => `<span class="mono">${escapar(usuario.NombreCorto)}</span>` },
      { titulo: "Nombre", celda: (usuario) => escapar(usuario.Nombre) },
      { titulo: "Correo", celda: (usuario) => escapar(oVacio(usuario.Correo)) },
      { titulo: "Último acceso", celda: (usuario) => escapar(usuario.UltimoAccesoEn ? formatearFecha(usuario.UltimoAccesoEn) : "Nunca") },
      { titulo: "Estado", celda: (usuario) => insigniaDeEstado(usuario.Estado) },
    ],
    vacio: { titulo: "Sin usuarios", detalle: "Da de alta al primero con “Nuevo usuario”." },
    alClicFila: (usuario, recargar) => abrirDetalleDeUsuario(usuario.Identificador, recargar),
  });
}

async function abrirDetalleDeUsuario(identificador, recargar) {
  let ficha;
  try { ficha = await llamarApi(`/gobierno/usuarios/${identificador}`); }
  catch (excepcion) { return avisar(excepcion.message, "error"); }
  const usuario = ficha.Usuario;
  const activo = usuario.Estado === "ACTIVO";
  const contenido = elemento(`<div>
    ${fichaHtml([
      ["Nombre de usuario", `<span class="mono">${escapar(usuario.NombreCorto)}</span>`],
      ["Nombre completo", escapar(usuario.Nombre)],
      ["Correo electrónico", escapar(oVacio(usuario.Correo))],
      ["Estado", insigniaDeEstado(usuario.Estado)],
      ["Último acceso", escapar(usuario.UltimoAccesoEn ? formatearFecha(usuario.UltimoAccesoEn) : "Nunca")],
      ["Alta en el sistema", escapar(formatearFecha(usuario.CreadoEn))],
    ])}
    <div class="acciones-panel">
      ${tienePermiso("roles.asignar") ? '<button class="boton-secundario compacto" data-asignar>Asignar rol</button>' : ""}
      ${tienePermiso("usuarios.editar") ? '<button class="boton-secundario compacto" data-editar>Editar</button>' : ""}
      ${tienePermiso("usuarios.desactivar") ? `<button class="boton-secundario compacto ${activo ? "peligro" : ""}" data-estado>${activo ? "Desactivar" : "Activar"}</button>` : ""}
    </div>
    <div class="seccion-panel">
      <h3>Roles asignados</h3>
      ${ficha.Asignaciones?.length ? ficha.Asignaciones.map((asignacion) => `
        <div class="fila-asignacion ${asignacion.Vigente ? "" : "revocada"}">
          <div><strong>${escapar(asignacion.CodigoRol)}</strong>
            <small>${escapar(asignacion.Rol)} · ${asignacion.AlcanceMina ? "alcance por mina" : "toda la empresa"}${asignacion.Vigente ? "" : " · revocada"}</small></div>
          ${asignacion.Vigente && tienePermiso("roles.asignar")
            ? `<button class="boton-secundario compacto peligro" data-revocar="${asignacion.Identificador}">Revocar</button>` : ""}
        </div>`).join("")
        : '<small>Sin roles todavía: este usuario entra sin permisos.</small>'}
    </div>
  </div>`);
  contenido.querySelector("[data-asignar]")?.addEventListener("click", () =>
    abrirAsignacionDeRol(identificador, usuario.NombreCorto, () => abrirDetalleDeUsuario(identificador, recargar)));
  contenido.querySelector("[data-editar]")?.addEventListener("click", () =>
    formularioModal({
      titulo: `Editar a ${usuario.NombreCorto}`,
      partes: [{ campos: [
        { nombre: "nombre", etiqueta: "Nombre completo", placeholder: usuario.Nombre },
        { nombre: "correo", etiqueta: "Correo electrónico", tipo: "email", opcional: true },
      ] }],
      textoEnviar: "Guardar cambios",
      alEnviar: async (datos) => {
        await llamarApi(`/gobierno/usuarios/${identificador}`, { metodo: "PUT", cuerpo: { nombre: datos.nombre, correo: datos.correo } });
        avisar("Usuario actualizado.");
        recargar?.();
        abrirDetalleDeUsuario(identificador, recargar);
      },
    }));
  contenido.querySelector("[data-estado]")?.addEventListener("click", async () => {
    const nuevoEstado = activo ? "INACTIVO" : "ACTIVO";
    const seguro = await confirmar(
      activo ? `¿Desactivar a ${usuario.NombreCorto}?` : `¿Activar a ${usuario.NombreCorto}?`,
      activo ? "Perderá el acceso al sistema de inmediato." : "Recuperará el acceso con sus roles vigentes.",
      activo ? "Desactivar" : "Activar");
    if (!seguro) return;
    try {
      await llamarApi(`/gobierno/usuarios/${identificador}/estado`, { metodo: "PATCH", cuerpo: { estado: nuevoEstado } });
      avisar(`Usuario ${nuevoEstado === "ACTIVO" ? "activado" : "desactivado"}.`);
      recargar?.();
      abrirDetalleDeUsuario(identificador, recargar);
    } catch (excepcion) { avisar(excepcion.message, "error"); }
  });
  contenido.querySelectorAll("[data-revocar]").forEach((boton) =>
    boton.addEventListener("click", async () => {
      const seguro = await confirmar("¿Revocar este rol?", "El usuario perderá esos permisos al instante; la revocación queda registrada.", "Revocar");
      if (!seguro) return;
      try {
        await llamarApi(`/gobierno/asignaciones/${boton.dataset.revocar}`, { metodo: "DELETE" });
        avisar("Rol revocado.");
        abrirDetalleDeUsuario(identificador, recargar);
      } catch (excepcion) { avisar(excepcion.message, "error"); }
    }));
  abrirPanel(usuario.NombreCorto, contenido);
}

async function abrirAltaDeUsuario() {
  const empleados = tienePermiso("catalogos.ver")
    ? await llamarApi("/catalogos/empleados?limite=100").then((datos) => datos.Elementos ?? []).catch(() => [])
    : [];
  formularioModal({
    titulo: "Nuevo usuario",
    nota: "Entra sin roles: asígnale uno desde su detalle para darle permisos.",
    partes: [{ campos: [
      { nombre: "usuario", etiqueta: "Nombre de usuario", placeholder: "op.norte", maxlength: 40, pista: "Minúsculas, números, punto, guion o guion bajo (3 a 40)." },
      { nombre: "nombre", etiqueta: "Nombre completo", placeholder: "JUAN PÉREZ" },
      { nombre: "correo", etiqueta: "Correo electrónico", tipo: "email", opcional: true },
      { nombre: "contrasena", etiqueta: "Contraseña", tipo: "password", requisitosContrasena: true },
      ...(empleados.length ? [{
        nombre: "id_empleado", etiqueta: "Vincular a un empleado", tipo: "select", opcional: true, placeholder: "Sin vínculo",
        opciones: empleados.map((empleado) => ({ valor: empleado.Identificador, texto: `${empleado.NumeroNomina} · ${empleado.NombreCompleto}` })),
      }] : []),
    ] }],
    textoEnviar: "Crear usuario",
    alEnviar: async (datos) => {
      await llamarApi("/gobierno/usuarios", { metodo: "POST", cuerpo: datos });
      avisar(`Usuario ${datos.usuario} creado.`);
      mostrarVista("usuarios");
    },
  });
}

async function abrirAsignacionDeRol(identificadorUsuario, nombreCorto, alTerminar) {
  const [roles, minas] = await Promise.all([
    llamarApi("/gobierno/roles").catch(() => []),
    tienePermiso("catalogos.ver")
      ? llamarApi("/catalogos/minas?limite=100").then((datos) => datos.Elementos ?? []).catch(() => [])
      : [],
  ]);
  formularioModal({
    titulo: `Asignar rol a ${nombreCorto}`,
    nota: "Sin mina elegida, el rol aplica a toda la empresa.",
    partes: [{ campos: [
      { nombre: "id_rol", etiqueta: "Rol", tipo: "select",
        opciones: roles.map((rol) => ({ valor: rol.Identificador, texto: `${rol.Codigo} — ${rol.Descripcion}` })) },
      { nombre: "id_mina", etiqueta: "Alcance", tipo: "select", opcional: true, placeholder: "Toda la empresa",
        opciones: minas.map((mina) => ({ valor: mina.Identificador, texto: mina.Nombre })) },
    ] }],
    textoEnviar: "Asignar rol",
    alEnviar: async (datos) => {
      await llamarApi("/gobierno/asignaciones", { metodo: "POST", cuerpo: { id_usuario: identificadorUsuario, ...datos } });
      avisar(`Rol asignado a ${nombreCorto}.`);
      alTerminar?.();
    },
  });
}

async function vistaRoles() {
  botonDeAccion("Nuevo rol", "roles.crear", () =>
    formularioModal({
      titulo: "Nuevo rol propio",
      nota: "Los roles de sistema no se modifican: para personalizar, crea uno propio y concédele permisos.",
      partes: [{ campos: [
        { nombre: "codigo", etiqueta: "Código del rol", placeholder: "SUPERVISOR_PATIO", maxlength: 40 },
        { nombre: "descripcion", etiqueta: "Descripción", placeholder: "Supervisa el patio de acarreo" },
      ] }],
      textoEnviar: "Crear rol",
      alEnviar: async (datos) => {
        await llamarApi("/gobierno/roles", { metodo: "POST", cuerpo: datos });
        avisar(`Rol ${datos.codigo} creado.`);
        mostrarVista("roles");
      },
    }));
  pintarEsqueleto($("#vista"), 3);
  const roles = await llamarApi("/gobierno/roles").catch((excepcion) => { avisar(excepcion.message, "error"); return []; });
  $("#vista").innerHTML = `<div class="rejilla-roles">
    ${roles.map((rol) => `
      <div class="tarjeta tarjeta-rol" data-id="${rol.Identificador}" data-codigo="${escapar(rol.Codigo)}">
        <div class="fila-titulo">
          <span class="codigo">${escapar(rol.Codigo)}</span>
          <span class="insignia ${rol.EsDeSistema ? "dorado" : "gris"}">${rol.EsDeSistema ? "SISTEMA" : "PROPIO"}</span>
        </div>
        <div class="descripcion">${escapar(rol.Descripcion)}</div>
        <div class="permisos">${rol.Permisos?.length
          ? rol.Permisos.map((permiso) => `<span class="chip">${escapar(permiso)}</span>`).join("")
          : "<small>Sin permisos aún.</small>"}</div>
        <footer>
          <small>${rol.Permisos?.length ?? 0} permisos</small>
          ${!rol.EsDeSistema && tienePermiso("roles.editar")
            ? '<button class="boton-secundario compacto" data-conceder>Conceder permiso</button>'
            : rol.EsDeSistema ? "<small>protegido por la base</small>" : ""}
        </footer>
      </div>`).join("")}
  </div>`;
  $("#vista").querySelectorAll("[data-conceder]").forEach((boton) =>
    boton.addEventListener("click", async () => {
      const tarjeta = boton.closest(".tarjeta-rol");
      if (!estado.catalogoPermisos.length) estado.catalogoPermisos = await llamarApi("/gobierno/permisos");
      const porModulo = {};
      for (const permiso of estado.catalogoPermisos) (porModulo[permiso.Modulo] ??= []).push(permiso);
      const selector = `<label class="campo"><span>Permiso del catálogo</span>
        <select name="permiso" required>
          ${Object.entries(porModulo).map(([modulo, permisos]) => `<optgroup label="${escapar(modulo)}">
            ${permisos.map((permiso) => `<option value="${escapar(permiso.Codigo)}">${escapar(permiso.Codigo)} — ${escapar(permiso.Descripcion)}</option>`).join("")}
          </optgroup>`).join("")}
        </select></label>`;
      const formulario = elemento(`<form novalidate>${selector}
        <div class="pie-modal">
          <button type="button" class="boton-secundario" data-cancelar>Cancelar</button>
          <button type="submit" class="boton-primario">Conceder</button>
        </div></form>`);
      formulario.querySelector("[data-cancelar]").addEventListener("click", cerrarModal);
      formulario.addEventListener("submit", async (evento) => {
        evento.preventDefault();
        try {
          await llamarApi(`/gobierno/roles/${tarjeta.dataset.id}/permisos`, { metodo: "POST", cuerpo: { permiso: formulario.permiso.value } });
          avisar(`Permiso concedido a ${tarjeta.dataset.codigo}.`);
          cerrarModal();
          mostrarVista("roles");
        } catch (excepcion) { avisar(excepcion.message, "error"); }
      });
      abrirModal(`Conceder permiso a ${tarjeta.dataset.codigo}`, formulario);
    }));
}

function numeroONulo(texto) {
  return texto === "" || texto == null ? null : Number(texto);
}

function vistaMinas() {
  botonDeAccion("Nueva mina", "catalogos.editar", () =>
    formularioModal({
      titulo: "Nueva mina",
      partes: [
        { campos: [
          { nombre: "nombre", etiqueta: "Nombre", placeholder: "Mina Norte" },
          { nombre: "area", etiqueta: "Área / zona", placeholder: "Norte" },
          { nombre: "niveles", etiqueta: "Niveles", opcional: true, placeholder: "1900-2400" },
        ] },
        { seccion: "Parámetros técnicos (opcionales)", campos: [
          { nombre: "densidad_promedio", etiqueta: "Densidad promedio (t/m³)", tipo: "number", step: "0.001", min: 0, opcional: true },
          { nombre: "seccion_obra_largo", etiqueta: "Sección de obra — largo (m)", tipo: "number", step: "0.01", min: 0, opcional: true },
          { nombre: "seccion_obra_ancho", etiqueta: "Sección de obra — ancho (m)", tipo: "number", step: "0.01", min: 0, opcional: true },
          { nombre: "seccion_veta_largo", etiqueta: "Sección de veta — largo (m)", tipo: "number", step: "0.01", min: 0, opcional: true },
          { nombre: "seccion_veta_ancho", etiqueta: "Sección de veta — ancho (m)", tipo: "number", step: "0.01", min: 0, opcional: true },
        ] },
      ],
      textoEnviar: "Crear mina",
      alEnviar: async (datos) => {
        await llamarApi("/catalogos/minas", { metodo: "POST", cuerpo: {
          nombre: datos.nombre, area: datos.area, niveles: datos.niveles,
          densidad_promedio: numeroONulo(datos.densidad_promedio),
          seccion_obra_largo: numeroONulo(datos.seccion_obra_largo),
          seccion_obra_ancho: numeroONulo(datos.seccion_obra_ancho),
          seccion_veta_largo: numeroONulo(datos.seccion_veta_largo),
          seccion_veta_ancho: numeroONulo(datos.seccion_veta_ancho),
        } });
        avisar(`Mina ${datos.nombre} creada.`);
        mostrarVista("minas");
      },
    }));
  listadoPaginado({
    ruta: "/catalogos/minas",
    placeholderBusqueda: "Buscar por nombre o área…",
    estados: [
      { valor: "ACTIVA", texto: "Activas" },
      { valor: "INACTIVA", texto: "Inactivas" },
      { valor: "TODAS", texto: "Todas" },
    ],
    columnas: [
      { titulo: "Mina", celda: (mina) => `<strong>${escapar(mina.Nombre)}</strong>` },
      { titulo: "Área", celda: (mina) => escapar(mina.Area) },
      { titulo: "Niveles", celda: (mina) => escapar(oVacio(mina.Niveles)) },
      { titulo: "Estado", celda: (mina) => insigniaDeEstado(mina.Estado) },
    ],
    vacio: { titulo: "Sin minas", detalle: "Crea la primera unidad minera." },
    alClicFila: async (mina, recargar) => {
      let detalle;
      try { detalle = await llamarApi(`/catalogos/minas/${mina.Identificador}`); }
      catch (excepcion) { return avisar(excepcion.message, "error"); }
      const activa = detalle.Estado === "ACTIVA";
      const contenido = elemento(`<div>
        ${fichaHtml([
          ["Nombre", `<strong>${escapar(detalle.Nombre)}</strong>`],
          ["Área / zona", escapar(detalle.Area)],
          ["Niveles", escapar(oVacio(detalle.Niveles))],
          ["Densidad promedio", escapar(detalle.DensidadPromedio ? `${detalle.DensidadPromedio} t/m³` : "—")],
          ["Sección de obra", escapar(detalle.SeccionObraLargo ? `${detalle.SeccionObraLargo} × ${detalle.SeccionObraAncho} m` : "—")],
          ["Sección de veta", escapar(detalle.SeccionVetaLargo ? `${detalle.SeccionVetaLargo} × ${detalle.SeccionVetaAncho} m` : "—")],
          ["Estado", insigniaDeEstado(detalle.Estado)],
          ["Alta en el sistema", escapar(formatearFecha(detalle.CreadoEn))],
        ])}
        ${tienePermiso("catalogos.editar") ? `<div class="acciones-panel">
          <button class="boton-secundario compacto ${activa ? "peligro" : ""}" data-estado>${activa ? "Desactivar" : "Activar"}</button>
        </div>` : ""}
      </div>`);
      contenido.querySelector("[data-estado]")?.addEventListener("click", async () => {
        const nuevo = activa ? "INACTIVA" : "ACTIVA";
        const seguro = await confirmar(`¿${activa ? "Desactivar" : "Activar"} ${detalle.Nombre}?`,
          activa ? "Dejará de aparecer en los catálogos activos." : "Volverá a estar disponible.", activa ? "Desactivar" : "Activar");
        if (!seguro) return;
        try {
          await llamarApi(`/catalogos/minas/${mina.Identificador}/estado`, { metodo: "PATCH", cuerpo: { estado: nuevo } });
          avisar(`Mina ${nuevo === "ACTIVA" ? "activada" : "desactivada"}.`);
          cerrarPanel();
          recargar();
        } catch (excepcion) { avisar(excepcion.message, "error"); }
      });
      abrirPanel(detalle.Nombre, contenido);
    },
  });
}

async function cargarOpciones(ruta) {
  return llamarApi(ruta).catch(() => []);
}

async function vistaEmpleados() {
  const [minas, departamentos, puestos, actividades] = await Promise.all([
    llamarApi("/catalogos/minas?limite=100").then((datos) => datos.Elementos ?? []).catch(() => []),
    cargarOpciones("/catalogos/departamentos"),
    cargarOpciones("/catalogos/puestos"),
    cargarOpciones("/catalogos/actividades"),
  ]);
  botonDeAccion("Nuevo empleado", "catalogos.editar", () =>
    formularioModal({
      titulo: "Nuevo empleado",
      partes: [
        { campos: [
          { nombre: "numero_nomina", etiqueta: "Número de nómina", placeholder: "N-0042", maxlength: 30 },
          { nombre: "nombre_completo", etiqueta: "Nombre completo", placeholder: "JUAN PÉREZ" },
          { nombre: "id_mina", etiqueta: "Mina", tipo: "select",
            opciones: minas.map((mina) => ({ valor: mina.Identificador, texto: mina.Nombre })) },
        ] },
        { seccion: "Organización (opcional)", campos: [
          { nombre: "id_departamento", etiqueta: "Departamento", tipo: "select", opcional: true,
            opciones: departamentos.map((opcion) => ({ valor: opcion.Identificador, texto: opcion.Descripcion })) },
          { nombre: "id_puesto", etiqueta: "Puesto", tipo: "select", opcional: true,
            opciones: puestos.map((opcion) => ({ valor: opcion.Identificador, texto: opcion.Descripcion })) },
          { nombre: "id_actividad", etiqueta: "Actividad principal", tipo: "select", opcional: true,
            opciones: actividades.map((opcion) => ({ valor: opcion.Identificador, texto: opcion.Descripcion })) },
          { nombre: "centro_costo", etiqueta: "Centro de costo", opcional: true, placeholder: "007-0001" },
          { nombre: "gerente_a_cargo", etiqueta: "Jefe directo", opcional: true, placeholder: "ING. RAMÍREZ" },
          { nombre: "grupo", etiqueta: "Cuadrilla / grupo", opcional: true, placeholder: "A" },
        ] },
      ],
      textoEnviar: "Dar de alta",
      alEnviar: async (datos) => {
        await llamarApi("/catalogos/empleados", { metodo: "POST", cuerpo: datos });
        avisar(`Empleado ${datos.numero_nomina} dado de alta.`);
        mostrarVista("empleados");
      },
    }));
  listadoPaginado({
    ruta: "/catalogos/empleados",
    placeholderBusqueda: "Buscar por nómina o nombre…",
    estados: [
      { valor: "ACTIVO", texto: "Activos" },
      { valor: "INACTIVO", texto: "Inactivos" },
      { valor: "TODOS", texto: "Todos" },
    ],
    columnas: [
      { titulo: "Nómina", celda: (empleado) => `<span class="mono">${escapar(empleado.NumeroNomina)}</span>` },
      { titulo: "Nombre", celda: (empleado) => escapar(empleado.NombreCompleto) },
      { titulo: "Mina", celda: (empleado) => escapar(empleado.Mina) },
      { titulo: "Departamento", celda: (empleado) => escapar(oVacio(empleado.Departamento)) },
      { titulo: "Puesto", celda: (empleado) => escapar(oVacio(empleado.Puesto)) },
      { titulo: "Estado", celda: (empleado) => insigniaDeEstado(empleado.Estado) },
    ],
    vacio: { titulo: "Sin empleados", detalle: "Da de alta al personal operativo." },
    alClicFila: async (empleado, recargar) => {
      let detalle;
      try { detalle = await llamarApi(`/catalogos/empleados/${empleado.Identificador}`); }
      catch (excepcion) { return avisar(excepcion.message, "error"); }
      const activo = detalle.Estado === "ACTIVO";
      const contenido = elemento(`<div>
        ${fichaHtml([
          ["Número de nómina", `<span class="mono">${escapar(detalle.NumeroNomina)}</span>`],
          ["Nombre completo", escapar(detalle.NombreCompleto)],
          ["Mina", escapar(detalle.Mina)],
          ["Departamento", escapar(oVacio(detalle.Departamento))],
          ["Puesto", escapar(oVacio(detalle.Puesto))],
          ["Actividad principal", escapar(oVacio(detalle.Actividad))],
          ["Centro de costo", escapar(oVacio(detalle.CentroCosto))],
          ["Jefe directo", escapar(oVacio(detalle.GerenteACargo))],
          ["Cuadrilla / grupo", escapar(oVacio(detalle.Grupo))],
          ["Estado", insigniaDeEstado(detalle.Estado)],
          ["Alta en el sistema", escapar(formatearFecha(detalle.CreadoEn))],
        ])}
        ${tienePermiso("catalogos.editar") ? `<div class="acciones-panel">
          <button class="boton-secundario compacto ${activo ? "peligro" : ""}" data-estado>${activo ? "Desactivar" : "Activar"}</button>
        </div>` : ""}
      </div>`);
      contenido.querySelector("[data-estado]")?.addEventListener("click", async () => {
        const nuevo = activo ? "INACTIVO" : "ACTIVO";
        const seguro = await confirmar(`¿${activo ? "Desactivar" : "Activar"} a ${detalle.NombreCompleto}?`,
          activo ? "Dejará de aparecer entre el personal activo." : "Volverá al personal activo.", activo ? "Desactivar" : "Activar");
        if (!seguro) return;
        try {
          await llamarApi(`/catalogos/empleados/${empleado.Identificador}/estado`, { metodo: "PATCH", cuerpo: { estado: nuevo } });
          avisar("Estado actualizado.");
          cerrarPanel();
          recargar();
        } catch (excepcion) { avisar(excepcion.message, "error"); }
      });
      abrirPanel(detalle.NombreCompleto, contenido);
    },
  });
}

async function vistaEquipos() {
  const [minas, tipos, modulos] = await Promise.all([
    llamarApi("/catalogos/minas?limite=100").then((datos) => datos.Elementos ?? []).catch(() => []),
    cargarOpciones("/catalogos/tipos-de-equipo"),
    cargarOpciones("/catalogos/modulos-de-trabajo"),
  ]);
  botonDeAccion("Nuevo equipo", "catalogos.editar", () =>
    formularioModal({
      titulo: "Nuevo equipo",
      nota: "El tiro (shaft) y la banda de extracción también son equipos de acarreo.",
      partes: [
        { campos: [
          { nombre: "codigo", etiqueta: "Código", placeholder: "CAM-07", maxlength: 30 },
          { nombre: "id_mina", etiqueta: "Mina", tipo: "select",
            opciones: minas.map((mina) => ({ valor: mina.Identificador, texto: mina.Nombre })) },
          { nombre: "id_tipo_equipo", etiqueta: "Tipo de equipo", tipo: "select",
            opciones: tipos.map((tipo) => ({ valor: tipo.Identificador, texto: `${tipo.Codigo} — ${tipo.Descripcion}` })) },
          { nombre: "id_modulo_trabajo", etiqueta: "Módulo de trabajo", tipo: "select",
            opciones: modulos.map((modulo) => ({ valor: modulo.Identificador, texto: `${modulo.Codigo} — ${modulo.Descripcion}` })) },
        ] },
        { seccion: "Datos del fabricante (opcional)", campos: [
          { nombre: "fabricante", etiqueta: "Fabricante", opcional: true, placeholder: "Sandvik" },
          { nombre: "modelo", etiqueta: "Modelo", opcional: true, placeholder: "TH663" },
          { nombre: "numero_serie", etiqueta: "Número de serie", opcional: true },
          { nombre: "anio_fabricacion", etiqueta: "Año de fabricación", tipo: "number", min: 1900, max: 2100, opcional: true },
          { nombre: "fecha_ingreso_mina", etiqueta: "Fecha de llegada a la mina", tipo: "date", opcional: true },
          { nombre: "modelo_perforadora", etiqueta: "Modelo de perforadora", opcional: true },
          { nombre: "capacidad_longitud", etiqueta: "Capacidad / longitud (m)", tipo: "number", step: "0.01", min: 0, opcional: true },
          { nombre: "descripcion", etiqueta: "Descripción", opcional: true },
        ] },
      ],
      textoEnviar: "Dar de alta",
      alEnviar: async (datos) => {
        await llamarApi("/catalogos/equipos", { metodo: "POST", cuerpo: {
          ...datos,
          anio_fabricacion: datos.anio_fabricacion === "" ? null : Number(datos.anio_fabricacion),
          capacidad_longitud: numeroONulo(datos.capacidad_longitud),
        } });
        avisar(`Equipo ${datos.codigo} dado de alta.`);
        mostrarVista("equipos");
      },
    }));
  listadoPaginado({
    ruta: "/catalogos/equipos",
    placeholderBusqueda: "Buscar por código, fabricante o modelo…",
    estados: [
      { valor: "TODOS", texto: "Todos" },
      { valor: "OPERATIVO", texto: "Operativos" },
      { valor: "MANTENIMIENTO", texto: "En mantenimiento" },
      { valor: "BAJA", texto: "De baja" },
    ],
    columnas: [
      { titulo: "Código", celda: (equipo) => `<strong class="mono">${escapar(equipo.Codigo)}</strong>` },
      { titulo: "Tipo", celda: (equipo) => escapar(equipo.Tipo) },
      { titulo: "Módulo", celda: (equipo) => `<span class="chip">${escapar(equipo.Modulo)}</span>` },
      { titulo: "Mina", celda: (equipo) => escapar(equipo.Mina) },
      { titulo: "Fabricante", celda: (equipo) => escapar(oVacio(equipo.Fabricante)) },
      { titulo: "Estado", celda: (equipo) => insigniaDeEstado(equipo.Estado) },
    ],
    vacio: { titulo: "Sin equipos", detalle: "Registra la flota de la operación." },
    alClicFila: async (equipo, recargar) => {
      let detalle;
      try { detalle = await llamarApi(`/catalogos/equipos/${equipo.Identificador}`); }
      catch (excepcion) { return avisar(excepcion.message, "error"); }
      const contenido = elemento(`<div>
        ${fichaHtml([
          ["Código", `<strong class="mono">${escapar(detalle.Codigo)}</strong>`],
          ["Tipo de equipo", escapar(detalle.Tipo)],
          ["Módulo de trabajo", escapar(detalle.Modulo)],
          ["Mina", escapar(detalle.Mina)],
          ["Fabricante", escapar(oVacio(detalle.Fabricante))],
          ["Modelo", escapar(oVacio(detalle.Modelo))],
          ["Número de serie", escapar(oVacio(detalle.NumeroSerie))],
          ["Año de fabricación", escapar(detalle.AnioFabricacion || "—")],
          ["Llegada a la mina", escapar(oVacio(detalle.FechaIngresoMina))],
          ["Modelo de perforadora", escapar(oVacio(detalle.ModeloPerforadora))],
          ["Capacidad / longitud", escapar(detalle.CapacidadLongitud ? `${detalle.CapacidadLongitud} m` : "—")],
          ["Descripción", escapar(oVacio(detalle.Descripcion))],
          ["Estado", insigniaDeEstado(detalle.Estado)],
          ["Alta en el sistema", escapar(formatearFecha(detalle.CreadoEn))],
        ])}
        ${tienePermiso("catalogos.editar") ? `<div class="acciones-panel">
          <label class="campo" style="flex:1;margin:0"><span>Cambiar estado</span>
            <select data-estado>
              ${["OPERATIVO", "MANTENIMIENTO", "BAJA"].map((opcion) =>
                `<option ${opcion === detalle.Estado ? "selected" : ""}>${opcion}</option>`).join("")}
            </select></label>
          <button class="boton-primario" data-aplicar style="align-self:flex-end">Aplicar</button>
        </div>` : ""}
      </div>`);
      contenido.querySelector("[data-aplicar]")?.addEventListener("click", async () => {
        const nuevo = contenido.querySelector("[data-estado]").value;
        if (nuevo === detalle.Estado) return cerrarPanel();
        try {
          await llamarApi(`/catalogos/equipos/${equipo.Identificador}/estado`, { metodo: "PATCH", cuerpo: { estado: nuevo } });
          avisar(`Equipo en estado ${nuevo}.`);
          cerrarPanel();
          recargar();
        } catch (excepcion) { avisar(excepcion.message, "error"); }
      });
      abrirPanel(detalle.Codigo, contenido);
    },
  });
}

async function vistaEmpresa() {
  let empresa;
  try { empresa = await llamarApi("/gobierno/empresa"); }
  catch (excepcion) { return avisar(excepcion.message, "error"); }
  aplicarMarca(empresa);
  const puedeConfigurar = tienePermiso("empresa.configurar");
  const vista = $("#vista");
  vista.innerHTML = "";
  const contenido = elemento(`<div>
    <div class="tarjeta">
      <h3>Logotipo</h3>
      <div class="zona-logo">
        <img class="vista-previa-logo" id="previa-logo" src="${escapar(empresa.LogoUrl || "")}" alt="">
        <div class="datos-archivo">
          <strong id="nombre-archivo">${empresa.LogoUrl ? "Logo actual" : "Sin logotipo"}</strong>
          <small>PNG, JPG o WebP · máximo 2 MB. El servidor analiza el archivo real (tipo, dimensiones y contenido) antes de aceptarlo.</small>
        </div>
        ${puedeConfigurar ? `
          <input type="file" id="archivo-logo" accept=".png,.jpg,.jpeg,.webp" hidden>
          <button class="boton-secundario" id="elegir-logo" type="button">Elegir imagen</button>
          <button class="boton-primario" id="subir-logo" type="button" disabled>Subir</button>` : ""}
      </div>
    </div>
    <form class="tarjeta" id="formulario-empresa">
      <h3>Identidad y contacto</h3>
      <div class="fila-doble">
        <label class="campo"><span>Identificación fiscal (RFC)</span>
          <input name="identificacion_fiscal" value="${escapar(empresa.IdentificacionFiscal)}" maxlength="20" ${puedeConfigurar ? "" : "disabled"}></label>
        <label class="campo"><span>Teléfono</span>
          <input name="telefono" value="${escapar(empresa.Telefono)}" maxlength="20" ${puedeConfigurar ? "" : "disabled"}></label>
      </div>
      <label class="campo"><span>Correo de contacto</span>
        <input name="correo_contacto" type="email" value="${escapar(empresa.CorreoContacto)}" ${puedeConfigurar ? "" : "disabled"}></label>
      <div class="fila-doble">
        <label class="campo"><span>Zona horaria</span>
          <input name="zona_horaria" value="${escapar(empresa.ZonaHoraria)}" ${puedeConfigurar ? "" : "disabled"}></label>
        <label class="campo"><span>Moneda</span>
          <select name="moneda" ${puedeConfigurar ? "" : "disabled"}>
            <option ${empresa.Moneda === "USD" ? "selected" : ""}>USD</option>
            <option ${empresa.Moneda === "MXN" ? "selected" : ""}>MXN</option>
          </select></label>
      </div>
      <label class="campo"><span>Color primario <small>(pinta toda la interfaz)</small></span>
        <span class="fila-color">
          <input type="color" id="selector-color" value="${empresa.ColorPrimario || "#d9a440"}" ${puedeConfigurar ? "" : "disabled"}>
          <input name="color_primario" id="texto-color" value="${escapar(empresa.ColorPrimario)}" placeholder="#D9A440" maxlength="7" ${puedeConfigurar ? "" : "disabled"}>
        </span></label>
      ${puedeConfigurar
        ? '<div class="pie-modal"><button type="submit" class="boton-primario">Guardar y repintar</button></div>'
        : '<p class="nota-formulario">Solo quien tenga el permiso de configurar la empresa puede editar esta sección.</p>'}
    </form>
  </div>`);
  vista.appendChild(contenido);

  const selectorColor = $("#selector-color");
  const textoColor = $("#texto-color");
  selectorColor?.addEventListener("input", () => { textoColor.value = selectorColor.value.toUpperCase(); });
  textoColor?.addEventListener("input", () => {
    if (/^#[0-9a-fA-F]{6}$/.test(textoColor.value)) selectorColor.value = textoColor.value;
  });

  const archivo = $("#archivo-logo");
  const subir = $("#subir-logo");
  $("#elegir-logo")?.addEventListener("click", () => archivo.click());
  archivo?.addEventListener("change", () => {
    const elegido = archivo.files[0];
    if (!elegido) return;
    const tiposValidos = ["image/png", "image/jpeg", "image/webp"];
    if (!tiposValidos.includes(elegido.type)) {
      avisar("Solo se aceptan imágenes PNG, JPG o WebP.", "error");
      archivo.value = "";
      return;
    }
    if (elegido.size > 2 * 1024 * 1024) {
      avisar("La imagen supera el máximo de 2 MB.", "error");
      archivo.value = "";
      return;
    }
    $("#nombre-archivo").textContent = elegido.name;
    $("#previa-logo").src = URL.createObjectURL(elegido);
    subir.disabled = false;
  });
  subir?.addEventListener("click", async () => {
    const elegido = archivo.files[0];
    if (!elegido) return;
    subir.disabled = true;
    subir.classList.add("ocupado");
    try {
      const formulario = new FormData();
      formulario.append("logo", elegido);
      await llamarApiArchivo("/gobierno/empresa/logo", formulario);
      avisar("Logotipo actualizado.");
      vistaEmpresa();
    } catch (excepcion) {
      avisar(excepcion.message, "error");
      subir.disabled = false;
    } finally {
      subir.classList.remove("ocupado");
    }
  });

  $("#formulario-empresa").addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const datos = Object.fromEntries(new FormData(evento.target).entries());
    try {
      await llamarApi("/gobierno/empresa", { metodo: "PUT", cuerpo: datos });
      avisar("Configuración guardada: interfaz repintada.");
      vistaEmpresa();
    } catch (excepcion) { avisar(excepcion.message, "error"); }
  });
}

function vistaEmpresas() {
  botonDeAccion("Nueva empresa", null, () =>
    formularioModal({
      titulo: "Nueva empresa",
      nota: "Al crearla se aprovisionan automáticamente sus roles, su matriz de permisos y los catálogos básicos.",
      partes: [
        { seccion: "Datos de la empresa", campos: [
          { nombre: "codigo", etiqueta: "Código", placeholder: "MIN4", maxlength: 12, pista: "2 a 12 letras mayúsculas o números." },
          { nombre: "razon_social", etiqueta: "Razón social", placeholder: "Minera Cuatro SA de CV" },
          { nombre: "identificacion_fiscal", etiqueta: "Identificación fiscal (RFC)", opcional: true, maxlength: 20 },
          { nombre: "correo_contacto", etiqueta: "Correo de contacto", tipo: "email", opcional: true },
          { nombre: "telefono", etiqueta: "Teléfono", opcional: true, maxlength: 20 },
          { nombre: "zona_horaria", etiqueta: "Zona horaria", opcional: true, placeholder: "America/Mexico_City" },
          { nombre: "moneda", etiqueta: "Moneda", tipo: "select", opciones: [
            { valor: "USD", texto: "USD" }, { valor: "MXN", texto: "MXN" }] },
          { nombre: "color_primario", etiqueta: "Color de marca", opcional: true, placeholder: "#D9A440", maxlength: 7 },
        ] },
        { seccion: "Administrador inicial", campos: [
          { nombre: "admin_usuario", etiqueta: "Nombre de usuario", placeholder: "admin.min4", maxlength: 40 },
          { nombre: "admin_nombre", etiqueta: "Nombre completo", placeholder: "ADMINISTRADOR GENERAL" },
          { nombre: "admin_correo", etiqueta: "Correo electrónico", tipo: "email", opcional: true },
          { nombre: "admin_contrasena", etiqueta: "Contraseña", tipo: "password", requisitosContrasena: true },
        ] },
      ],
      textoEnviar: "Crear y aprovisionar",
      alEnviar: async (datos) => {
        await llamarApi("/plataforma/empresas", { metodo: "POST", cuerpo: {
          codigo: datos.codigo, razon_social: datos.razon_social,
          identificacion_fiscal: datos.identificacion_fiscal, correo_contacto: datos.correo_contacto,
          telefono: datos.telefono, zona_horaria: datos.zona_horaria, moneda: datos.moneda,
          color_primario: datos.color_primario,
          admin: { usuario: datos.admin_usuario, nombre: datos.admin_nombre, correo: datos.admin_correo, contrasena: datos.admin_contrasena },
        } });
        avisar(`Empresa ${datos.codigo} creada con su administrador.`);
        mostrarVista("empresas");
      },
    }));
  listadoPaginado({
    ruta: "/plataforma/empresas",
    placeholderBusqueda: "Buscar por código o razón social…",
    estados: [
      { valor: "TODAS", texto: "Todas" },
      { valor: "ACTIVA", texto: "Activas" },
      { valor: "INACTIVA", texto: "Suspendidas" },
    ],
    columnas: [
      { titulo: "Código", celda: (empresa) => `<strong class="mono">${escapar(empresa.Codigo)}</strong>` },
      { titulo: "Razón social", celda: (empresa) => escapar(empresa.RazonSocial) },
      { titulo: "Usuarios", celda: (empresa) => `<span class="mono">${empresa.TotalUsuarios ?? 0}</span>` },
      { titulo: "Moneda", celda: (empresa) => escapar(empresa.Moneda) },
      { titulo: "Estado", celda: (empresa) => insigniaDeEstado(empresa.Estado) },
    ],
    vacio: { titulo: "Sin empresas", detalle: "Crea el primer tenant de la plataforma." },
    alClicFila: async (empresa, recargar) => {
      let detalle;
      try { detalle = await llamarApi(`/plataforma/empresas/${empresa.Identificador}`); }
      catch (excepcion) { return avisar(excepcion.message, "error"); }
      const activa = detalle.Estado === "ACTIVA";
      const contenido = elemento(`<div>
        ${fichaHtml([
          ["Código", `<strong class="mono">${escapar(detalle.Codigo)}</strong>`],
          ["Razón social", escapar(detalle.RazonSocial)],
          ["Identificación fiscal", escapar(oVacio(detalle.IdentificacionFiscal))],
          ["Correo de contacto", escapar(oVacio(detalle.CorreoContacto))],
          ["Teléfono", escapar(oVacio(detalle.Telefono))],
          ["Zona horaria", escapar(detalle.ZonaHoraria)],
          ["Moneda", escapar(detalle.Moneda)],
          ["Usuarios", `<span class="mono">${detalle.TotalUsuarios}</span>`],
          ["Minas", `<span class="mono">${detalle.TotalMinas}</span>`],
          ["Estado", insigniaDeEstado(detalle.Estado)],
          ["Alta en la plataforma", escapar(formatearFecha(detalle.CreadoEn))],
        ])}
        <div class="acciones-panel">
          <button class="boton-secundario compacto ${activa ? "peligro" : ""}" data-estado>${activa ? "Suspender" : "Reactivar"}</button>
        </div>
      </div>`);
      contenido.querySelector("[data-estado]").addEventListener("click", async () => {
        const nuevo = activa ? "INACTIVA" : "ACTIVA";
        const seguro = await confirmar(
          activa ? `¿Suspender a ${detalle.Codigo}?` : `¿Reactivar a ${detalle.Codigo}?`,
          activa ? "Nadie de esa empresa podrá iniciar sesión mientras esté suspendida." : "Sus usuarios podrán volver a entrar.",
          activa ? "Suspender" : "Reactivar");
        if (!seguro) return;
        try {
          await llamarApi(`/plataforma/empresas/${empresa.Identificador}/estado`, { metodo: "PATCH", cuerpo: { estado: nuevo } });
          avisar(nuevo === "ACTIVA" ? "Empresa reactivada." : "Empresa suspendida.");
          cerrarPanel();
          recargar();
        } catch (excepcion) { avisar(excepcion.message, "error"); }
      });
      abrirPanel(detalle.RazonSocial, contenido);
    },
  });
}

function guardarSesion() { localStorage.setItem(CLAVE_SESION, JSON.stringify(estado.sesion)); }
function cerrarSesion() {
  localStorage.removeItem(CLAVE_SESION);
  estado.sesion = null;
  estado.empresa = null;
  document.documentElement.style.removeProperty("--acento");
  $("#aplicacion").hidden = true;
  $("#pantalla-login").hidden = false;
  cerrarPanel();
  cerrarModal();
}

async function arrancarAplicacion() {
  $("#pantalla-login").hidden = true;
  $("#aplicacion").hidden = false;
  $("#usuario-sesion").textContent = estado.sesion.usuario;
  $("#avatar-sesion").textContent = estado.sesion.usuario.slice(0, 1).toUpperCase();
  $("#aplicacion").classList.toggle("riel", localStorage.getItem(CLAVE_RIEL) === "1");
  construirNavegacion();
  if (esPlataforma()) {
    $("#ambito-sesion").textContent = "Superadmin";
    $("#nombre-identidad").textContent = "Panel de plataforma";
    $("#detalle-identidad").textContent = "Administración de tenants";
    $("#logo-empresa").hidden = true;
    $("#logo-generico").hidden = false;
    mostrarVista("empresas");
    return;
  }
  $("#ambito-sesion").textContent = `${estado.sesion.permisos?.length ?? 0} permisos`;
  try { aplicarMarca(await llamarApi("/gobierno/empresa")); }
  catch (excepcion) { avisar(excepcion.message, "error"); }
  mostrarVista("panel");
}

function prepararLogin() {
  const pestanaEmpresa = $("#pestana-empresa");
  const pestanaPlataforma = $("#pestana-plataforma");
  const formularioEmpresa = $("#formulario-empresa");
  const formularioPlataforma = $("#formulario-plataforma");

  function activarPestana(esEmpresa) {
    pestanaEmpresa.classList.toggle("activo", esEmpresa);
    pestanaPlataforma.classList.toggle("activo", !esEmpresa);
    pestanaEmpresa.setAttribute("aria-selected", esEmpresa);
    pestanaPlataforma.setAttribute("aria-selected", !esEmpresa);
    formularioEmpresa.hidden = !esEmpresa;
    formularioPlataforma.hidden = esEmpresa;
  }
  pestanaEmpresa.addEventListener("click", () => activarPestana(true));
  pestanaPlataforma.addEventListener("click", () => activarPestana(false));

  document.querySelectorAll(".boton-ojo").forEach((boton) => {
    boton.addEventListener("click", () => {
      const entrada = document.getElementById(boton.dataset.objetivo);
      const visible = entrada.type === "text";
      entrada.type = visible ? "password" : "text";
      boton.querySelector(".icono-ojo-abierto").hidden = !visible;
      boton.querySelector(".icono-ojo-cerrado").hidden = visible;
      boton.title = visible ? "Mostrar contraseña" : "Ocultar contraseña";
      entrada.focus();
    });
  });

  $("#demo-empresa").addEventListener("click", () => {
    $("#entrada-codigo-empresa").value = "MIN";
    $("#entrada-usuario-empresa").value = "admin.mina";
    $("#entrada-contrasena-empresa").value = "Mina#2026";
    $("#boton-entrar-empresa").focus();
  });
  $("#demo-plataforma").addEventListener("click", () => {
    $("#entrada-usuario-plataforma").value = "plataforma";
    $("#entrada-contrasena-plataforma").value = "Plataforma#2026";
    $("#boton-entrar-plataforma").focus();
  });

  async function entrar(boton, error, peticion) {
    error.hidden = true;
    boton.disabled = true;
    boton.classList.add("ocupado");
    try {
      estado.sesion = await peticion();
      guardarSesion();
      await arrancarAplicacion();
    } catch (excepcion) {
      const segundos = excepcion.datos?.reintentar_en_segundos;
      error.textContent = segundos ? `Demasiados intentos. Espera ${segundos} segundos.` : excepcion.message;
      error.hidden = false;
    } finally {
      boton.disabled = false;
      boton.classList.remove("ocupado");
    }
  }

  formularioEmpresa.addEventListener("submit", (evento) => {
    evento.preventDefault();
    entrar($("#boton-entrar-empresa"), $("#error-empresa"), () =>
      llamarApi("/sesiones", { metodo: "POST", cuerpo: {
        codigo_empresa: $("#entrada-codigo-empresa").value.trim().toUpperCase(),
        usuario: $("#entrada-usuario-empresa").value.trim(),
        contrasena: $("#entrada-contrasena-empresa").value,
      } }));
  });
  formularioPlataforma.addEventListener("submit", (evento) => {
    evento.preventDefault();
    entrar($("#boton-entrar-plataforma"), $("#error-plataforma"), () =>
      llamarApi("/plataforma/sesiones", { metodo: "POST", cuerpo: {
        usuario: $("#entrada-usuario-plataforma").value.trim(),
        contrasena: $("#entrada-contrasena-plataforma").value,
      } }));
  });
}

prepararLogin();

$("#boton-salir").addEventListener("click", cerrarSesion);
$("#boton-riel").addEventListener("click", () => {
  const aplicacion = $("#aplicacion");
  aplicacion.classList.toggle("riel");
  localStorage.setItem(CLAVE_RIEL, aplicacion.classList.contains("riel") ? "1" : "0");
});
$("#cerrar-modal").addEventListener("click", cerrarModal);
$("#capa-modal").addEventListener("click", (evento) => { if (evento.target.id === "capa-modal") cerrarModal(); });
$("#cerrar-panel").addEventListener("click", cerrarPanel);
$("#fondo-panel").addEventListener("click", cerrarPanel);
document.addEventListener("keydown", (evento) => {
  if (evento.key !== "Escape") return;
  if (!$("#capa-confirmacion").hidden) return $("#boton-cancelar-confirmacion").click();
  if (!$("#capa-modal").hidden) return cerrarModal();
  if (!$("#panel-detalle").hidden) cerrarPanel();
});

const sesionGuardada = localStorage.getItem(CLAVE_SESION);
if (sesionGuardada) {
  try {
    estado.sesion = JSON.parse(sesionGuardada);
    arrancarAplicacion();
  } catch {
    cerrarSesion();
  }
}
