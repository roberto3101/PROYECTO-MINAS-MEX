const CLAVE_SESION = "sesion.plataforma.minera";

const estado = {
  sesion: null,
  empresa: null,
  catalogoPermisos: [],
  vistaActual: "panel",
};

const $ = (selector) => document.querySelector(selector);
const elemento = (html) => {
  const plantilla = document.createElement("template");
  plantilla.innerHTML = html.trim();
  return plantilla.content.firstElementChild;
};
const escapar = (texto) =>
  String(texto ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));

async function llamarApi(ruta, opciones = {}) {
  const cabeceras = { "Content-Type": "application/json", ...opciones.cabeceras };
  if (estado.sesion) cabeceras.Authorization = `Bearer ${estado.sesion.token}`;
  const respuesta = await fetch(ruta, {
    method: opciones.metodo ?? "GET",
    headers: cabeceras,
    body: opciones.cuerpo ? JSON.stringify(opciones.cuerpo) : undefined,
  });
  if (respuesta.status === 401 && estado.sesion) {
    cerrarSesion();
    throw new Error("La sesión expiró. Vuelve a entrar.");
  }
  const texto = await respuesta.text();
  const datos = texto ? JSON.parse(texto) : null;
  if (!respuesta.ok) throw new Error(datos?.error ?? `Error ${respuesta.status}`);
  return datos;
}

const tienePermiso = (codigo) => estado.sesion?.permisos?.includes(codigo) ?? false;

function avisar(mensaje, tipo = "exito") {
  const toast = elemento(`<div class="toast ${tipo === "exito" ? "" : tipo}">${escapar(mensaje)}</div>`);
  $("#toasts").appendChild(toast);
  setTimeout(() => {
    toast.classList.add("saliendo");
    setTimeout(() => toast.remove(), 320);
  }, 3400);
}

function aplicarMarca(empresa) {
  estado.empresa = empresa;
  if (empresa.ColorPrimario) document.documentElement.style.setProperty("--acento", empresa.ColorPrimario);
  $("#nombre-empresa").textContent = empresa.RazonSocial;
  $("#codigo-empresa").textContent = `${empresa.Codigo} · ${empresa.Moneda} · ${empresa.ZonaHoraria.split("/").pop().replaceAll("_", " ")}`;
  const logo = $("#logo-empresa");
  const generico = $("#logo-generico");
  if (empresa.LogoUrl) {
    logo.src = empresa.LogoUrl;
    logo.hidden = false;
    generico.style.display = "none";
    logo.onerror = () => { logo.hidden = true; generico.style.display = ""; };
  } else {
    logo.hidden = true;
    generico.style.display = "";
  }
}

function guardarSesion() { localStorage.setItem(CLAVE_SESION, JSON.stringify(estado.sesion)); }
function cerrarSesion() {
  localStorage.removeItem(CLAVE_SESION);
  estado.sesion = null;
  $("#aplicacion").hidden = true;
  $("#pantalla-login").style.display = "";
  $("#campo-contrasena").value = "";
}

async function entrar(evento) {
  evento.preventDefault();
  const boton = $("#boton-entrar");
  const error = $("#error-login");
  error.hidden = true;
  boton.disabled = true;
  boton.classList.add("ocupado");
  try {
    const datos = await llamarApi("/sesiones", {
      metodo: "POST",
      cuerpo: {
        codigo_empresa: $("#campo-empresa").value.trim().toUpperCase(),
        usuario: $("#campo-usuario").value.trim(),
        contrasena: $("#campo-contrasena").value,
      },
    });
    estado.sesion = datos;
    guardarSesion();
    await arrancarAplicacion();
  } catch (excepcion) {
    error.textContent = excepcion.message;
    error.hidden = false;
  } finally {
    boton.disabled = false;
    boton.classList.remove("ocupado");
  }
}

async function arrancarAplicacion() {
  $("#pantalla-login").style.display = "none";
  $("#aplicacion").hidden = false;
  $("#nombre-sesion").textContent = estado.sesion.usuario;
  $("#avatar-usuario").textContent = estado.sesion.usuario.slice(0, 1).toUpperCase();
  $("#permisos-sesion").textContent = `${estado.sesion.permisos.length} permisos vigentes`;
  construirNavegacion();
  try {
    aplicarMarca(await llamarApi("/gobierno/empresa"));
  } catch (excepcion) {
    avisar(excepcion.message, "error");
  }
  mostrarVista("panel");
}

