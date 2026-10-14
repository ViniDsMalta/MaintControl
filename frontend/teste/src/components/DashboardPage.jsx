import { Activity, AlertTriangle, ArrowRight, Boxes, Network, RefreshCw, ShieldCheck } from "lucide-react";

function number(value) {
  return typeof value === "number" ? value : null;
}

function score(value) {
  return number(value) === null ? "—" : Math.round(value);
}

function timeAgo(value) {
  if (!value) return "Sem dados";
  const diff = Math.max(0, Date.now() - new Date(value).getTime());
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "Agora";
  if (minutes < 60) return `${minutes} min atrás`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h atrás`;
  return new Date(value).toLocaleDateString("pt-BR");
}

function statusInfo(status) {
  const map = {
    NORMAL: { label: "Normal", tone: "normal" },
    DEGRADATION: { label: "Atenção", tone: "warning" },
    FAILURE: { label: "Crítico", tone: "critical" },
    AGUARDANDO_DADOS: { label: "Aguardando", tone: "waiting" },
  };
  return map[status] || map.AGUARDANDO_DADOS;
}

export function DashboardPage({ machines, lines, lineDetails, refreshing, refresh, navigate }) {
  const withPrediction = machines.filter((machine) => number(machine.machine_status?.health_score) !== null);
  const averageHealth = withPrediction.length
    ? withPrediction.reduce((total, machine) => total + machine.machine_status.health_score, 0) / withPrediction.length
    : null;
  const critical = machines.filter((machine) => machine.machine_status?.status === "FAILURE").length;
  const warning = machines.filter((machine) => machine.machine_status?.status === "DEGRADATION").length;
  const normal = machines.filter((machine) => machine.machine_status?.status === "NORMAL").length;
  const waiting = machines.length - critical - warning - normal;
  const conditionRows = [
    { label: "Operando normalmente", value: normal, tone: "normal" },
    { label: "Sinais de degradação", value: warning, tone: "warning" },
    { label: "Falha provável", value: critical, tone: "critical" },
    { label: "Sem dados suficientes", value: waiting, tone: "waiting" },
  ];
  const recentMachines = [...machines]
    .sort((a, b) => new Date(b.machine_status?.updated_at || 0) - new Date(a.machine_status?.updated_at || 0))
    .slice(0, 5);

  return (
    <main className="page dashboard-page">
      <header className="page-heading">
        <div><p className="eyebrow">Operação em tempo real</p><h1>Visão geral</h1><p>Acompanhe a condição dos ativos e os pontos que exigem atenção.</p></div>
        <button type="button" className="button button-secondary" onClick={refresh} disabled={refreshing}>
          <RefreshCw className={refreshing ? "spin" : ""} size={17} /> Atualizar dados
        </button>
      </header>

      <section className="metrics-grid" aria-label="Indicadores principais">
        <article className="metric-card">
          <span className="metric-icon orange"><Boxes size={20} /></span>
          <div><p>Máquinas monitoradas</p><strong>{machines.length}</strong><small>{withPrediction.length} enviando dados</small></div>
        </article>
        <article className="metric-card">
          <span className="metric-icon green"><ShieldCheck size={20} /></span>
          <div><p>Saúde média</p><strong>{averageHealth === null ? "—" : `${Math.round(averageHealth)}%`}</strong><small>{averageHealth === null ? "Aguardando telemetria" : "Média dos ativos ativos"}</small></div>
        </article>
        <article className="metric-card">
          <span className="metric-icon red"><AlertTriangle size={20} /></span>
          <div><p>Exigem atenção</p><strong>{critical + warning}</strong><small>{critical} crítico(s), {warning} em atenção</small></div>
        </article>
        <article className="metric-card">
          <span className="metric-icon blue"><Network size={20} /></span>
          <div><p>Fluxos produtivos</p><strong>{lines.length}</strong><small>{Object.values(lineDetails).reduce((sum, detail) => sum + (detail?.machines?.length || 0), 0)} vínculos ativos</small></div>
        </article>
      </section>

      <section className="dashboard-grid">
        <article className="panel condition-panel">
          <div className="panel-heading"><div><h2>Condição da operação</h2><p>Distribuição atual dos ativos</p></div><Activity size={20} /></div>
          <div className="condition-content">
            <div className="donut" style={{ "--normal": normal, "--warning": warning, "--critical": critical, "--total": Math.max(machines.length, 1) }}>
              <div><strong>{machines.length}</strong><span>ativos</span></div>
            </div>
            <ul className="legend-list">
              <li><span className="legend-dot normal" /><span>Normal</span><strong>{normal}</strong></li>
              <li><span className="legend-dot warning" /><span>Atenção</span><strong>{warning}</strong></li>
              <li><span className="legend-dot critical" /><span>Crítico</span><strong>{critical}</strong></li>
              <li><span className="legend-dot waiting" /><span>Aguardando dados</span><strong>{waiting}</strong></li>
            </ul>
            <div className="condition-breakdown">
              {conditionRows.map((item) => {
                const percentage = machines.length ? Math.round((item.value / machines.length) * 100) : 0;
                return (
                  <div key={item.tone}>
                    <span><b>{item.label}</b><strong>{percentage}%</strong></span>
                    <div className={`condition-bar ${item.tone}`}><i style={{ width: `${percentage}%` }} /></div>
                  </div>
                );
              })}
            </div>
          </div>
        </article>
      </section>

      <section className="panel assets-panel">
        <div className="panel-heading">
          <div><h2>Máquinas recentes</h2><p>Última leitura recebida por ativo</p></div>
          <button type="button" className="text-button" onClick={() => navigate("machines")}>Ver todas <ArrowRight size={16} /></button>
        </div>
        {recentMachines.length ? (
          <div className="data-table-wrap">
            <table className="data-table">
              <thead><tr><th>Máquina</th><th>Condição</th><th>Saúde</th><th>Risco</th><th>Última leitura</th></tr></thead>
              <tbody>
                {recentMachines.map((machine) => {
                  const status = statusInfo(machine.machine_status?.status);
                  return (
                    <tr key={machine.id}>
                      <td><span className="machine-cell"><span className="machine-type-icon"><Boxes size={17} /></span><span><strong>{machine.name}</strong><small>{machine.type}</small></span></span></td>
                      <td><span className={`status-badge ${status.tone}`}><i />{status.label}</span></td>
                      <td><span className="score-cell"><span><i style={{ width: `${score(machine.machine_status?.health_score) === "—" ? 0 : score(machine.machine_status.health_score)}%` }} /></span><b>{score(machine.machine_status?.health_score)}{score(machine.machine_status?.health_score) === "—" ? "" : "%"}</b></span></td>
                      <td>{score(machine.machine_status?.risk_score)}{score(machine.machine_status?.risk_score) === "—" ? "" : "%"}</td>
                      <td>{timeAgo(machine.machine_status?.updated_at)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="empty-state compact"><span><Boxes size={24} /></span><h3>Nenhuma máquina cadastrada</h3><p>Cadastre o primeiro ativo para começar o monitoramento.</p><button type="button" className="button button-primary" onClick={() => navigate("machines")}>Cadastrar máquina</button></div>
        )}
      </section>
    </main>
  );
}
