"use strict";

// El cliente usa la API del TP3 desde el mismo origen que sirve estos archivos.
const entities = {
  libros: { singular: "Libro", empty: "Todavía no hay libros. Agregá el primero.", numbers: ["id_autor", "id_editorial", "cantidad_paginas"] },
  autores: { singular: "Autor", empty: "Todavía no hay autores. Sumá uno para empezar.", numbers: [] },
  editoriales: { singular: "Editorial", empty: "Todavía no hay editoriales. Agregá una para tus libros.", numbers: [] },
  usuarios: { singular: "Usuario", empty: "Todavía no hay usuarios registrados.", numbers: [] },
};
const records = { libros: [], autores: [], editoriales: [], usuarios: [] };
let busy = false;
const loadState = { libros: "loading", autores: "loading", editoriales: "loading", usuarios: "loading" };
let loadVersion = 0;
const requestTimeout = 15000; // Decisión del proyecto: no dejar una petición esperando sin límite.

function showMessage(entity, text, error = false) {
  const message = document.getElementById(`${entity}-message`);
  message.textContent = text;
  message.classList.toggle("error", error);
}

async function request(path, options = {}) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), requestTimeout);
  try {
    let response;
    try {
      response = await fetch(path, { ...options, signal: controller.signal });
    } catch {
      throw new Error("No se pudo conectar con el servidor. Revisá la conexión e intentá nuevamente.");
    }
    if (response.status === 204) return null;
    let data;
    try {
      data = await response.json();
    } catch {
      throw new Error("El servidor devolvió una respuesta inesperada. Intentá nuevamente.");
    }
    if (!response.ok) throw new Error(data.error || `No se pudo completar la operación (${response.status}).`);
    return data;
  } catch (error) {
    if (controller.signal.aborted) {
      const message = options.method && options.method !== "GET"
        ? "El servidor tardó demasiado. No se pudo confirmar el cambio; recargá la página para comprobar si se guardó antes de repetirlo."
        : "El servidor tardó demasiado en responder.";
      throw new Error(message);
    }
    throw error;
  } finally {
    clearTimeout(timer);
  }
}

function setBusy(value) {
  busy = value;
  for (const entity of Object.keys(entities)) {
    const loading = loadState[entity] === "loading";
    const referencesLoading = entity === "libros" &&
      (loadState.autores === "loading" || loadState.editoriales === "loading");
    document.querySelector(`#${entity}-form button`).disabled = busy || loading || referencesLoading;
    document.querySelectorAll(`#${entity}-list .delete-button`).forEach(button => {
      button.disabled = busy || loading;
    });
  }
}

function updateSelect(entity, selectID) {
  const select = document.getElementById(selectID);
  const previous = select.value;
  const placeholder = new Option(records[entity].length ? "Seleccioná una opción" : `Primero agregá ${entity}`, "");
  select.replaceChildren(placeholder);
  for (const item of records[entity]) select.add(new Option(item.nombre, String(item.id)));
  if (records[entity].some(item => String(item.id) === previous)) select.value = previous;
}