const ICONOS = {
  panel: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="7.5" height="7.5" rx="1.6"/><rect x="13.5" y="3" width="7.5" height="7.5" rx="1.6"/><rect x="3" y="13.5" width="7.5" height="7.5" rx="1.6"/><rect x="13.5" y="13.5" width="7.5" height="7.5" rx="1.6"/></svg>',
  usuarios: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="9" cy="8" r="3.4"/><path d="M3.5 20c.6-3.5 2.8-5.4 5.5-5.4S13.9 16.5 14.5 20"/><circle cx="17" cy="9" r="2.6"/><path d="M15.8 14.8c2.6.2 4.3 1.9 4.7 4.7"/></svg>',
  roles: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M12 3l7 3v5c0 4.6-3 8.4-7 10-4-1.6-7-5.4-7-10V6z"/><path d="M9.2 12.2l2 2 3.6-4"/></svg>',
  minas: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 20L10 6l4.5 8L17 10l4 10z"/><path d="M3 20h18"/></svg>',
  empleados: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 10h16M6 10V7a2 2 0 012-2h8a2 2 0 012 2v3M6 10l1 10h10l1-10"/><circle cx="12" cy="14.5" r="1.4" fill="currentColor" stroke="none"/></svg>',
  equipos: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 16h3l2-6h6l2 4h4v4h-2"/><circle cx="7.5" cy="18" r="1.8"/><circle cx="16.5" cy="18" r="1.8"/><path d="M9.3 18h5.4"/></svg>',
  empresa: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="3.2"/><path d="M12 2.8v2.6M12 18.6v2.6M2.8 12h2.6M18.6 12h2.6M5.5 5.5l1.9 1.9M16.6 16.6l1.9 1.9M18.5 5.5l-1.9 1.9M7.4 16.6l-1.9 1.9"/></svg>',
};

const VISTAS = [
  { clave: "panel", titulo: "Panel", grupo: null, permiso: null, subtitulo: "Resumen de la operación y tu sesión" },
  { clave: "usuarios", titulo: "Usuarios", grupo: "Gobierno", permiso: "usuarios.ver", subtitulo: "Altas, roles y revocaciones de tu empresa" },
  { clave: "roles", titulo: "Roles y permisos", grupo: "Gobierno", permiso: "roles.ver", subtitulo: "Matriz RBAC: roles de sistema y propios" },
  { clave: "minas", titulo: "Minas", grupo: "Catálogos", permiso: "catalogos.ver", subtitulo: "Unidades mineras de la empresa" },
  { clave: "empleados", titulo: "Empleados", grupo: "Catálogos", permiso: "catalogos.ver", subtitulo: "Personal operativo por mina" },
  { clave: "equipos", titulo: "Equipos", grupo: "Catálogos", permiso: "catalogos.ver", subtitulo: "Flota: camiones, scoops, tiro, banda…" },
  { clave: "empresa", titulo: "Mi empresa", grupo: "Configuración", permiso: null, subtitulo: "Branding e identidad del tenant" },
];

function construirNavegacion() {
  const navegacion = $("#navegacion");
  navegacion.innerHTML = "";
  let grupoPrevio = null;
  for (const vista of VISTAS) {
    if (vista.permiso && !tienePermiso(vista.permiso)) continue;
    if (vista.grupo && vista.grupo !== grupoPrevio) {
      navegacion.appendChild(elemento(`<div class="grupo">${vista.grupo}</div>`));
      grupoPrevio = vista.grupo;
    }
    const enlace = elemento(
      `<button class="enlace-nav" data-vista="${vista.clave}">${ICONOS[vista.clave]}<span>${vista.titulo}</span></button>`
    );
    enlace.addEventListener("click", () => mostrarVista(vista.clave));
    navegacion.appendChild(enlace);
  }
}

function mostrarVista(clave) {
  estado.vistaActual = clave;
  const definicion = VISTAS.find((vista) => vista.clave === clave);
  $("#titulo-vista").textContent = definicion.titulo;
  $("#subtitulo-vista").textContent = definicion.subtitulo;
  document.querySelectorAll(".enlace-nav").forEach((enlace) =>
    enlace.classList.toggle("activo", enlace.dataset.vista === clave)
  );
  $("#acciones-vista").innerHTML = "";
  const render = { panel, usuarios, roles, minas, empleados, equipos, empresa }[clave];
  render();
}

