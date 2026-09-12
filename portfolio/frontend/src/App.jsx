import { useCallback, useEffect, useState } from 'react';
import { api, samples } from './api.js';

const emptyForm = { id: '', title: '', description: '', tech: '', url: '', repo: '', featured: false };

function useProjects(view) {
  const [projects, setProjects] = useState([]);
  const [status, setStatus] = useState('loading');
  const [offline, setOffline] = useState(false);

  const reload = useCallback(async () => {
    setStatus('loading');
    try {
      const data = await api.list();
      setProjects(data);
      setOffline(false);
      setStatus('ready');
    } catch {
      setProjects(samples);
      setOffline(true);
      setStatus('ready');
    }
  }, []);

  useEffect(() => {
    if (view === 'home' || view === 'admin') reload();
  }, [view, reload]);

  return { projects, status, offline, reload };
}

export default function App() {
  const [view, setView] = useState('home');
  const [selectedId, setSelectedId] = useState(null);
  const [detail, setDetail] = useState(null);
  const { projects, status, offline, reload } = useProjects(view === 'detail' ? 'home' : view);

  const [token, setToken] = useState(() => localStorage.getItem('admin_token') || '');
  const [form, setForm] = useState(emptyForm);
  const [editing, setEditing] = useState(null);
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  const openDetail = async (id) => {
    setSelectedId(id);
    setView('detail');
    try {
      setDetail(await api.get(id));
    } catch {
      setDetail(samples.find((p) => p.id === id) || null);
    }
  };

  const saveToken = (v) => {
    setToken(v);
    localStorage.setItem('admin_token', v);
  };

  const toPayload = (f) => ({
    id: f.id.trim(),
    title: f.title.trim(),
    description: f.description.trim(),
    tech: f.tech.split(',').map((s) => s.trim()).filter(Boolean),
    url: f.url.trim(),
    repo: f.repo.trim(),
    featured: !!f.featured,
  });

  const onSubmit = async (e) => {
    e.preventDefault();
    setError('');
    if (!token) {
      setError('Informe o token admin para salvar.');
      return;
    }
    setSaving(true);
    try {
      if (editing) {
        await api.update(editing, toPayload({ ...form, id: editing }), token);
      } else {
        await api.create(toPayload(form), token);
      }
      setForm(emptyForm);
      setEditing(null);
      await reload();
      setView('admin');
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const onEdit = (p) => {
    setEditing(p.id);
    setForm({
      id: p.id,
      title: p.title || '',
      description: p.description || '',
      tech: (p.tech || []).join(', '),
      url: p.url || '',
      repo: p.repo || '',
      featured: !!p.featured,
    });
    window.scrollTo({ top: 0 });
  };

  const onDelete = async (id) => {
    if (!confirm(`Excluir "${id}"?`)) return;
    setError('');
    try {
      await api.remove(id, token);
      if (editing === id) {
        setEditing(null);
        setForm(emptyForm);
      }
      await reload();
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="wrap">
      <header className="topbar">
        <div className="brand">
          portfólio<span>.</span>
        </div>
        <nav className="nav">
          <button className="btn" onClick={() => setView('home')}>Projetos</button>
          <button className="btn" onClick={() => setView('admin')}>Admin</button>
        </nav>
      </header>

      {view === 'home' && (
        <>
          <section className="hero">
            <h1>Projetos em Go e React, do estudo ao deploy.</h1>
            <p>
              Registro público do site de investimentos, do site de mangá e dos próximos
              projetos. Cada item tem stack, descrição e links.
            </p>
            <div className="meta">Go · React · Docker — 4 vCPU / 8 GB</div>
            {offline && <p className="hint">API indisponível — mostrando cópia local.</p>}
          </section>

          <div className="section-title">Todos os projetos</div>
          {status === 'loading' ? (
            <p className="hint">Carregando…</p>
          ) : (
            <div className="grid">
              {projects.map((p) => (
                <article className="card" key={p.id}>
                  <h3>{p.title}</h3>
                  <p>{p.description || 'Sem descrição.'}</p>
                  <div className="tags">{(p.tech || []).map((t) => <span className="tag" key={t}>{t}</span>)}</div>
                  <div className="row">
                    <button className="btn" onClick={() => openDetail(p.id)}>Detalhe</button>
                    {p.url && <a className="btn" href={p.url} target="_blank" rel="noreferrer">Abrir</a>}
                  </div>
                </article>
              ))}
            </div>
          )}
        </>
      )}

      {view === 'detail' && (
        <>
          <button className="btn" onClick={() => setView('home')}>← Voltar</button>
          <div style={{ height: 12 }} />
          {!detail ? (
            <p className="hint">Projeto não encontrado.</p>
          ) : (
            <article className="detail">
              <h2>{detail.title}</h2>
              <p>{detail.description}</p>
              <dl className="kv">
                <dt>ID</dt><dd>{detail.id}</dd>
                <dt>Stack</dt><dd>{(detail.tech || []).join(', ') || '—'}</dd>
                <dt>Site</dt><dd>{detail.url ? <a href={detail.url}>{detail.url}</a> : '—'}</dd>
                <dt>Repo</dt><dd>{detail.repo ? <a href={detail.repo}>{detail.repo}</a> : '—'}</dd>
              </dl>
              <div className="tags">{(detail.tech || []).map((t) => <span className="tag" key={t}>{t}</span>)}</div>
            </article>
          )}
        </>
      )}

      {view === 'admin' && (
        <>
          <h2 style={{ margin: '4px 0 6px' }}>Admin — registrar projetos</h2>
          <p className="hint">
            Escritas exigem o mesmo <code>ADMIN_TOKEN</code> do backend. O token fica só no seu navegador.
          </p>
          <label style={{ display: 'grid', gap: 6, fontSize: 14, maxWidth: 420 }}>
            Token admin
            <input
              type="password"
              value={token}
              onChange={(e) => saveToken(e.target.value)}
              placeholder="Bearer token"
              style={{ background: '#1c2329', border: '1px solid #2a3238', borderRadius: 8, color: '#e9e6df', padding: '9px 11px' }}
            />
          </label>

          {error && <p className="error">{error}</p>}

          <form className="form" onSubmit={onSubmit}>
            <div className="two">
              <label>ID (slug) {!editing && <input value={form.id} onChange={(e) => setForm({ ...form, id: e.target.value })} placeholder="site-investimentos" disabled={!!editing} />}</label>
              <label>Destaque <input type="checkbox" checked={form.featured} onChange={(e) => setForm({ ...form, featured: e.target.checked })} /></label>
            </div>
            <label>Título <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="Site de Investimentos" /></label>
            <label>Descrição <textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="O que o projeto faz" /></label>
            <label>Stack (vírgula) <input value={form.tech} onChange={(e) => setForm({ ...form, tech: e.target.value })} placeholder="Go, React, Docker" /></label>
            <div className="two">
              <label>URL <input value={form.url} onChange={(e) => setForm({ ...form, url: e.target.value })} placeholder="https://…" /></label>
              <label>Repo <input value={form.repo} onChange={(e) => setForm({ ...form, repo: e.target.value })} placeholder="https://github.com/…" /></label>
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button className="btn primary" type="submit" disabled={saving}>{saving ? 'Salvando…' : editing ? 'Atualizar' : 'Adicionar'}</button>
              {editing && <button type="button" className="btn" onClick={() => { setEditing(null); setForm(emptyForm); }}>Cancelar</button>}
            </div>
          </form>

          <div className="section-title">Registrados ({projects.length})</div>
          <div className="admin-list">
            {projects.map((p) => (
              <div className="admin-row" key={p.id}>
                <span><strong>{p.title}</strong> <span className="hint">· {p.id}</span></span>
                <span className="actions">
                  <button className="btn" onClick={() => onEdit(p)}>Editar</button>
                  <button className="btn danger" onClick={() => onDelete(p.id)}>Excluir</button>
                </span>
              </div>
            ))}
          </div>
        </>
      )}

      <footer>
        <span>Portfólio — Go + React + Docker · TDD no backend</span>
        <span>API: {offline ? 'offline (local)' : 'conectada'}</span>
      </footer>
    </div>
  );
}
