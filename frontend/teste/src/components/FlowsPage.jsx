import { useEffect, useMemo, useState } from "react";
import { Boxes, Edit3, GripVertical, Network, Plus, RefreshCw, Trash2, Unlink } from "lucide-react";
import { api } from "../api/api";
import { ConfirmDialog, Modal } from "./Modal";

const typeLabels = { motor: "Motor", bomba: "Bomba", compressor: "Compressor" };

export function FlowsPage({ token, machines, lines, lineDetails, refresh, notify }) {
  const [selectedId, setSelectedId] = useState("");
  const [formModal, setFormModal] = useState(null);
  const [deletingLine, setDeletingLine] = useState(null);
  const [removingMachine, setRemovingMachine] = useState(null);
  const [machineId, setMachineId] = useState("");
  const [position, setPosition] = useState(1);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!selectedId && lines.length) setSelectedId(lines[0].id);
    if (selectedId && !lines.some((line) => line.id === selectedId)) setSelectedId(lines[0]?.id || "");
  }, [lines, selectedId]);

  const selectedLine = lines.find((line) => line.id === selectedId);
  const selectedDetail = lineDetails[selectedId];
  const assignedMachines = selectedDetail?.machines || [];
  const assignedIds = useMemo(() => new Set(assignedMachines.map((machine) => machine.id)), [assignedMachines]);
  const availableMachines = machines.filter((machine) => !assignedIds.has(machine.id));

  useEffect(() => {
    const next = Math.max(0, ...assignedMachines.map((machine) => machine.position || 0)) + 1;
    setPosition(next);
    setMachineId("");
  }, [selectedId, assignedMachines.length]);

  async function saveLine(name) {
    setBusy(true);
    try {
      if (formModal?.line) {
        await api.updateProductionLine(token, formModal.line.id, name);
        notify("success", "Fluxo renomeado com sucesso.");
      } else {
        const created = await api.createProductionLine(token, name);
        setSelectedId(created.id);
        notify("success", `Fluxo ${created.name} criado.`);
      }
      setFormModal(null);
      await refresh();
    } catch (error) {
      notify("error", error.message);
    } finally {
      setBusy(false);
    }
  }

  async function addMachine(event) {
    event.preventDefault();
    setBusy(true);
    try {
      await api.addMachineToLine(token, selectedId, machineId, Number(position));
      notify("success", "Máquina adicionada ao fluxo.");
      setMachineId("");
      await refresh();
    } catch (error) {
      notify("error", error.message);
    } finally {
      setBusy(false);
    }
  }

  async function removeMachine() {
    if (!removingMachine) return;
    setBusy(true);
    try {
      await api.removeMachineFromLine(token, selectedId, removingMachine.id);
      notify("success", `${removingMachine.name} removida do fluxo.`);
      setRemovingMachine(null);
      await refresh();
    } catch (error) {
      notify("error", error.message);
    } finally {
      setBusy(false);
    }
  }

  async function deleteLine() {
    if (!deletingLine) return;
    setBusy(true);
    try {
      await api.deleteProductionLine(token, deletingLine.id);
      notify("success", `Fluxo ${deletingLine.name} excluído.`);
      setDeletingLine(null);
      setSelectedId("");
      await refresh();
    } catch (error) {
      notify("error", error.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="page flows-page">
      <header className="page-heading">
        <div><p className="eyebrow">Organização produtiva</p><h1>Fluxos</h1><p>Agrupe as máquinas e defina a ordem dos ativos em cada processo.</p></div>
        <button type="button" className="button button-primary" onClick={() => setFormModal({ line: null })}><Plus size={18} /> Novo fluxo</button>
      </header>

      <section className="flows-layout">
        <aside className="panel flow-list-panel">
          <div className="panel-heading"><div><h2>Fluxos produtivos</h2><p>{lines.length} cadastrado(s)</p></div><button type="button" className="icon-button bordered" onClick={refresh} aria-label="Atualizar fluxos" title="Atualizar"><RefreshCw size={17} /></button></div>
          <div className="flow-list">
            {lines.map((line) => {
              const count = lineDetails[line.id]?.machines?.length || 0;
              return <button type="button" key={line.id} className={selectedId === line.id ? "active" : ""} onClick={() => setSelectedId(line.id)}><span className="flow-icon"><Network size={19} /></span><span><strong>{line.name}</strong><small>{count} máquina(s)</small></span><i /></button>;
            })}
          </div>
          {!lines.length && <div className="empty-state compact"><span><Network size={24} /></span><h3>Nenhum fluxo</h3><p>Crie um fluxo para organizar a produção.</p></div>}
        </aside>

        <article className="panel flow-detail-panel">
          {selectedLine ? (
            <>
              <div className="flow-detail-header">
                <div><p className="eyebrow">Fluxo selecionado</p><h2>{selectedLine.name}</h2><p>{assignedMachines.length} máquina(s) no processo</p></div>
                <div className="icon-actions"><button type="button" className="icon-button bordered" onClick={() => setFormModal({ line: selectedLine })} aria-label="Renomear fluxo" title="Renomear"><Edit3 size={17} /></button><button type="button" className="icon-button danger-outline" onClick={() => setDeletingLine(selectedLine)} aria-label="Excluir fluxo" title="Excluir"><Trash2 size={17} /></button></div>
              </div>

              <form className="assign-form" onSubmit={addMachine}>
                <label>Máquina disponível<select value={machineId} onChange={(event) => setMachineId(event.target.value)} required><option value="">Selecione uma máquina</option>{availableMachines.map((machine) => <option key={machine.id} value={machine.id}>{machine.name} · {typeLabels[machine.type]}</option>)}</select></label>
                <label className="position-field">Posição<input type="number" min="1" value={position} onChange={(event) => setPosition(event.target.value)} required /></label>
                <button type="submit" className="button button-primary" disabled={!machineId || busy}><Plus size={17} /> Adicionar</button>
              </form>
              {!availableMachines.length && machines.length > 0 && <p className="form-hint">Todas as máquinas já estão vinculadas a este fluxo.</p>}

              <div className="flow-sequence">
                {assignedMachines.map((machine, index) => (
                  <div className="sequence-item" key={machine.id}>
                    <span className="sequence-position">{machine.position}</span>
                    <span className="drag-mark"><GripVertical size={18} /></span>
                    <span className="machine-type-icon"><Boxes size={17} /></span>
                    <span className="sequence-copy"><strong>{machine.name}</strong><small>{typeLabels[machine.type] || machine.type}</small></span>
                    <button type="button" className="icon-button subtle danger-text" onClick={() => setRemovingMachine(machine)} aria-label={`Remover ${machine.name}`} title="Remover do fluxo"><Unlink size={18} /></button>
                    {index < assignedMachines.length - 1 && <span className="sequence-line" />}
                  </div>
                ))}
                {!assignedMachines.length && <div className="empty-state"><span><Network size={27} /></span><h3>Fluxo ainda vazio</h3><p>Selecione uma máquina acima para montar a sequência produtiva.</p></div>}
              </div>
            </>
          ) : (
            <div className="empty-state full"><span><Network size={30} /></span><h3>Selecione ou crie um fluxo</h3><p>Os detalhes e a sequência de máquinas aparecerão aqui.</p><button type="button" className="button button-primary" onClick={() => setFormModal({ line: null })}><Plus size={17} /> Criar fluxo</button></div>
          )}
        </article>
      </section>

      {formModal && <LineForm line={formModal.line} busy={busy} onSubmit={saveLine} onClose={() => setFormModal(null)} />}
      {deletingLine && <ConfirmDialog title="Excluir fluxo?" message={`O fluxo “${deletingLine.name}” será excluído. As máquinas continuarão cadastradas.`} onConfirm={deleteLine} onClose={() => setDeletingLine(null)} loading={busy} />}
      {removingMachine && <ConfirmDialog title="Remover máquina do fluxo?" message={`“${removingMachine.name}” será desvinculada deste fluxo, mas continuará cadastrada no sistema.`} confirmLabel="Remover" onConfirm={removeMachine} onClose={() => setRemovingMachine(null)} loading={busy} />}
    </main>
  );
}

function LineForm({ line, busy, onSubmit, onClose }) {
  const [name, setName] = useState(line?.name || "");
  return (
    <Modal title={line ? "Renomear fluxo" : "Novo fluxo"} description="Use um nome que identifique claramente o processo produtivo." onClose={onClose} size="small">
      <form onSubmit={(event) => { event.preventDefault(); onSubmit(name); }}>
        <div className="form-stack"><label>Nome do fluxo<input value={name} onChange={(event) => setName(event.target.value)} placeholder="Ex.: Linha de envase" autoFocus required /></label></div>
        <div className="modal-actions"><button type="button" className="button button-secondary" onClick={onClose}>Cancelar</button><button type="submit" className="button button-primary" disabled={busy || !name.trim()}>{busy ? "Salvando..." : line ? "Salvar nome" : "Criar fluxo"}</button></div>
      </form>
    </Modal>
  );
}