function pintarCargando(columnas = 4, filas = 4) {
  $("#vista").innerHTML = `
    <div class="tarjeta"><table class="tabla"><tbody>
      ${Array.from({ length: filas }, () =>
        `<tr class="fila-esqueleto">${Array.from({ length: columnas }, () => '<td><div class="esqueleto"></div></td>').join("")}</tr>`
      ).join("")}
    </tbody></table></div>`;
}

function estadoVacio(titulo, detalle) {
  return `<div class="estado-vacio">
    ${ICONOS.minas}
    <strong>${escapar(titulo)}</strong>${escapar(detalle)}
  </div>`;
}

function botonDeAccion(texto, permiso, manejador) {
  if (permiso && !tienePermiso(permiso)) return;
  const boton = elemento(`<button class="boton-accion">＋ ${escapar(texto)}</button>`);
  boton.addEventListener("click", manejador);
  $("#acciones-vista").appendChild(boton);
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

function formularioModal({ titulo, nota, campos, textoEnviar, alEnviar }) {
  const formulario = elemento(`<form>
    ${nota ? `<p class="nota-modal">${escapar(nota)}</p>` : ""}
    ${campos.map((campo) => {
      if (campo.tipo === "select") {
        return `<label class="campo"><span>${escapar(campo.etiqueta)}</span>
          <select name="${campo.nombre}" ${campo.opcional ? "" : "required"}>
            ${campo.opcional ? `<option value="">— ${escapar(campo.textoVacio ?? "Sin especificar")} —</option>` : ""}
            ${campo.opciones.map((opcion) => `<option value="${escapar(opcion.valor)}">${escapar(opcion.texto)}</option>`).join("")}
          </select></label>`;
      }
      return `<label class="campo"><span>${escapar(campo.etiqueta)}</span>
        <input name="${campo.nombre}" type="${campo.tipo ?? "text"}" placeholder="${escapar(campo.placeholder ?? "")}" ${campo.opcional ? "" : "required"}></label>`;
    }).join("")}
    <div class="pie-modal">
      <button type="button" class="boton-secundario" data-cancelar>Cancelar</button>
      <button type="submit" class="boton-accion">${escapar(textoEnviar)}</button>
    </div>
  </form>`);
  formulario.querySelector("[data-cancelar]").addEventListener("click", cerrarModal);
  formulario.addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const datos = Object.fromEntries(new FormData(formulario).entries());
    const enviar = formulario.querySelector('[type="submit"]');
    enviar.disabled = true;
    try {
      await alEnviar(datos);
      cerrarModal();
    } catch (excepcion) {
      avisar(excepcion.message, "error");
    } finally {
      enviar.disabled = false;
    }
  });
  abrirModal(titulo, formulario);
}

async function panel() {
  pintarCargando(3, 3);
  const consultas = [
    { nombre: "Usuarios", clave: "usuarios", permiso: "usuarios.ver", ruta: "/gobierno/usuarios", icono: ICONOS.usuarios },
    { nombre: "Roles", clave: "roles", permiso: "roles.ver", ruta: "/gobierno/roles", icono: ICONOS.roles },
    { nombre: "Minas", clave: "minas", permiso: "catalogos.ver", ruta: "/catalogos/minas", icono: ICONOS.minas },
    { nombre: "Empleados", clave: "empleados", permiso: "catalogos.ver", ruta: "/catalogos/empleados", icono: ICONOS.empleados },
    { nombre: "Equipos", clave: "equipos", permiso: "catalogos.ver", ruta: "/catalogos/equipos", icono: ICONOS.equipos },
  ].filter((consulta) => tienePermiso(consulta.permiso));

  const resultados = await Promise.all(
    consultas.map((consulta) => llamarApi(consulta.ruta).catch(() => null))
  );

  const porModulo = {};
  for (const permiso of estado.sesion.permisos) {
    const modulo = permiso.split(".")[0];
    (porModulo[modulo] ??= []).push(permiso);
  }

  $("#vista").innerHTML = `
    <div class="rejilla-metricas">
      ${consultas.map((consulta, indice) => `
        <div class="metrica" data-vista="${consulta.clave}">
          ${consulta.icono}
          <div class="valor">${resultados[indice]?.length ?? "—"}</div>
          <div class="nombre">${consulta.nombre}</div>
        </div>`).join("")}
    </div>
    <div class="tarjeta">
      <h3>Tus permisos vigentes (${estado.sesion.permisos.length})</h3>
      ${Object.entries(porModulo).map(([modulo, permisos]) => `
        <div style="margin-bottom:10px">
          <small style="text-transform:uppercase;letter-spacing:.1em">${escapar(modulo)}</small><br>
          ${permisos.map((permiso) => `<span class="chip acento">${escapar(permiso)}</span>`).join("")}
        </div>`).join("")}
    </div>`;
  document.querySelectorAll(".metrica").forEach((metrica) =>
    metrica.addEventListener("click", () => mostrarVista(metrica.dataset.vista))
  );
}

