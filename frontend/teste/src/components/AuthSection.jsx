import { useState } from "react";
import { api } from "../api/api";

export function AuthSection({ token, user, onAuthenticated, onLogout, onActivity }) {
  const [form, setForm] = useState({ name: "", email: "", password: "" });
  const [loading, setLoading] = useState("");

  function updateField(event) {
    setForm((current) => ({ ...current, [event.target.name]: event.target.value }));
  }

  async function authenticate(mode) {
    setLoading(mode);
    try {
      const result = mode === "register"
        ? await api.register(form.name, form.email, form.password)
        : await api.login(form.email, form.password);
      onAuthenticated(result);
      onActivity("success", mode === "register" ? "Usuário registrado." : "Login realizado.");
    } catch (error) {
      onActivity("error", error.message);
    } finally {
      setLoading("");
    }
  }

  return (
    <section>
      <div className="section-heading">
        <h2>Autenticação</h2>
        <span>{token ? "Sessão ativa" : "Não autenticado"}</span>
      </div>

      {user ? (
        <div className="authenticated-user">
          <strong>{user.name}</strong>
          <span>{user.email}</span>
          <button type="button" className="secondary" onClick={onLogout}>Sair</button>
        </div>
      ) : (
        <form className="form-grid" onSubmit={(event) => event.preventDefault()}>
          <label>
            Nome
            <input name="name" value={form.name} onChange={updateField} autoComplete="name" />
          </label>
          <label>
            E-mail
            <input name="email" type="email" value={form.email} onChange={updateField} autoComplete="email" required />
          </label>
          <label>
            Senha
            <input name="password" type="password" value={form.password} onChange={updateField} autoComplete="current-password" required />
          </label>
          <div className="form-actions">
            <button type="button" onClick={() => authenticate("register")} disabled={Boolean(loading)}>
              {loading === "register" ? "Registrando..." : "Registrar"}
            </button>
            <button type="button" className="secondary" onClick={() => authenticate("login")} disabled={Boolean(loading)}>
              {loading === "login" ? "Entrando..." : "Entrar"}
            </button>
          </div>
        </form>
      )}
    </section>
  );
}
