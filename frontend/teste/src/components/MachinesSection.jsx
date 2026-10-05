import { useState } from "react";
import { api } from "../api/api";

function formatNumber(value) {
  return typeof value === "number" ? value.toFixed(2) : "—";
}

function formatDate(value) {
  return value ? new Date(value).toLocaleString("pt-BR") : "—";
}

export function MachinesSection({ token, machines, loading, refresh, onActivity }) {
  const [name, setName] = useState("");
  const [type, setType] = useState("motor");
  const [saving, setSaving] = useState(false);

  async function createMachine(event) {
    event.preventDefault();
    setSaving(true);
    try {
      await api.createMachine(token, name, type);
      setName("");
      onActivity("success", `Máquina ${name} criada.`);
      await refresh();
    } catch (error) {
      onActivity("error", error.message);
    } finally {
      setSaving(false);
    }
  }

  async function deleteMachine(machine) {
    if (!window.confirm(`Excluir a máquina "${machine.name}" e suas telemetrias?`)) return;
    try {
      await api.deleteMachine(token, machine.id);
      onActivity("success", `Máquina ${machine.name} excluída.`);
      await refresh();
    } catch (error) {
      onActivity("error", error.message);
    }
  }

  return (
    <section>
      <div className="section-heading">
        <h2>Máquinas</h2>
        <button type="button" className="secondary" onClick={refresh} disabled={!token || loading}>
          {loading ? "Atualizando..." : "Atualizar"}
        </button>
      </div>

      <form className="inline-form" onSubmit={createMachine}>
        <label>
          Nome
          <input value={name} onChange={(event) => setName(event.target.value)} required disabled={!token} />
        </label>
        <label>
          Tipo
          <select value={type} onChange={(event) => setType(event.target.value)} disabled={!token}>
            <option value="motor">Motor</option>
            <option value="bomba">Bomba</option>
            <option value="compressor">Compressor</option>
          </select>
        </label>
        <button type="submit" disabled={!token || saving}>{saving ? "Criando..." : "Criar máquina"}</button>
      </form>

      {!token ? <p className="muted">Faça login para gerenciar máquinas.</p> : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr><th>Nome</th><th>Tipo</th><th>Status</th><th>Saúde</th><th>Risco</th><th>Atualizado</th><th>Ação</th></tr>
            </thead>
            <tbody>
              {machines.map((machine) => (
                <tr key={machine.id}>
                  <td><strong>{machine.name}</strong><small>{machine.id}</small></td>
                  <td>{machine.type}</td>
                  <td><span className={`status status-${machine.machine_status?.status || "unknown"}`}>{machine.machine_status?.status || "—"}</span></td>
                  <td>{formatNumber(machine.machine_status?.health_score)}</td>
                  <td>{formatNumber(machine.machine_status?.risk_score)}</td>
                  <td>{formatDate(machine.machine_status?.updated_at)}</td>
                  <td><button type="button" className="danger" onClick={() => deleteMachine(machine)}>Excluir</button></td>
                </tr>
              ))}
              {!loading && machines.length === 0 && <tr><td colSpan="7">Nenhuma máquina cadastrada.</td></tr>}
            </tbody>
          </table>
        </div>
      )}
      <p className="muted">Atualização automática a cada 5 segundos.</p>
    </section>
  );
}
