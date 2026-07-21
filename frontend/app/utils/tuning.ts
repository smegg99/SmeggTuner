// CENTS_ACCEPTABLE: amber past the technician's tolerance but within ten cents. "In tune" is deliberately not here — that's the backend's inTol, and a second threshold could disagree.
export const CENTS_ACCEPTABLE = 10

// liveReedsUsable: whether a live measurement's per-reed numbers may be read as reeds. Same rule as record.ts reedsUsable, off the engine's flags; a merged pair not recovered from the beat is lobes, not reeds.
export function liveReedsUsable(separated: boolean, fromBeat: boolean): boolean {
  return separated || fromBeat
}

/** Whether the reeds were recovered from the measured beat rather than seen apart. */
export function liveReedsDerived(separated: boolean, fromBeat: boolean): boolean {
  return !separated && fromBeat
}

const PITCH_CLASSES = 12

// core/tuning/notes.go names detected notes; unplayed pitch classes (a transposition target) aren't named there, so these tables mirror it — keep in step.
const SHARP_NAMES = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'A#', 'B']
const GERMAN_NAMES = ['C', 'C#', 'D', 'D#', 'E', 'F', 'F#', 'G', 'G#', 'A', 'B', 'H']
const DO_RE_MI = ['Do', 'Do#', 'Re', 'Re#', 'Mi', 'Fa', 'Fa#', 'Sol', 'Sol#', 'La', 'La#', 'Si']

// Polish spelling — the one this app's technician reads. Must cover every common/config/config.cue scale_naming value; when missing it silently fell through to sharps (Gis4 vs G#4 in one window).
const POLISH_NAMES = ['C', 'Cis', 'D', 'Dis', 'E', 'F', 'Fis', 'G', 'Gis', 'A', 'Ais', 'H']

function scaleNames(naming: string | undefined): readonly string[] {
  switch (naming) {
    case 'cdefgah': return GERMAN_NAMES
    case 'doremi': return DO_RE_MI
    case 'polish': return POLISH_NAMES
    default: return SHARP_NAMES
  }
}

// pitchClassName names a pitch class, counting semitones from C; wraps both ways, so -1 is the B below rather than nothing.
export function pitchClassName(semitones: number, naming: string | undefined): string {
  const index = ((Math.trunc(semitones) % PITCH_CLASSES) + PITCH_CLASSES) % PITCH_CLASSES
  return scaleNames(naming)[index] ?? ''
}

// noteName names a MIDI note as pitch class plus octave, C4 = middle C.
export function noteName(note: number, naming?: string): string {
  const midi = Math.trunc(note)
  // MIDI 24 is C1: octave is note/12 less the octave MIDI counts below C0.
  const octave = Math.floor(midi / PITCH_CLASSES) - 1
  return `${pitchClassName(midi, naming)}${octave}`
}