async function usuarios() {
  botonDeAccion("Nuevo usuario", "usuarios.crear", abrirAltaDeUsuario);
  pintarCargando(5);
  const lista = await llamarApi("/gobierno/usuarios").catch((excepcion) => { avisar(excepcion.message, "error"); return []; });
  if (!lista?.length) { $("#vista").innerHTML = estadoVacio("Sin usuarios", "Da de alta al primero con “Nuevo usuario”."); return; }
  $("#vista").innerHTML = `<div class="tarjeta" style="padding:6px 0"><table class="tabla">
    <thead><tr><th>Usuario</th><th>Nombre</th><th>Correo</th><th>Estado</th><th></th></tr></thead>
    <tbody>${lista.map((usuario) => `
      <tr data-id="${usuario.Identificador}" data-usuario="${escapar(usuario.NombreCorto)}">
        <td class="mono">${escapar(usuario.NombreCorto)}</td>
        <td>${escapar(usuario.Nombre)}</td>
        <td>${escapar(usuario.Correo) || "—"}</td>
        <td><span class="insignia ${usuario.Estado === "ACTIVO" ? "activa" : "inactiva"}">${escapar(usuario.Estado)}</span></td>
        <td class="celda-acciones">
          ${tienePermiso("roles.asignar") ? '<button class="boton-secundario" data-accion="asignar">Asignar rol</button>' : ""}
          ${tienePermiso("roles.ver") ? '<button class="boton-secundario" data-accion="asignaciones">Asignaciones</button>' : ""}
          ${tienePermiso("usuarios.desactivar") && usuario.Estado === "ACTIVO" ? '<button class="boton-secundario" data-accion="desactivar">Desactivar</button>' : ""}
        </td>
      </tr>`).join("")}
    </tbody></table></div>`;

  $("#vista").addEventListener("click", async (evento) => {
    const boton = evento.target.closest("[data-accion]");
    if (!boton) return;
    const fila = boton.closest("tr");
    const identificador = fila.dataset.id;
    const nombreCorto = fila.dataset.usuario;
    if (boton.dataset.accion === "asignar") return abrirAsignacionDeRol(identificador, nombreCorto);
    if (boton.dataset.accion === "asignaciones") return abrirAsignaciones(identificador, nombreCorto);
    if (boton.dataset.accion === "desactivar") {
      if (!confirm(`¿Desactivar a ${nombreCorto}? Perderá el acceso de inmediato.`)) return;
      try {
        await llamarApi(`/gobierno/usuarios/${identificador}`, { metodo: "DELETE" });
        avisar(`${nombreCorto} quedó inactivo.`);
        mostrarVista("usuarios");
      } catch (excepcion) { avisar(excepcion.message, "error"); }
    }
  });
}

async function abrirAltaDeUsuario() {
  const empleados = tienePermiso("catalogos.ver")
    ? await llamarApi("/catalogos/empleados").catch(() => [])
    : [];
  formularioModal({
    titulo: "Nuevo usuario",
    nota: "El usuario entra sin roles: asígnale uno después para que tenga permisos.",
    campos: [
      { nombre: "usuario", etiqueta: "Usuario (login)", placeholder: "op.norte" },
      { nombre: "nombre", etiqueta: "Nombre completo", placeholder: "JUAN PÉREZ" },
      { nombre: "correo", etiqueta: "Correo", tipo: "email", opcional: true },
      { nombre: "contrasena", etiqueta: "Contraseña", tipo: "password", placeholder: "mínimo 8 caracteres" },
      ...(empleados?.length ? [{
        nombre: "id_empleado", etiqueta: "Vincular a empleado", tipo: "select", opcional: true,
        textoVacio: "Sin vínculo",
        opciones: empleados.map((empleado) => ({ valor: empleado.Identificador, texto: `${empleado.NumeroNomina} · ${empleado.NombreCompleto}` })),
      }] : []),
    ],
    textoEnviar: "Crear usuario",
    alEnviar: async (datos) => {
      await llamarApi("/gobierno/usuarios", { metodo: "POST", cuerpo: datos });
      avisar(`Usuario ${datos.usuario} creado.`);
      mostrarVista("usuarios");
    },
  });
}

