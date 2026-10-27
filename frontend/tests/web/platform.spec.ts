// tests/web/platform.spec.ts
import { describe, expect, it } from 'vitest'
import { detectPlatform } from '~/web/platform'

describe('detectPlatform', () => {
  it.each([
    ['android', 'Mozilla/5.0 (Linux; Android 14; V2352) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0 Mobile Safari/537.36', '', 5],
    ['ios', 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1', '', 5],
    ['ios', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15', '', 5],
    ['macos', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15', '', 0],
    ['windows', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0 Safari/537.36 Edg/154.0', 'Windows', 0],
    ['linux', 'Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0', '', 0],
    ['android', 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0 Safari/537.36', 'Android', 5],
  ])('detects %s', (want, ua, uaPlatform, touch) => {
    expect(detectPlatform(ua, uaPlatform as string, touch as number)).toBe(want)
  })
})
