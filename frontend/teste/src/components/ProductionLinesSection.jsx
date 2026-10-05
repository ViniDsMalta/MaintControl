import { useEffect, useMemo, useState } from "react";
import { api } from "../api/api";

export function ProductionLinesSection({ token, lines, machines, refreshLines, onActivity }) {
  const [name, setName] = useState("");
  const [selectedId, setSelectedId] = useState("");
  const [selectedLine, setSelectedLine] = useState(null);
  const [machineId, setMachineId] = useState("");
  const [position, setPosition] = useState(1);
  const [loading, setLoading] = useState(false);

  const assignedIds = useMemo(() => new Set((selectedLine?.machines || []).map((machine) => machine.id)), [selectedLine]);
  const availableMachines = machines.filter((machine) => !assignedIds.has(machine.id));

  useEffect(() => {
    if (!selectedId || !token) {
      setSelectedLine(null);
      return;
    }
    loadSelectedLine(selectedId);
  }, [selectedId, token]);

  async function loadSelectedLine(id = selectedId) {
    if (!id) return;
    setLoading(true);
    try {
      const detail = await api.getProductionLine(token, id);
      setSelectedLine(detail);
      const nextPosition = Math.max(0, ...detail.machines.map((machine) => machine.position)) + 1;
      setPosition(nextPosition);
    } catch (error) {
      onActivity("error", error.message);
    } finally {
      setLoading(false);
    }
  }

  async function createLine(event) {
    event.preventDefault();
    setLoading(true);
    try {
      const created = await api.createProductionLine(token, name);
      setName("");
      await refreshLines();
      setSelectedId(created.id);
      onActivity("success", `Linha ${created.name} criada.`);
    } catch (error) {
      onActivity("error", error.message);
    } finally {
      setLoading(false);
    }
  }

  async function addMachine(event) {
    event.preventDefault();
    setLoading(true);
    try {
      await api.addMachineToLine(token, selectedId, machineId, Number(position));
      setMachineId("");
      await loadSelectedLine();
      onActivity("success", "Máquina adicionada à linha.");
    } catch (error) {
      onActivity("error", error.message);
    } finally {
      setLoading(false);
    }
  }

  async function removeMachine(machine) {
    setLoading(true);
    try {
      await api.removeMachineFromLine(token, selectedId, machine.id);
      await loadSelectedLine();
      onActivity("success", `Máquina ${machine.name} removida da linha.`);
    } catch (error) {
      onActivity("error", error.message);
    } finally {
      setLoading(false);
    }
  }

  async function deleteLine() {
    if (!selectedId || !window.confirm("Excluir esta linha de produção?")) return;
    setLoading(true);
    try {
      await api.deleteProductionLine(token, selectedId);
      setSelectedId("");
      setSelectedLine(null);
      await refreshLines();
      onActivity("success", "Linha de produção excluída.");
    } catch (error) {
      onActivity("error", error.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <section>
      <div className="section-heading"><h2>Linhas de produção</h2><span>{loading ? "Carregando..." : `${lines.length} linha(s)`}</span></div>
      <form className="inline-form" onSubmit={createLine}>
        <label>Nome<input value={name} onChange={(event) => setName(event.target.value)} required disabled={!token} /></label>
        <button type="submit" disabled={!token || loading}>Criar linha</button>
      </form>

      <div className="line-selector">
        <label>
          Linha selecionada
          <select value={selectedId} onChange={(event) => setSelectedId(event.target.value)} disabled={!token}>
            <option value="">Selecione uma linha</option>
            {lines.map((line) => <option key={line.id} value={line.id}>{line.name}</option>)}
          </select>
        </label>
        <button type="button" className="secondary" onClick={() => loadSelectedLine()} disabled={!selectedId || loading}>Atualizar linha</button>
        <button type="button" className="danger" onClick={deleteLine} disabled={!selectedId || loading}>Excluir linha</button>
      </div>

      {selectedLine && (
        <>
          <form className="inline-form" onSubmit={addMachine}>
            <label>
              Máquina disponível
              <select value={machineId} onChange={(event) => setMachineId(event.target.value)} required>
                <option value="">Selecione uma máquina</option>
                {availableMachines.map((machine) => <option key={machine.id} value={machine.id}>{machine.name} ({machine.type})</option>)}
              </select>
            </label>
            <label>Posição<input type="number" min="1" value={position} onChange={(event) => setPosition(event.target.value)} required /></label>
            <button type="submit" disabled={!machineId || loading}>Adicionar máquina</button>
          </form>
          <div className="table-wrap">
            <table>
              <thead><tr><th>Posição</th><th>Máquina</th><th>Tipo</th><th>Ação</th></tr></thead>
              <tbody>
                {selectedLine.machines.map((machine) => (
                  <tr key={machine.id}>
                    <td>{machine.position}</td><td>{machine.name}</td><td>{machine.type}</td>
                    <td><button type="button" className="danger" onClick={() => removeMachine(machine)}>Remover</button></td>
                  </tr>
                ))}
                {selectedLine.machines.length === 0 && <tr><td colSpan="4">Nenhuma máquina nesta linha.</td></tr>}
              </tbody>
            </table>
          </div>
        </>
      )}
    </section>
  );
}
