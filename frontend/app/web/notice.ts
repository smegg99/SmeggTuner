// app/web/notice.ts
// The first-open notice is remembered per browser; with storage blocked it shows on every visit.
const KEY = 'smeggtuner.web.noticeSeen'

export function noticeSeen(): boolean {
  try {
    return localStorage.getItem(KEY) === '1'
  }
  catch {
    return false
  }
}

export function markNoticeSeen(): void {
  try {
    localStorage.setItem(KEY, '1')
  }
  catch {
    // Shown again next visit.
  }
}
