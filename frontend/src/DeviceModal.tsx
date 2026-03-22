import { useEffect, useState } from 'react';
import { createDevice, updateDevice } from './device';
import type { Device } from './device';

type ToastKind = 'ok' | 'err';

interface Props {
  isOpen: boolean;
  closeModal: () => void;
  deviceToEdit?: Device;
  onSaved: (msg: string) => void | Promise<void>;
  showToast: (text: string, kind: ToastKind) => void;
}

// простая проверка IPv4 на клиенте
function isIPv4(value: string): boolean {
  const parts = value.split('.');
  if (parts.length !== 4) return false;

  return parts.every(part => {
    if (part === '' || !/^\d+$/.test(part)) return false;
    const n = Number(part);
    return n >= 0 && n <= 255;
  });
}

export default function DeviceModal({ isOpen, closeModal, deviceToEdit, onSaved, showToast }: Props) {
  const [hostname, setHostname] = useState('');
  const [ip, setIp] = useState('');
  const [location, setLocation] = useState('');
  const [isActive, setIsActive] = useState(true);
  const [err, setErr] = useState('');
  const [saving, setSaving] = useState(false);

  // заполняем форму если редактируем
  useEffect(() => {
    if (deviceToEdit) {
      setHostname(deviceToEdit.hostname);
      setIp(deviceToEdit.ip);
      setLocation(deviceToEdit.location);
      setIsActive(deviceToEdit.is_active);
    } else {
      setHostname(''); setIp(''); setLocation(''); setIsActive(true);
    }
    setErr('');
  }, [deviceToEdit]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const data = { hostname: hostname.trim(), ip: ip.trim(), location: location.trim(), is_active: isActive };
    if (!data.hostname || !data.ip || !data.location) { setErr('Заполните все поля'); return; }
    if (data.hostname.length > 64) { setErr('Hostname слишком длинный'); return; }
    if (!isIPv4(data.ip)) { setErr('Некорректный IP'); return; }
    try {
      setSaving(true); setErr('');
      if (deviceToEdit) await updateDevice(deviceToEdit.id, data);
      else await createDevice(data);
      await onSaved(deviceToEdit ? 'Обновлено' : 'Добавлено');
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Ошибка сохранения';
      setErr(message);
      showToast(message, 'err');
    } finally { setSaving(false); }
  }

  if (!isOpen) return null;

  return (
    <div className="overlay" onClick={e => { if (e.target === e.currentTarget) closeModal(); }}>
      <div className="modal">
        <h2>{deviceToEdit ? 'Редактировать устройство' : 'Новое устройство'}</h2>
        <form onSubmit={handleSubmit}>
          <div className="field">
            <label>Hostname</label>
            <input value={hostname} onChange={e => setHostname(e.target.value)} placeholder="gw-msk-01" required />
          </div>
          <div className="field">
            <label>IP Address</label>
            <input value={ip} onChange={e => setIp(e.target.value)} placeholder="192.168.1.10" required />
          </div>
          <div className="field">
            <label>Location</label>
            <input value={location} onChange={e => setLocation(e.target.value)} placeholder="Moscow" required />
          </div>
          <label className="form-check">
            <span>Active</span>
            <input type="checkbox" checked={isActive} onChange={e => setIsActive(e.target.checked)} />
          </label>
          {err && <div className="form-error">{err}</div>}
          <div className="form-btns">
            <button type="button" className="btn btn-ghost" onClick={closeModal} disabled={saving}>Отмена</button>
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? '...' : deviceToEdit ? 'Сохранить' : 'Добавить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
