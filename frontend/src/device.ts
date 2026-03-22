export interface Device {
    id: number;
    hostname: string;
    ip: string;
    location: string;
    is_active: boolean;
    created_at: string;
}

export type DevicePayload = Pick<Device, 'hostname' | 'ip' | 'location' | 'is_active'>;

// если VITE_MOCK_API=true, работаем без бэкенда
const useMock = import.meta.env.VITE_MOCK_API === 'true';

let mockModule: typeof import('./mockApi') | null = null;
if (useMock) {
    mockModule = await import('./mockApi');
}

// пытаемся достать ошибку из JSON ответа
async function request<T>(url: string, opts?: RequestInit): Promise<T> {
    const res = await fetch(url, opts);
    if (!res.ok) {
        let message = `HTTP ${res.status}`;
        try {
            const data = await res.json() as { error?: string };
            if (data?.error) message = data.error;
        } catch {
            // не удалось распарсить JSON
        }
        throw new Error(message);
    }
    if (res.status === 204) return undefined as T;
    return res.json() as Promise<T>;
}

export function fetchDevices(query: string): Promise<Device[]> {
    if (mockModule) return mockModule.fetchDevices(query);
    return request<Device[]>(`/devices${query}`);
}

export function createDevice(payload: DevicePayload): Promise<Device> {
    if (mockModule) return mockModule.createDevice(payload);
    return request<Device>('/devices', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    });
}

export function updateDevice(id: number, payload: DevicePayload): Promise<Device> {
    if (mockModule) return mockModule.updateDevice(id, payload);
    return request<Device>(`/devices/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    });
}

export function deleteDevice(id: number): Promise<void> {
    if (mockModule) return mockModule.deleteDevice(id);
    return request<void>(`/devices/${id}`, { method: 'DELETE' });
}
