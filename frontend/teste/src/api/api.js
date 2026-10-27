const API_URL = (import.meta.env.VITE_API_URL || "/api").replace(/\/$/, "");

async function request(path, { method = "GET", token, body } = {}) {
  const headers = { Accept: "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  if (body !== undefined) headers["Content-Type"] = "application/json";

  let response;
  try {
    response = await fetch(`${API_URL}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new Error("Backend indisponível");
  }

  const data = response.status === 204 ? null : await response.json().catch(() => null);
  if (!response.ok) {
    const messages = {
      "invalid input": "Confira os dados informados.",
      "invalid credentials": "E-mail ou senha incorretos.",
      "email already exists": "Este e-mail já está cadastrado.",
      "not found": "Registro não encontrado.",
      "machine already belongs to this production line": "Esta máquina já pertence ao fluxo.",
      "position already in use": "Esta posição já está ocupada.",
      unauthorized: "Sua sessão expirou. Entre novamente.",
    };
    const message = data?.error || `Erro HTTP ${response.status}`;
    const error = new Error(messages[message] || message);
    error.status = response.status;
    throw error;
  }
  return data;
}

export const api = {
  health: () => request("/health"),
  register: (name, email, password) =>
    request("/auth/register", { method: "POST", body: { name, email, password } }),
  login: (email, password) =>
    request("/auth/login", { method: "POST", body: { email, password } }),
  getMe: (token) => request("/auth/me", { token }),

  getMachines: (token) => request("/machines", { token }),
  getMachine: (token, id) => request(`/machines/${id}`, { token }),
  createMachine: (token, name, type) =>
    request("/machines", { method: "POST", token, body: { name, type } }),
  updateMachine: (token, id, name, type) =>
    request(`/machines/${id}`, { method: "PUT", token, body: { name, type } }),
  deleteMachine: (token, id) => request(`/machines/${id}`, { method: "DELETE", token }),

  getProductionLines: (token) => request("/production-lines", { token }),
  getProductionLine: (token, id) => request(`/production-lines/${id}`, { token }),
  createProductionLine: (token, name) =>
    request("/production-lines", { method: "POST", token, body: { name } }),
  updateProductionLine: (token, id, name) =>
    request(`/production-lines/${id}`, { method: "PUT", token, body: { name } }),
  deleteProductionLine: (token, id) =>
    request(`/production-lines/${id}`, { method: "DELETE", token }),
  getLineMachines: (token, id) => request(`/production-lines/${id}/machines`, { token }),
  addMachineToLine: (token, lineId, machineId, position) =>
    request(`/production-lines/${lineId}/machines`, {
      method: "POST",
      token,
      body: { machine_id: machineId, position },
    }),
  removeMachineFromLine: (token, lineId, machineId) =>
    request(`/production-lines/${lineId}/machines/${machineId}`, { method: "DELETE", token }),
};
