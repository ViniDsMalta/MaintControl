import { useMemo, useState } from "react";
import { Boxes, Clipboard, ClipboardCheck, Edit3, Eye, EyeOff, MoreHorizontal, Plus, RefreshCw, Search, Trash2, X } from "lucide-react";
import { api } from "../api/api";
import { ConfirmDialog, Modal } from "./Modal";

const typeLabels = { motor: "Motor", bomba: "Bomba", compressor: "Compressor" };
const statusLabels = {
  NORMAL: ["Normal", "normal"],
  DEGRADATION: ["Atenção", "warning"],
  FAILURE: ["Crítico", "critical"],
  AGUARDANDO_DADOS: ["Aguardando dados", "waiting"],
};

function formatScore(value) {
  return typeof value === "number" ? `${Math.round(value)}%` : "—";
}

function formatDate(value) {
  return value ? new Date(value).toLocaleString("pt-BR", { dateStyle: "short", timeStyle: "short" }) : "Sem leituras";
}

export function MachinesPage({ token, machines, refreshing, refresh, notify }) {
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [formModal, setFormModal] = useState(null);
  const [selected, setSelected] = useState(null);
  const [deleting, setDeleting] = useState(null);
  const [busy, setBusy] = useState(false);
  const [visibleKeys, setVisibleKeys] = useState(new Set());
  const [copiedId, setCopiedId] = useState("");

  const filtered = useMemo(() => machines.filter((machine) => {
    const matchesSearch = `${machine.name} ${machine.type}`.toLowerCase().includes(search.toLowerCase());
    const matchesStatus = statusFilter === "all" || machine.machine_status?.status === statusFilter;
    return matchesSearch && matchesStatus;
  }), [machines, search, statusFilter]);

  async function saveMachine(values) {
    setBusy(true);
    try {
      if (formModal?.machine) {
        await api.updateMachine(token, formModal.machine.id, values.name, values.type);
        notify("success", "Máquina atualizada com sucesso.");
      } else {
        const created = await api.createMachine(token, values.name, values.type);
        notify("success", `Máquina ${created.name} criada. A chave API já está disponível.`);
        setSelected(created);
      }
      setFormModal(null);
      await refresh();
    } catch (error) {
      notify("error", error.message);
    } finally {
      setBusy(false);
    }
  }

  async function deleteMachine() {
    if (!deleting) return;
    setBusy(true);
    try {
      await api.deleteMachine(token, deleting.id);
      if (selected?.id === deleting.id) setSelected(null);
      notify("success", `Máquina ${deleting.name} excluída.`);
      setDeleting(null);
      await refresh();
    } catch (error) {
      notify("error", error.message);
    } finally {
      setBusy(false);
    }
  }

  async function copyKey(machine) {
    try {
      await navigator.clipboard.writeText(machine.api_key);
      setCopiedId(machine.id);
      notify("success", "Chave API copiada.");
      window.setTimeout(() => setCopiedId(""), 1800);
    } catch {
      notify("error", "Não foi possível copiar a chave.");
    }
  }

  function toggleKey(id) {
    setVisibleKeys((current) => {
      const next = new Set(current);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  }

  return (
    <main className="page machines-page">
      <header className="page-heading">
        <div><p className="eyebrow">Gestão de ativos</p><h1>Máquinas</h1><p>Cadastre equipamentos e acompanhe a condição de cada ativo.</p></div>
        <button type="button" className="button button-primary" onClick={() => setFormModal({ machine: null })}><Plus size={18} /> Nova máquina</button>
      </header>

      <section className="panel list-panel">
        <div className="list-toolbar">
          <div className="search-field"><Search size={18} /><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Buscar por nome ou tipo" aria-label="Buscar máquinas" />{search && <button type="button" onClick={() => setSearch("")} aria-label="Limpar busca"><X size={16} /></button>}</div>
          <select value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)} aria-label="Filtrar por condição">
            <option value="all">Todas as condições</option>
            <option value="NORMAL">Normal</option>
            <option value="DEGRADATION">Atenção</option>
            <option value="FAILURE">Crítico</option>
            <option value="AGUARDANDO_DADOS">Aguardando dados</option>
          </select>
          <button type="button" className="icon-button bordered" onClick={() => refresh(true)} disabled={refreshing} aria-label="Atualizar máquinas" title="Atualizar">
            <RefreshCw className={refreshing ? "spin" : ""} size={18} />
          </button>
          <span className="result-count">{filtered.length} de {machines.length}</span>
        </div>

        {filtered.length ? (
          <div className="data-table-wrap">
            <table className="data-table machines-table">
              <thead><tr><th>Máquina</th><th>Condição</th><th>Saúde</th><th>Risco</th><th>Última atualização</th><th><span className="sr-only">Ações</span></th></tr></thead>
              <tbody>
                {filtered.map((machine) => {
                  const status = statusLabels[machine.machine_status?.status] || statusLabels.AGUARDANDO_DADOS;
                  return (
                    <tr key={machine.id} onClick={() => setSelected(machine)} className="clickable-row">
                      <td><span className="machine-cell"><span className="machine-type-icon"><Boxes size={17} /></span><span><strong>{machine.name}</strong><small>{typeLabels[machine.type] || machine.type}</small></span></span></td>
                      <td><span className={`status-badge ${status[1]}`}><i />{status[0]}</span></td>
                      <td>{formatScore(machine.machine_status?.health_score)}</td>
                      <td>{formatScore(machine.machine_status?.risk_score)}</td>
                      <td>{formatDate(machine.machine_status?.updated_at)}</td>
                      <td><button type="button" className="icon-button subtle" onClick={(event) => { event.stopPropagation(); setSelected(machine); }} aria-label={`Detalhes de ${machine.name}`} title="Ver detalhes"><MoreHorizontal size={19} /></button></td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="empty-state"><span><Boxes size={27} /></span><h3>{machines.length ? "Nenhum resultado" : "Cadastre sua primeira máquina"}</h3><p>{machines.length ? "Tente ajustar os filtros da busca." : "Adicione um ativo para iniciar o monitoramento inteligente."}</p>{!machines.length && <button type="button" className="button button-primary" onClick={() => setFormModal({ machine: null })}><Plus size={17} /> Nova máquina</button>}</div>
        )}
      </section>

      {formModal && <MachineForm machine={formModal.machine} busy={busy} onSubmit={saveMachine} onClose={() => setFormModal(null)} />}
      {selected && (
        <MachineDetails machine={machines.find((item) => item.id === selected.id) || selected} visible={visibleKeys.has(selected.id)} copied={copiedId === selected.id} onToggleKey={() => toggleKey(selected.id)} onCopy={() => copyKey(selected)} onEdit={() => setFormModal({ machine: selected })} onDelete={() => setDeleting(selected)} onClose={() => setSelected(null)} />
      )}
      {deleting && <ConfirmDialog title="Excluir máquina?" message={`Todas as telemetrias de “${deleting.name}” também serão excluídas. Esta ação não pode ser desfeita.`} onConfirm={deleteMachine} onClose={() => setDeleting(null)} loading={busy} />}
    </main>
  );
}

function MachineForm({ machine, busy, onSubmit, onClose }) {
  const [name, setName] = useState(machine?.name || "");
  const [type, setType] = useState(machine?.type || "motor");
  return (
    <Modal title={machine ? "Editar máquina" : "Nova máquina"} description={machine ? "Atualize a identificação do ativo." : "Cadastre o equipamento que será monitorado."} onClose={onClose}>
      <form onSubmit={(event) => { event.preventDefault(); onSubmit({ name, type }); }}>
        <div className="form-stack">
          <label>Nome da máquina<input value={name} onChange={(event) => setName(event.target.value)} placeholder="Ex.: Motor da esteira 01" autoFocus required /></label>
          <label>Tipo do equipamento<select value={type} onChange={(event) => setType(event.target.value)}><option value="motor">Motor</option><option value="bomba">Bomba</option><option value="compressor">Compressor</option></select></label>
        </div>
        <div className="modal-actions"><button type="button" className="button button-secondary" onClick={onClose}>Cancelar</button><button type="submit" className="button button-primary" disabled={busy || !name.trim()}>{busy ? "Salvando..." : machine ? "Salvar alterações" : "Criar máquina"}</button></div>
      </form>
    </Modal>
  );
}

function MachineDetails({ machine, visible, copied, onToggleKey, onCopy, onEdit, onDelete, onClose }) {
  const status = statusLabels[machine.machine_status?.status] || statusLabels.AGUARDANDO_DADOS;
  return (
    <div className="drawer-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <aside className="details-drawer" role="dialog" aria-modal="true" aria-labelledby="machine-details-title">
        <header><div><p className="eyebrow">Detalhes do ativo</p><h2 id="machine-details-title">{machine.name}</h2></div><button type="button" className="icon-button" onClick={onClose} aria-label="Fechar"><X size={20} /></button></header>
        <div className="drawer-body">
          <div className="detail-status"><span className={`status-badge ${status[1]}`}><i />{status[0]}</span><small>Atualizado em {formatDate(machine.machine_status?.updated_at)}</small></div>
          <div className="score-pair"><div><span>Saúde</span><strong>{formatScore(machine.machine_status?.health_score)}</strong><div className="progress health"><i style={{ width: formatScore(machine.machine_status?.health_score) === "—" ? 0 : formatScore(machine.machine_status.health_score) }} /></div></div><div><span>Risco</span><strong>{formatScore(machine.machine_status?.risk_score)}</strong><div className="progress risk"><i style={{ width: formatScore(machine.machine_status?.risk_score) === "—" ? 0 : formatScore(machine.machine_status.risk_score) }} /></div></div></div>
          <dl className="details-list"><div><dt>Tipo</dt><dd>{typeLabels[machine.type] || machine.type}</dd></div><div><dt>Criada em</dt><dd>{formatDate(machine.created_at)}</dd></div><div><dt>ID da máquina</dt><dd className="mono">{machine.id}</dd></div></dl>
          <section className="api-key-box"><div><span>Chave API da máquina</span><small>Use esta chave para autenticar o envio de telemetria.</small></div><div className="key-value"><code>{visible ? machine.api_key : `••••••••••••••••${machine.api_key?.slice(-6) || ""}`}</code><button type="button" className="icon-button" onClick={onToggleKey} aria-label={visible ? "Ocultar chave" : "Mostrar chave"} title={visible ? "Ocultar chave" : "Mostrar chave"}>{visible ? <EyeOff size={18} /> : <Eye size={18} />}</button><button type="button" className="icon-button" onClick={onCopy} aria-label="Copiar chave" title="Copiar chave">{copied ? <ClipboardCheck size={18} /> : <Clipboard size={18} />}</button></div><p>Trate esta chave como uma senha e não a compartilhe publicamente.</p></section>
        </div>
        <footer><button type="button" className="button button-danger-quiet" onClick={onDelete}><Trash2 size={17} /> Excluir</button><button type="button" className="button button-secondary" onClick={onEdit}><Edit3 size={17} /> Editar</button></footer>
      </aside>
    </div>
  );
}