function describe(entity, item) {
  switch (entity) {
    case "libros": {
      const author = records.autores.find(author => author.id === item.id_autor);
      const publisher = records.editoriales.find(publisher => publisher.id === item.id_editorial);
      return [item.titulo, `${author?.nombre || `Autor #${item.id_autor}`} · ${publisher?.nombre || `Editorial #${item.id_editorial}`}`, `${item.cantidad_paginas} páginas · Lanzamiento: ${item.fecha_lanzamiento}`];
    }
    case "autores": return [item.nombre, `${item.pais} · Nacimiento: ${item.fecha_nacimiento}`, item.biografia];
    case "editoriales": return [item.nombre, `Sede: ${item.sede}`];
    case "usuarios": return [item.nombre, item.email];
  }
}

function renderList(entity) {
  const list = document.getElementById(`${entity}-list`);
  list.replaceChildren();
  const count = records[entity].length;
  document.getElementById(`${entity}-count`).textContent = `${count} ${count === 1 ? "registro" : "registros"}`;
  if (!count) {
    const empty = document.createElement("li");
    empty.className = "empty";
    empty.textContent = entities[entity].empty;
    list.append(empty);
    return;
  }
  for (const item of records[entity]) {
    const row = document.createElement("li");
    row.className = "entity-card";
    const info = document.createElement("div");
    info.className = "entity-info";
    describe(entity, item).forEach((text, index) => {
      const node = document.createElement(index === 0 ? "h3" : "p");
      // textContent muestra los datos como texto; nunca interpreta HTML del usuario.
      node.textContent = text;
      info.append(node);
    });
    const remove = document.createElement("button");
    remove.type = "button";
    remove.className = "delete-button";
    remove.dataset.id = String(item.id);
    remove.textContent = "Eliminar";
    remove.setAttribute("aria-label", `Eliminar ${item.titulo || item.nombre}`);
    remove.disabled = busy;
    row.append(info, remove);
    list.append(row);
  }
}

async function loadLists() {
  const version = ++loadVersion;
  const failed = [];
  for (const entity of Object.keys(entities)) loadState[entity] = "loading";
  setBusy(busy);
  await Promise.all(Object.keys(entities).map(async entity => {
    try {
      const data = await request(`/${entity}`);
      // Una respuesta de una carga anterior no debe pisar datos más recientes.
      if (version !== loadVersion) return;
      if (!Array.isArray(data)) throw new Error("El servidor no devolvió una lista válida.");
      records[entity] = data;
      loadState[entity] = "ready";
      showMessage(entity, "");
      renderList(entity);
      if (entity === "autores" || entity === "editoriales") {
        updateSelect(entity, entity === "autores" ? "libro-autor" : "libro-editorial");
        // Si los libros llegaron antes que sus referencias, actualizar los nombres.
        if (loadState.libros === "ready") renderList("libros");
      }
    } catch (error) {
      if (version !== loadVersion) return;
      loadState[entity] = "error";
      showMessage(entity, `No se pudo actualizar la lista. ${error.message} Recargá la página para reintentar.`, true);
      document.getElementById(`${entity}-count`).textContent = "Sin actualizar";
      const list = document.getElementById(`${entity}-list`);
      if (!records[entity].length) list.replaceChildren();
      if (entity === "autores" || entity === "editoriales") {
        updateSelect(entity, entity === "autores" ? "libro-autor" : "libro-editorial");
      }
      failed.push(entity);
    } finally {
      // Cada sección puede usarse apenas termina su propia consulta.
      if (version === loadVersion) setBusy(busy);
    }
  }));
  return failed;
}

async function mutate(entity, options, path, form = null) {
  if (busy) return;
  const fields = form ? Array.from(form.querySelectorAll("input, select, textarea")) : [];
  setBusy(true);
  fields.forEach(field => { field.disabled = true; });
  showMessage(entity, "Guardando cambios…");
  try {
    try {
      await request(path, options);
    } catch (error) {
      showMessage(entity, error.message, true);
      return;
    }
    // Los campos permanecen bloqueados hasta terminar: no se borra un borrador nuevo.
    if (form) form.reset();
    const failed = await loadLists();
    const action = form ? "creado" : "eliminado";
    showMessage(entity, `${entities[entity].singular} ${action}.${failed.length ? " El cambio se guardó, pero no se pudieron actualizar todas las listas. Recargá la página." : ""}`, failed.length > 0);
  } finally {
    fields.forEach(field => { field.disabled = false; });
    setBusy(false);
  }
}

for (const [entity, config] of Object.entries(entities)) {
  const form = document.getElementById(`${entity}-form`);
  form.addEventListener("submit", event => {
    event.preventDefault();
    if (busy || !form.reportValidity()) return;
    const data = Object.fromEntries(new FormData(form));
    for (const key of Object.keys(data)) {
      if (key !== "contrasena") data[key] = data[key].trim();
    }
    if (Object.entries(data).some(([key, value]) => key !== "contrasena" && value === "")) {
      showMessage(entity, "Completá los campos con texto, no solo espacios.", true);
      return;
    }
    for (const key of config.numbers) data[key] = Number(data[key]);
    void mutate(entity, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(data) }, `/${entity}`, form);
  });
  // Delegación: el listener vive en la lista, aunque sus botones se regeneren.
  document.getElementById(`${entity}-list`).addEventListener("click", event => {
    const button = event.target.closest("button[data-id]");
    if (!button || busy) return;
    void mutate(entity, { method: "DELETE" }, `/${entity}/${button.dataset.id}`);
  });
}

const today = new Date();
const localDate = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}-${String(today.getDate()).padStart(2, "0")}`;
document.getElementById("libro-fecha").max = localDate;
void loadLists();