async function abrirAsignacionDeRol(identificadorUsuario, nombreCorto) {
  const [rolesDisponibles, minasDisponibles] = await Promise.all([
    llamarApi("/gobierno/roles"),
    tienePermiso("catalogos.ver") ? llamarApi("/catalogos/minas").catch(() => []) : [],
  ]);
  formularioModal({
    titulo: `Asignar rol a ${nombreCorto}`,
    nota: "El alcance por mina es opcional: sin mina, el rol aplica a toda la empresa.",
    campos: [
      { nombre: "id_rol", etiqueta: "Rol", tipo: "select",
        opciones: rolesDisponibles.map((rol) => ({ valor: rol.Identificador, texto: `${rol.Codigo} — ${rol.Descripcion}` })) },
      { nombre: "id_mina", etiqueta: "Alcance (mina)", tipo: "select", opcional: true, textoVacio: "Toda la empresa",
        opciones: (minasDisponibles ?? []).map((mina) => ({ valor: mina.Identificador, texto: mina.Nombre })) },
    ],
    textoEnviar: "Asignar",
    alEnviar: async (datos) => {
      await llamarApi("/gobierno/asignaciones", { metodo: "POST", cuerpo: { id_usuario: identificadorUsuario, ...datos } });
      avisar(`Rol asignado a ${nombreCorto}.`);
    },
  });
}

async function abrirAsignaciones(identificadorUsuario, nombreCorto) {
  const asignaciones = await llamarApi(`/gobierno/usuarios/${identificadorUsuario}/asignaciones`).catch((excepcion) => { avisar(excepcion.message, "error"); return []; });
  const contenido = elemento(`<div class="lista-asignaciones">
    ${asignaciones?.length ? asignaciones.map((asignacion) => `
      <div class="fila-asignacion ${asignacion.Vigente ? "" : "revocada"}">
        <div class="detalle">
          <strong>${escapar(asignacion.CodigoRol)}</strong>
          <small>${escapar(asignacion.Rol)} · ${asignacion.AlcanceMina ? "alcance por mina" : "toda la empresa"}${asignacion.Vigente ? "" : " · revocada"}</small>
        </div>
        ${asignacion.Vigente && tienePermiso("roles.asignar")
          ? `<button class="boton-secundario" data-revocar="${asignacion.Identificador}">Revocar</button>` : ""}
      </div>`).join("")
      : `<p style="color:var(--texto-3)">Sin asignaciones todavía.</p>`}
  </div>`);
  contenido.addEventListener("click", async (evento) => {
    const boton = evento.target.closest("[data-revocar]");
    if (!boton) return;
    try {
      await llamarApi(`/gobierno/asignaciones/${boton.dataset.revocar}`, { metodo: "DELETE" });
      avisar(`Rol revocado a ${nombreCorto}. Sus permisos se actualizan al instante.`);
      cerrarModal();
    } catch (excepcion) { avisar(excepcion.message, "error"); }
  });
  abrirModal(`Asignaciones de ${nombreCorto}`, contenido);
}

async function roles() {
  botonDeAccion("Nuevo rol", "roles.crear", abrirCreacionDeRol);
  pintarCargando(3);
  const lista = await llamarApi("/gobierno/roles").catch((excepcion) => { avisar(excepcion.message, "error"); return []; });
  if (!lista?.length) { $("#vista").innerHTML = estadoVacio("Sin roles", "Crea el primero."); return; }
  $("#vista").innerHTML = `<div class="rejilla-roles">
    ${lista.map((rol) => `
      <div class="tarjeta tarjeta-rol" data-id="${rol.Identificador}" data-codigo="${escapar(rol.Codigo)}">
        <div class="fila-titulo">
          <span class="codigo">${escapar(rol.Codigo)}</span>
          <span class="insignia ${rol.EsDeSistema ? "sistema" : "propio"}">${rol.EsDeSistema ? "SISTEMA" : "PROPIO"}</span>
        </div>
        <div class="descripcion">${escapar(rol.Descripcion)}</div>
        <div class="permisos">${rol.Permisos.length
          ? rol.Permisos.map((permiso) => `<span class="chip">${escapar(permiso)}</span>`).join("")
          : '<small>Sin permisos aún.</small>'}</div>
        <footer>
          <small>${rol.Permisos.length} permisos</small>
          ${!rol.EsDeSistema && tienePermiso("roles.editar")
            ? '<button class="boton-secundario" data-conceder>Conceder permiso</button>'
            : rol.EsDeSistema ? '<small>protegido por la base</small>' : ""}
        </footer>
      </div>`).join("")}
  </div>`;
  $("#vista").addEventListener("click", (evento) => {
    const boton = evento.target.closest("[data-conceder]");
    if (!boton) return;
    const tarjeta = boton.closest(".tarjeta-rol");
    abrirConcesionDePermiso(tarjeta.dataset.id, tarjeta.dataset.codigo);
  });
}

