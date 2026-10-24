// app/utils/desktopViewport.ts
/** Fit the desktop's minimum window without changing its arrangement. */
export function desktopScale(width: number, height: number): number {
  if (width <= 0 || height <= 0) return 1
  return Math.floor(Math.min(1, width / 960, height / 640) * 1e6) / 1e6
}
