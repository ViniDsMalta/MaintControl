import { Activity, Boxes, Gauge, LogOut, Menu, Network, X } from "lucide-react";
import { useState } from "react";

const navigation = [
  { id: "dashboard", label: "Visão geral", icon: Gauge },
  { id: "machines", label: "Máquinas", icon: Boxes },
  { id: "flows", label: "Fluxos", icon: Network },
];

export function AppShell({ user, page, navigate, logout, booting, children }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const initials = user?.name?.split(" ").slice(0, 2).map((part) => part[0]).join("").toUpperCase() || "MC";

  function goTo(nextPage) {
    navigate(nextPage);
    setMenuOpen(false);
  }

  return (
    <div className="app-shell">
      <aside className={`sidebar ${menuOpen ? "sidebar-open" : ""}`}>
        <div className="sidebar-top">
          <div className="brand brand-light"><span className="brand-mark"><Activity size={21} /></span>MaintControl</div>
          <button type="button" className="icon-button mobile-close" onClick={() => setMenuOpen(false)} aria-label="Fechar menu"><X size={20} /></button>
        </div>
        <nav className="main-nav" aria-label="Navegação principal">
          <p>Monitoramento</p>
          {navigation.map((item) => {
            const Icon = item.icon;
            return (
              <button key={item.id} type="button" className={page === item.id ? "active" : ""} onClick={() => goTo(item.id)}>
                <Icon size={19} /><span>{item.label}</span>
              </button>
            );
          })}
        </nav>
        <div className="sidebar-user">
          <span className="avatar">{initials}</span>
          <span className="user-copy"><strong>{user?.name || "Carregando..."}</strong><small>{user?.email || ""}</small></span>
          <button type="button" className="icon-button dark" onClick={logout} aria-label="Sair" title="Sair"><LogOut size={18} /></button>
        </div>
      </aside>
      {menuOpen && <button type="button" className="sidebar-backdrop" aria-label="Fechar menu" onClick={() => setMenuOpen(false)} />}
      <div className="app-content">
        <header className="mobile-header">
          <button type="button" className="icon-button" onClick={() => setMenuOpen(true)} aria-label="Abrir menu"><Menu size={22} /></button>
          <div className="brand"><span className="brand-mark"><Activity size={19} /></span>MaintControl</div>
          <span className="avatar small">{initials}</span>
        </header>
        {booting ? (
          <div className="page-loading"><span className="loader" /><p>Preparando sua operação...</p></div>
        ) : children}
      </div>
    </div>
  );
}
