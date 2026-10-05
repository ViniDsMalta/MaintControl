import { useCallback, useEffect, useState } from "react";
import { api } from "./api/api";
import { AuthSection } from "./components/AuthSection";
import { MachinesSection } from "./components/MachinesSection";
import { ProductionLinesSection } from "./components/ProductionLinesSection";
import { SystemStatus } from "./components/SystemStatus";

const TOKEN_KEY = "maintcontrol_test_token";
const POLLING_INTERVAL = 5000;

export default function App() {
  const [token, setToken] = useState(() => localStorage.getItem(TOKEN_KEY) || "");
  const [user, setUser] = useState(null);
  const [machines, setMachines] = useState([]);
  const [lines, setLines] = useState([]);
  const [machinesLoading, setMachinesLoading] = useState(false);
  const [backendHealth, setBackendHealth] = useState(null);
  const [activity, setActivity] = useState({ type: "", message: "" });

  const reportActivity = useCallback((type, message) => {
    setActivity({ type, message: `[${new Date().toLocaleTimeString("pt-BR")}] ${message}` });
  }, []);

  const refreshMachines = useCallback(async () => {
    if (!token) return;
    setMachinesLoading(true);
    try {
      setMachines(await api.getMachines(token));
    } catch (error) {
      reportActivity("error", error.message);
    } finally {
      setMachinesLoading(false);
    }
  }, [token, reportActivity]);

  const refreshLines = useCallback(async () => {
    if (!token) return;
    try {
      setLines(await api.getProductionLines(token));
    } catch (error) {
      reportActivity("error", error.message);
    }
  }, [token, reportActivity]);

  useEffect(() => {
    async function checkHealth() {
      try {
        setBackendHealth(await api.health());
      } catch {
        setBackendHealth(null);
      }
    }
    checkHealth();
    const interval = window.setInterval(checkHealth, 10000);
    return () => window.clearInterval(interval);
  }, []);

  useEffect(() => {
    if (!token) {
      setUser(null);
      setMachines([]);
      setLines([]);
      return;
    }
    Promise.all([api.getMe(token), api.getMachines(token), api.getProductionLines(token)])
      .then(([currentUser, currentMachines, currentLines]) => {
        setUser(currentUser);
        setMachines(currentMachines);
        setLines(currentLines);
      })
      .catch((error) => {
        reportActivity("error", error.message);
        logout();
      });
  }, [token, reportActivity]);

  useEffect(() => {
    if (!token) return undefined;
    const interval = window.setInterval(refreshMachines, POLLING_INTERVAL);
    return () => window.clearInterval(interval);
  }, [token, refreshMachines]);

  function authenticated(result) {
    localStorage.setItem(TOKEN_KEY, result.token);
    setToken(result.token);
    setUser(result.user);
  }

  function logout() {
    localStorage.removeItem(TOKEN_KEY);
    setToken("");
    setUser(null);
  }

  return (
    <main>
      <header className="page-header">
        <div><p className="eyebrow">Validação funcional</p><h1>MaintControl Test Panel</h1></div>
        <span>API: /api</span>
      </header>
      <AuthSection token={token} user={user} onAuthenticated={authenticated} onLogout={logout} onActivity={reportActivity} />
      <MachinesSection token={token} machines={machines} loading={machinesLoading} refresh={refreshMachines} onActivity={reportActivity} />
      <ProductionLinesSection token={token} lines={lines} machines={machines} refreshLines={refreshLines} onActivity={reportActivity} />
      <SystemStatus backendHealth={backendHealth} machines={machines} activity={activity} />
    </main>
  );
}
