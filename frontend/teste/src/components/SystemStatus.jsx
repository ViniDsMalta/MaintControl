export function SystemStatus({ backendHealth, machines, activity }) {
  const predictedMachines = machines.filter((machine) =>
    ["NORMAL", "DEGRADATION", "FAILURE"].includes(machine.machine_status?.status),
  ).length;

  return (
    <section>
      <div className="section-heading"><h2>Status do sistema</h2></div>
      <dl className="system-status">
        <div><dt>Backend</dt><dd>{backendHealth?.status === "ok" ? "Disponível" : "Indisponível"}</dd></div>
        <div><dt>PostgreSQL</dt><dd>{backendHealth?.database === "ok" ? "Disponível" : "Indisponível"}</dd></div>
        <div><dt>IA</dt><dd>{predictedMachines > 0 ? `Previsões recebidas (${predictedMachines})` : "Aguardando previsões"}</dd></div>
      </dl>
      <h3>Última atividade</h3>
      <pre className={activity.type === "error" ? "activity-error" : ""}>{activity.message || "Nenhuma atividade nesta sessão."}</pre>
    </section>
  );
}
