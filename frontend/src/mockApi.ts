import type { Device, DevicePayload } from './device';

// мок для разработки без бэкенда
let nextId = 4;
let store: Device[] = [
    { id: 1, hostname: 'gw-msk-01', ip: '192.168.1.10', location: 'Moscow', is_active: true, created_at: '2025-12-01T10:00:00Z' },
    { id: 2, hostname: 'sw-spb-03', ip: '172.16.5.20', location: 'SPB', is_active: true, created_at: '2025-12-05T14:30:00Z' },
    { id: 3, hostname: 'fw-old-02', ip: '10.10.0.3', location: 'Kazan', is_active: false, created_at: '2025-11-20T09:00:00Z' },
];

function delay(ms = 150): Promise<void> {
    return new Promise(r => setTimeout(r, ms));
}

export async function fetchDevices(query: string): Promise<Device[]> {
    await delay();
    const params = new URLSearchParams(query.replace(/^\?/, ''));
    let result = [...store];

    const active = params.get('is_active');
    if (active === 'true') result = result.filter(d => d.is_active);
    if (active === 'false') result = result.filter(d => !d.is_active);

    const search = params.get('search');
    if (search) {
        const q = search.toLowerCase();
        result = result.filter(d => d.hostname.toLowerCase().includes(q));
    }

    return result.sort((a, b) => b.id - a.id);
}

export async function createDevice(payload: DevicePayload): Promise<Device> {
    await delay();
    const device: Device = {
        id: nextId++,
        ...payload,
        created_at: new Date().toISOString(),
    };
    store.push(device);
    return device;
}

export async function updateDevice(id: number, payload: DevicePayload): Promise<Device> {
    await delay();
    const idx = store.findIndex(d => d.id === id);
    if (idx === -1) throw new Error('not found');
    store[idx] = { ...store[idx], ...payload };
    return store[idx];
}

// soft delete как на бэкенде
export async function deleteDevice(id: number): Promise<void> {
    await delay();
    const idx = store.findIndex(d => d.id === id);
    if (idx === -1) throw new Error('not found');
    store[idx].is_active = false;
}
