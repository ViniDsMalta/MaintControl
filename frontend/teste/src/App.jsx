import { useCallback, useEffect, useMemo, useState } from "react";
import { CheckCircle2, XCircle } from "lucide-react";
import { api } from "./api/api";
import { AppShell } from "./components/AppShell";
import { AuthPage } from "./components/AuthPage";
import { DashboardPage } from "./components/DashboardPage";
import { FlowsPage } from "./components/FlowsPage";
import { MachinesPage } from "./components/MachinesPage";

const TOKEN_KEY = "maintcontrol_test_token";
const VALID_PAGES = new Set(["dashboard", "machines", "flows"]);

function pageFromHash() {
  const value = window.location.hash.replace("#", "");
  return VALID_PAGES.has(value) ? value : "dashboard";
}

export default function App() {
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || "");
  const [user, setUser] = useState(null);
  const [page, setPage] = useState(pageFromHash);
  const [machines, setMachines] = useState([]);
  const [lines, setLines] = useState([]);
  const [lineDetails, setLineDetails] = useState({});
  const [booting, setBooting] = useState(Boolean(token));
  const [refreshing, setRefreshing] = useState(false);
  const [toast, setToast] = useState(null);

  const notify = useCallback((type, message) => {
    setToast({ id: Date.now(), type, message });
  }, []);

  useEffect(() => {
    if (!toast) return undefined;
    const timeout = window.setTimeout(() => setToast(null), 4200);
    return () => window.clearTimeout(timeout);
  }, [toast]);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setToken("");
    setUser(null);
    setMachines([]);
    setLines([]);
    setLineDetails({});
  }, []);

  const handleError = useCallback((error) => {
    if (error.status === 401) logout();
    notify("error", error.message);
  }, [logout, notify]);

  const refreshMachines = useCallback(async (showLoader = false) => {
    if (!token) return;
    if (showLoader) setRefreshing(true);
    try {
      setMachines(await api.getMachines(token));
    } catch (error) {
      handleError(error);
    } finally {
      if (showLoader) setRefreshing(false);
    }
  }, [token, handleError]);

  const refreshLines = useCallback(async () => {
    if (!token) return;
    try {
      const nextLines = await api.getProductionLines(token);
      const detailEntries = await Promise.all(
        nextLines.map(async (line) => [line.id, await api.getProductionLine(token, line.id)]),
      );
      setLines(nextLines);
      setLineDetails(Object.fromEntries(detailEntries));
    } catch (error) {
      handleError(error);
    }
  }, [token, handleError]);

  const refreshAll = useCallback(async () => {
    setRefreshing(true);
    await Promise.all([refreshMachines(), refreshLines()]);
    setRefreshing(false);
  }, [refreshMachines, refreshLines]);

  useEffect(() => {
    const onHashChange = () => setPage(pageFromHash());
    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
  }, []);

  useEffect(() => {
    if (!token) {
      setBooting(false);
      return;
    }
    setBooting(true);
    Promise.all([api.getMe(token), api.getMachines(token), api.getProductionLines(token)])
      .then(async ([currentUser, currentMachines, currentLines]) => {
        setUser(currentUser);
        setMachines(currentMachines);
        setLines(currentLines);
        const details = await Promise.all(
          currentLines.map(async (line) => [line.id, await api.getProductionLine(token, line.id)]),
        );
        setLineDetails(Object.fromEntries(details));
      })
      .catch(handleError)
      .finally(() => setBooting(false));
  }, [token, handleError]);

  useEffect(() => {
    if (!token) return undefined;
    const interval = window.setInterval(() => refreshMachines(), 5000);
    return () => window.clearInterval(interval);
  }, [token, refreshMachines]);

  function authenticated(result) {
    localStorage.setItem(TOKEN_KEY, result.token);
    setToken(result.token);
    setUser(result.user);
    window.location.hash = "dashboard";
    notify("success", `Bem-vindo, ${result.user.name.split(" ")[0]}.`);
  }

  function navigate(nextPage) {
    window.location.hash = nextPage;
    setPage(nextPage);
  }

  const pageContent = useMemo(() => {
    if (page === "machines") {
      return <MachinesPage token={token} machines={machines} refreshing={refreshing} refresh={refreshMachines} notify={notify} />;
    }
    if (page === "flows") {
      return <FlowsPage token={token} machines={machines} lines={lines} lineDetails={lineDetails} refresh={refreshLines} notify={notify} />;
    }
    return <DashboardPage machines={machines} lines={lines} lineDetails={lineDetails} refreshing={refreshing} refresh={refreshAll} navigate={navigate} />;
  }, [page, token, machines, lines, lineDetails, refreshing, refreshMachines, refreshLines, refreshAll, notify]);

  if (!token) return <AuthPage onAuthenticated={authenticated} />;

  return (
    <>
      <AppShell user={user} page={page} navigate={navigate} logout={logout} booting={booting}>
        {pageContent}
      </AppShell>
      {toast && (
        <div className={`toast toast-${toast.type}`} role="status">
          {toast.type === "success" ? <CheckCircle2 size={18} /> : <XCircle size={18} />}
          <span>{toast.message}</span>
          <button type="button" aria-label="Fechar aviso" onClick={() => setToast(null)}>×</button>
        </div>
      )}
    </>
  );
}