function abrirCreacionDeRol() {
  formularioModal({
    titulo: "Nuevo rol propio",
    nota: "Los roles de sistema no se tocan: para personalizar, se crea un rol propio y se le conceden permisos.",
    campos: [
      { nombre: "codigo", etiqueta: "Código", placeholder: "SUPERVISOR_PATIO" },
      { nombre: "descripcion", etiqueta: "Descripción", placeholder: "Supervisa el patio de acarreo" },
    ],
    textoEnviar: "Crear rol",
    alEnviar: async (datos) => {
      await llamarApi("/gobierno/roles", { metodo: "POST", cuerpo: datos });
      avisar(`Rol ${datos.codigo} creado.`);
      mostrarVista("roles");
    },
  });
}

async function abrirConcesionDePermiso(identificadorRol, codigoRol) {
  if (!estado.catalogoPermisos.length) estado.catalogoPermisos = await llamarApi("/gobierno/permisos");
  formularioModal({
    titulo: `Conceder permiso a ${codigoRol}`,
    campos: [{
      nombre: "permiso", etiqueta: "Permiso del catálogo", tipo: "select",
      opciones: estado.catalogoPermisos.map((permiso) => ({ valor: permiso.Codigo, texto: `${permiso.Codigo} — ${permiso.Descripcion}` })),
    }],
    textoEnviar: "Conceder",
    alEnviar: async (datos) => {
      await llamarApi(`/gobierno/roles/${identificadorRol}/permisos`, { metodo: "POST", cuerpo: datos });
      avisar(`Permiso concedido a ${codigoRol}.`);
      mostrarVista("roles");
    },
  });
}

async function minas() {
  botonDeAccion("Nueva mina", "catalogos.editar", () =>
    formularioModal({
      titulo: "Nueva mina",
      campos: [
        { nombre: "nombre", etiqueta: "Nombre", placeholder: "Mina Norte" },
        { nombre: "area", etiqueta: "Área / zona", opcional: true, placeholder: "Norte" },
      ],
      textoEnviar: "Crear mina",
      alEnviar: async (datos) => {
        await llamarApi("/catalogos/minas", { metodo: "POST", cuerpo: datos });
        avisar(`Mina ${datos.nombre} creada.`);
        mostrarVista("minas");
      },
    })
  );
  pintarCargando(2);
  const lista = await llamarApi("/catalogos/minas").catch((excepcion) => { avisar(excepcion.message, "error"); return []; });
  if (!lista?.length) { $("#vista").innerHTML = estadoVacio("Sin minas", "Crea la primera unidad minera."); return; }
  $("#vista").innerHTML = `<div class="tarjeta" style="padding:6px 0"><table class="tabla">
    <thead><tr><th>Mina</th><th>Área</th><th class="mono">Identificador</th></tr></thead>
    <tbody>${lista.map((mina) => `
      <tr><td><strong>${escapar(mina.Nombre)}</strong></td><td>${escapar(mina.Area) || "—"}</td><td class="mono">${mina.Identificador.slice(0, 8)}…</td></tr>`).join("")}
    </tbody></table></div>`;
}

