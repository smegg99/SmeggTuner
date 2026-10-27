// app/web/platform.ts
export type Platform = 'linux' | 'windows' | 'macos' | 'android' | 'ios' | 'other'

// detectPlatform checks handhelds first: an iPad asking for desktop sites reports Macintosh but has touch points,
// and Chrome on Android in desktop mode reports X11 Linux but says Android in userAgentData.
export function detectPlatform(ua: string, uaPlatform = '', touchPoints = 0): Platform {
  const s = `${ua} ${uaPlatform}`.toLowerCase()
  if (/android/.test(s)) return 'android'
  if (/iphone|ipad|ipod/.test(s)) return 'ios'
  if (/macintosh|mac os/.test(s)) return touchPoints > 1 ? 'ios' : 'macos'
  if (/windows/.test(s)) return 'windows'
  if (/linux|x11|cros/.test(s)) return 'linux'
  return 'other'
}

export function currentPlatform(): Platform {
  const data = (navigator as Navigator & { userAgentData?: { platform?: string } }).userAgentData
  return detectPlatform(navigator.userAgent, data?.platform ?? '', navigator.maxTouchPoints ?? 0)
}
