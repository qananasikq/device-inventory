type Device = {
  id: number
  hostname: string
  ip: string
}

async function fetchDevices(): Promise<Device[]> {
  const res = await fetch('/api/devices')

  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }

  // в идеале тут стоит добавить runtime-валидацию (zod или io-ts)
  return (await res.json()) as Device[]
}

// Убрал глобальные devices / isLoading  они создавали race condition
// при быстрых повторных вызовах
export async function loadAndFilterDevices(
  search: string,
  signal?: AbortSignal
): Promise<Device[]> {
  const data = await fetchDevices()

  // нормализуем регистр и пробелы чтобы поиск был case-insensitive
  const needle = search.trim().toLowerCase()

  const filtered = data.filter((d) =>
    d.hostname.toLowerCase().includes(needle)
  )

  return filtered
}

// debounce  задержка вызова, чтобы не долбить API на каждый символ
function debounce<T extends (...args: unknown[]) => void>(
  fn: T,
  delayMs: number
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout> | undefined
  return (...args: Parameters<T>) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delayMs)
  }
}

// Пример использования (упрощённо)
async function example() {
  const searchInput: HTMLInputElement | null =
    document.querySelector('#search')
  if (!searchInput) return

  let abortController: AbortController | null = null

  // debounce + AbortController для отмены предыдущего fetch
  const handleInput = debounce(async () => {
    abortController?.abort()
    abortController = new AbortController()

    try {
      const list = await loadAndFilterDevices(
        searchInput.value,
        abortController.signal
      )
      console.log('Devices:', list)
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        return
      }
      console.error('Failed to load devices:', err)
    }
  }, 300)

  searchInput.addEventListener('input', handleInput)
}

example()
