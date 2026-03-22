import { useEffect, useMemo, useState, useCallback } from 'react';
import DeviceModal from './DeviceModal';
import { fetchDevices, deleteDevice, updateDevice } from './device';
import type { Device } from './device';
import './styles.css';

const DEBOUNCE = 600; // чтобы не дёргать API на каждый символ
type ToastKind = 'ok' | 'err';

export default function App() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [debounced, setDebounced] = useState('');
  const [activeOnly, setActiveOnly] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Device | undefined>();
  const [toast, setToast] = useState<{ text: string; kind: ToastKind } | null>(null);

  useEffect(() => {
    const t = setTimeout(() => setDebounced(search.trim()), DEBOUNCE);
    return () => clearTimeout(t);
  }, [search]);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), 2500);
    return () => clearTimeout(t);
  }, [toast]);

  const query = useMemo(() => {
    const p = new URLSearchParams();
    if (debounced) p.set('search', debounced);
    if (activeOnly) p.set('is_active', 'true');
    const s = p.toString();
    return s ? `?${s}` : '';
  }, [activeOnly, debounced]);

  // загрузка списка, пересоздаётся при смене фильтров
  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      setDevices(await fetchDevices(query));
    } catch {
      setError('Не удалось загрузить');
    } finally {
      setLoading(false);
    }
  }, [query]);

  useEffect(() => { void load(); }, [load]);

  function notify(text: string, kind: ToastKind) { setToast({ text, kind }); }

  async function handleDeactivate(id: number) {
    try { await deleteDevice(id); await load(); notify('Деактивировано', 'ok'); }
    catch { notify('Ошибка', 'err'); }
  }

  // переключение is_active без удаления
  async function handleToggle(d: Device) {
    try {
      await updateDevice(d.id, { hostname: d.hostname, ip: d.ip, location: d.location, is_active: !d.is_active });
      await load();
      notify(d.is_active ? 'Деактивировано' : 'Активировано', 'ok');
    } catch { notify('Ошибка', 'err'); }
  }

  function openModal(d?: Device) { setEditing(d); setModalOpen(true); }

  async function onDone(msg: string) {
    setModalOpen(false); setEditing(undefined); await load(); notify(msg, 'ok');
  }
  function onClose() { setModalOpen(false); setEditing(undefined); void load(); }

  return (
    <main className="page">
      {toast && <div className={`toast ${toast.kind === 'ok' ? 'toast-ok' : 'toast-err'}`}>{toast.text}</div>}

      <div className="header">
        <div className="header-left">
          <h1>Device Inventory</h1>
          <p>Управление сетевым оборудованием</p>
        </div>
        <div className="header-right">
          <span className="count">{devices.length} устройств</span>
          <button className="btn btn-primary" onClick={() => openModal()}>+ Добавить</button>
        </div>
      </div>

      <div className="toolbar">
        <input type="text" placeholder="Поиск по hostname..." value={search}
          onChange={e => setSearch(e.target.value)} />
        <label>
          <input type="checkbox" checked={activeOnly} onChange={e => setActiveOnly(e.target.checked)} />
          Только активные
        </label>
      </div>

      {error && <div className="error-msg">{error}</div>}

      {loading ? (
        <div className="card"><div className="loading-box">Загрузка...</div></div>
      ) : (
        <div className="card">
          <table>
            <thead>
              <tr>
                <th className="c-center" style={{ width: 44 }}>ID</th>
                <th>Hostname</th>
                <th>IP</th>
                <th>Location</th>
                <th className="c-center">Status</th>
                <th>Created</th>
                <th className="c-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {devices.length === 0 ? (
                <tr><td colSpan={7} className="empty-cell">Нет устройств</td></tr>
              ) : devices.map(d => (
                <tr key={d.id}>
                  <td className="c-id">{d.id}</td>
                  <td className="c-host">{d.hostname}</td>
                  <td className="c-ip">{d.ip}</td>
                  <td>{d.location}</td>
                  <td className="c-center">
                    <span className={`status ${d.is_active ? 's-on' : 's-off'}`}>
                      {d.is_active ? 'active' : 'inactive'}
                    </span>
                  </td>
                  <td className="c-date">{new Date(d.created_at).toLocaleDateString()}</td>
                  <td className="c-actions">
                    <span className="btns">
                      <button className="btn btn-ghost btn-sm" onClick={() => openModal(d)}>Изменить</button>
                      {d.is_active
                        ? <button className="btn btn-ghost btn-sm btn-red" onClick={() => handleDeactivate(d.id)}>Выкл</button>
                        : <button className="btn btn-ghost btn-sm btn-green" onClick={() => handleToggle(d)}>Вкл</button>
                      }
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <DeviceModal isOpen={modalOpen} closeModal={onClose}
        deviceToEdit={editing} onSaved={onDone} showToast={notify} />
    </main>
  );
}