async function empleados() {
  const minasDisponibles = await llamarApi("/catalogos/minas").catch(() => []);
  botonDeAccion("Nuevo empleado", "catalogos.editar", () =>
    formularioModal({
      titulo: "Nuevo empleado",
      campos: [
        { nombre: "numero_nomina", etiqueta: "Número de nómina", placeholder: "N-0042" },
        { nombre: "nombre_completo", etiqueta: "Nombre completo", placeholder: "JUAN PÉREZ" },
        { nombre: "id_mina", etiqueta: "Mina", tipo: "select",
          opciones: minasDisponibles.map((mina) => ({ valor: mina.Identificador, texto: mina.Nombre })) },
      ],
      textoEnviar: "Contratar",
      alEnviar: async (datos) => {
        await llamarApi("/catalogos/empleados", { metodo: "POST", cuerpo: datos });
        avisar(`Empleado ${datos.numero_nomina} dado de alta.`);
        mostrarVista("empleados");
      },
    })
  );
  pintarCargando(3);
  const lista = await llamarApi("/catalogos/empleados").catch((excepcion) => { avisar(excepcion.message, "error"); return []; });
  if (!lista?.length) { $("#vista").innerHTML = estadoVacio("Sin empleados", "Da de alta al personal operativo."); return; }
  $("#vista").innerHTML = `<div class="tarjeta" style="padding:6px 0"><table class="tabla">
    <thead><tr><th>Nómina</th><th>Nombre</th><th>Mina</th></tr></thead>
    <tbody>${lista.map((empleado) => `
      <tr><td class="mono">${escapar(empleado.NumeroNomina)}</td><td>${escapar(empleado.NombreCompleto)}</td><td>${escapar(empleado.Mina)}</td></tr>`).join("")}
    </tbody></table></div>`;
}

async function equipos() {
  const [minasDisponibles, tipos, modulos] = await Promise.all([
    llamarApi("/catalogos/minas").catch(() => []),
    llamarApi("/catalogos/tipos-de-equipo").catch(() => []),
    llamarApi("/catalogos/modulos-de-trabajo").catch(() => []),
  ]);
  botonDeAccion("Nuevo equipo", "catalogos.editar", () =>
    formularioModal({
      titulo: "Nuevo equipo",
      nota: "Tiro (shaft) y banda de extracción también son equipos de acarreo.",
      campos: [
        { nombre: "codigo", etiqueta: "Código", placeholder: "CAM-07" },
        { nombre: "id_mina", etiqueta: "Mina", tipo: "select",
          opciones: minasDisponibles.map((mina) => ({ valor: mina.Identificador, texto: mina.Nombre })) },
        { nombre: "id_tipo_equipo", etiqueta: "Tipo de equipo", tipo: "select",
          opciones: tipos.map((tipo) => ({ valor: tipo.Identificador, texto: `${tipo.Codigo} — ${tipo.Descripcion}` })) },
        { nombre: "id_modulo_trabajo", etiqueta: "Módulo de trabajo", tipo: "select",
          opciones: modulos.map((modulo) => ({ valor: modulo.Identificador, texto: `${modulo.Codigo} — ${modulo.Descripcion}` })) },
        { nombre: "fabricante", etiqueta: "Fabricante", opcional: true, placeholder: "Sandvik" },
      ],
      textoEnviar: "Dar de alta",
      alEnviar: async (datos) => {
        await llamarApi("/catalogos/equipos", { metodo: "POST", cuerpo: datos });
        avisar(`Equipo ${datos.codigo} dado de alta.`);
        mostrarVista("equipos");
      },
    })
  );
  pintarCargando(5);
  const lista = await llamarApi("/catalogos/equipos").catch((excepcion) => { avisar(excepcion.message, "error"); return []; });
  if (!lista?.length) { $("#vista").innerHTML = estadoVacio("Sin equipos", "Registra la flota de la operación."); return; }
  $("#vista").innerHTML = `<div class="tarjeta" style="padding:6px 0"><table class="tabla">
    <thead><tr><th>Código</th><th>Tipo</th><th>Módulo</th><th>Mina</th><th>Fabricante</th></tr></thead>
    <tbody>${lista.map((equipo) => `
      <tr>
        <td class="mono"><strong>${escapar(equipo.Codigo)}</strong></td>
        <td>${escapar(equipo.Tipo)}</td>
        <td><span class="chip">${escapar(equipo.Modulo)}</span></td>
        <td>${escapar(equipo.Mina)}</td>
        <td>${escapar(equipo.Fabricante) || "—"}</td>
      </tr>`).join("")}
    </tbody></table></div>`;
}

