import { computed } from 'vue'
import { useDisplay } from 'vuetify'

// The same boundary as the fixed phone/tablet layouts in instrument.scss.
export function useCompactLayout() {
  const { width, height } = useDisplay()
  return computed(() => width.value < 960 || height.value < 560)
}

/** Landscape keeps navigation in the toolbar to preserve the short workspace. */
export function useLandscapeLayout() {
  const { width, height } = useDisplay()
  return computed(() => width.value >= 560 && height.value < 560)
}

/** Lists follow the same logical viewport as the desktop shell. */
export function usePagedLayout() {
  const { width, height } = useDisplay()
  return computed(() => width.value < 960 || height.value < 560)
}

/** A landscape phone wide enough for the note keys in the top bar: the tuner then gives the plot its whole
 *  width. The same sizes as the media queries of TunerView and NotePanel. */
export function useWideLandscapeLayout() {
  const { width, height } = useDisplay()
  return computed(() => width.value >= 780 && height.value < 560)
}
