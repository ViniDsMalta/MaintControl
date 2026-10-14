import { X } from "lucide-react";
import { useEffect } from "react";

export function Modal({ title, description, children, onClose, size = "medium" }) {
  useEffect(() => {
    function closeOnEscape(event) {
      if (event.key === "Escape") onClose();
    }
    document.addEventListener("keydown", closeOnEscape);
    return () => document.removeEventListener("keydown", closeOnEscape);
  }, [onClose]);

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className={`modal modal-${size}`} role="dialog" aria-modal="true" aria-labelledby="modal-title">
        <header className="modal-header">
          <div><h2 id="modal-title">{title}</h2>{description && <p>{description}</p>}</div>
          <button type="button" className="icon-button" onClick={onClose} aria-label="Fechar"><X size={20} /></button>
        </header>
        {children}
      </section>
    </div>
  );
}

export function ConfirmDialog({ title, message, confirmLabel = "Excluir", onConfirm, onClose, loading }) {
  return (
    <Modal title={title} description={message} onClose={onClose} size="small">
      <div className="modal-actions">
        <button type="button" className="button button-secondary" onClick={onClose} disabled={loading}>Cancelar</button>
        <button type="button" className="button button-danger" onClick={onConfirm} disabled={loading}>{loading ? "Aguarde..." : confirmLabel}</button>
      </div>
    </Modal>
  );
}