async function empresa() {
  const datos = estado.empresa ?? await llamarApi("/gobierno/empresa");
  const puedeConfigurar = tienePermiso("empresa.configurar");
  $("#vista").innerHTML = "";
  const tarjeta = elemento(`<div class="vista-empresa">
    <form class="tarjeta" id="formulario-empresa">
      <h3>Identidad visual</h3>
      <div class="previsualizacion-marca">
        ${datos.LogoUrl ? `<img src="${escapar(datos.LogoUrl)}" alt="logo">` : ICONOS.minas}
        <div><strong>${escapar(datos.RazonSocial)}</strong><br><small>${escapar(datos.Codigo)} · el color pinta toda la interfaz</small></div>
      </div>
      <label class="campo"><span>URL del logo</span>
        <input name="logo_url" value="${escapar(datos.LogoUrl)}" placeholder="https://cdn.miempresa.mx/logo.png" ${puedeConfigurar ? "" : "disabled"}></label>
      <label class="campo"><span>Color primario</span>
        <div class="fila-color">
          <input type="color" id="selector-color" value="${datos.ColorPrimario || "#d9a440"}" ${puedeConfigurar ? "" : "disabled"}>
          <input type="text" name="color_primario" id="texto-color" value="${escapar(datos.ColorPrimario)}" placeholder="#D9A440" ${puedeConfigurar ? "" : "disabled"}>
        </div></label>
      <label class="campo"><span>Zona horaria</span>
        <input name="zona_horaria" value="${escapar(datos.ZonaHoraria)}" ${puedeConfigurar ? "" : "disabled"}></label>
      <label class="campo"><span>Moneda</span>
        <select name="moneda" ${puedeConfigurar ? "" : "disabled"}>
          <option ${datos.Moneda === "USD" ? "selected" : ""}>USD</option>
          <option ${datos.Moneda === "MXN" ? "selected" : ""}>MXN</option>
        </select></label>
      ${puedeConfigurar ? '<div class="pie-modal"><button type="submit" class="boton-accion">Guardar y repintar</button></div>' : '<p class="nota-modal">Solo el ADMIN_EMPRESA puede configurar el branding.</p>'}
    </form>
    <div class="tarjeta">
      <h3>Cómo funciona</h3>
      <p style="color:var(--texto-2);font-size:13.5px">
        Esta identidad vive en la tabla <span class="chip">gobierno.empresa</span> de <b>tu</b> tenant.
        Al guardar, el backend valida el color (#RRGGBB) en la base y esta interfaz se repinta
        al instante con tu color — cada empresa ve el sistema con su propia marca.
      </p>
      <p style="color:var(--texto-2);font-size:13.5px">
        El logo es una URL: el archivo vive en tu almacenamiento (CDN/objetos), la base solo guarda la referencia.
      </p>
    </div>
  </div>`);
  $("#vista").appendChild(tarjeta);

  const selectorColor = $("#selector-color");
  const textoColor = $("#texto-color");
  selectorColor?.addEventListener("input", () => { textoColor.value = selectorColor.value.toUpperCase(); });
  textoColor?.addEventListener("input", () => { if (/^#[0-9a-fA-F]{6}$/.test(textoColor.value)) selectorColor.value = textoColor.value; });

  $("#formulario-empresa").addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const formulario = Object.fromEntries(new FormData(evento.target).entries());
    try {
      await llamarApi("/gobierno/empresa", { metodo: "PUT", cuerpo: formulario });
      aplicarMarca(await llamarApi("/gobierno/empresa"));
      avisar("Branding guardado: interfaz repintada con tu marca.");
      mostrarVista("empresa");
    } catch (excepcion) { avisar(excepcion.message, "error"); }
  });
}

$("#formulario-login").addEventListener("submit", entrar);
$("#usar-demo").addEventListener("click", () => {
  $("#campo-empresa").value = "MIN";
  $("#campo-usuario").value = "admin.mina";
  $("#campo-contrasena").value = "Mina#2026";
  $("#boton-entrar").focus();
});
$("#boton-salir").addEventListener("click", cerrarSesion);
$("#cerrar-modal").addEventListener("click", cerrarModal);
$("#capa-modal").addEventListener("click", (evento) => { if (evento.target.id === "capa-modal") cerrarModal(); });
document.addEventListener("keydown", (evento) => { if (evento.key === "Escape" && !$("#capa-modal").hidden) cerrarModal(); });

const sesionGuardada = localStorage.getItem(CLAVE_SESION);
if (sesionGuardada) {
  estado.sesion = JSON.parse(sesionGuardada);
  arrancarAplicacion();
}
