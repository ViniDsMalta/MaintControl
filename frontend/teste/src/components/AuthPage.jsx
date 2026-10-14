import { useState } from "react";
import { Activity, ArrowRight, Check, Eye, EyeOff, LoaderCircle } from "lucide-react";
import { api } from "../api/api";

export function AuthPage({ onAuthenticated }) {
  const [mode, setMode] = useState("login");
  const [form, setForm] = useState({ name: "", email: "", password: "" });
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  function updateField(event) {
    setForm((current) => ({ ...current, [event.target.name]: event.target.value }));
    setError("");
  }

  async function submit(event) {
    event.preventDefault();
    setLoading(true);
    setError("");
    try {
      const result = mode === "register"
        ? await api.register(form.name, form.email, form.password)
        : await api.login(form.email, form.password);
      onAuthenticated(result);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setLoading(false);
    }
  }

  function switchMode(nextMode) {
    setMode(nextMode);
    setError("");
  }

  return (
    <main className="auth-page">
      <section className="auth-brand-panel">
        <div className="brand brand-light"><span className="brand-mark"><Activity size={22} /></span>MaintControl</div>
        <div className="auth-brand-copy">
          <p className="eyebrow">Inteligência industrial</p>
          <h1>Decisões antes da parada.</h1>
          <p>Monitore a saúde dos seus ativos, identifique riscos e organize toda a operação em um único lugar.</p>
          <div className="signal-visual" aria-hidden="true">
            {[36, 52, 42, 68, 58, 82, 72, 94, 66, 78, 54, 64].map((height, index) => (
              <span key={index} style={{ height: `${height}%` }} />
            ))}
          </div>
        </div>
        <ul className="auth-benefits">
          <li><Check size={16} /> Diagnóstico contínuo por IA</li>
          <li><Check size={16} /> Indicadores de saúde e risco</li>
          <li><Check size={16} /> Organização por fluxos produtivos</li>
        </ul>
      </section>

      <section className="auth-form-panel">
        <div className="auth-form-wrap">
          <div className="mobile-brand brand"><span className="brand-mark"><Activity size={21} /></span>MaintControl</div>
          <div className="auth-heading">
            <p className="eyebrow">{mode === "login" ? "Acesso à plataforma" : "Nova conta"}</p>
            <h2>{mode === "login" ? "Bem-vindo de volta" : "Comece agora"}</h2>
            <p>{mode === "login" ? "Entre para acompanhar sua operação." : "Crie sua conta para configurar os primeiros ativos."}</p>
          </div>

          <div className="auth-tabs" role="tablist" aria-label="Acesso">
            <button type="button" className={mode === "login" ? "active" : ""} onClick={() => switchMode("login")}>Entrar</button>
            <button type="button" className={mode === "register" ? "active" : ""} onClick={() => switchMode("register")}>Criar conta</button>
          </div>

          <form className="auth-form" onSubmit={submit}>
            {mode === "register" && (
              <label>Nome completo<input name="name" value={form.name} onChange={updateField} placeholder="Como devemos chamar você?" autoComplete="name" required /></label>
            )}
            <label>E-mail<input name="email" type="email" value={form.email} onChange={updateField} placeholder="voce@empresa.com" autoComplete="email" required /></label>
            <label>
              Senha
              <span className="password-field">
                <input name="password" type={showPassword ? "text" : "password"} value={form.password} onChange={updateField} placeholder="Mínimo de 6 caracteres" autoComplete={mode === "login" ? "current-password" : "new-password"} minLength="6" required />
                <button type="button" className="icon-button subtle" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? "Ocultar senha" : "Mostrar senha"} title={showPassword ? "Ocultar senha" : "Mostrar senha"}>
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </span>
            </label>
            {error && <p className="form-error" role="alert">{error}</p>}
            <button type="submit" className="button button-primary button-wide" disabled={loading}>
              {loading ? <LoaderCircle className="spin" size={18} /> : <>{mode === "login" ? "Entrar" : "Criar minha conta"}<ArrowRight size={18} /></>}
            </button>
          </form>
          <p className="auth-switch">
            {mode === "login" ? "Ainda não possui uma conta?" : "Já possui uma conta?"}
            <button type="button" onClick={() => switchMode(mode === "login" ? "register" : "login")}>
              {mode === "login" ? "Cadastre-se" : "Entrar"}
            </button>
          </p>
        </div>
      </section>
    </main>
  );
}
