import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { MAX_BASS_REGISTERS, MAX_REGISTERS, MAX_SWITCH_NAME } from '~/types/session'
import { canAddBass, canAddRegister } from '~/utils/accordion'

/*
 * THE CAPS MUST AGREE WITH core/session, EXACTLY.
 *
 * A control that lets you do something the backend refuses is worse than one that is missing:
 * the click lands, nothing happens, and there is nothing on screen to say why. Every rule below
 * is enforced in Go as well; these keep the button's disabled state honest about it.
 */

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..')
const go = readFileSync(path.join(root, 'backend/core/session/session.go'), 'utf8')

function declared(name: string): number {
  const m = new RegExp(`${name}\\s*=\\s*(\\d+)`).exec(go)
  expect(m, `core/session no longer declares ${name}`).not.toBeNull()
  return Number(m![1])
}

describe('the caps', () => {
  it('are the ones the backend enforces', () => {
    expect(MAX_REGISTERS).toBe(declared('MaxRegisters'))
    expect(MAX_BASS_REGISTERS).toBe(declared('MaxBassRegisters'))
    expect(MAX_SWITCH_NAME).toBe(declared('MaxSwitchName'))
  })
})

const LM = ['L', 'M1'] as const
const MMM = ['M1', 'M2', 'M3'] as const

describe('adding a treble switch', () => {
  it('is allowed when the ranks are new', () => {
    expect(canAddRegister([{ name: 'LM', banks: [...LM] }], [...MMM])).toBe(true)
  })

  // A switch IS the ranks it sounds. The same ranks under another name is the same switch, and a
  // card has one column per rank with no way to show them apart.
  it('is refused when another switch already sounds those ranks', () => {
    expect(canAddRegister([{ name: 'LM', banks: [...LM] }], [...LM])).toBe(false)
  })

  it('does not care what order the ranks were picked in', () => {
    expect(canAddRegister([{ name: 'LM', banks: ['M1', 'L'] }], [...LM])).toBe(false)
  })

  it('is refused with nothing picked', () => {
    expect(canAddRegister([], [])).toBe(false)
  })

  it('is refused once the list is full', () => {
    const full = Array.from({ length: MAX_REGISTERS }, (_, i) => ({
      name: `r${i}`, banks: [i % 2 ? 'M1' : 'L'] as never,
    }))
    expect(canAddRegister(full, [...MMM])).toBe(false)
  })
})

describe('adding a bass switch', () => {
  it('is allowed with feet and a name of its own', () => {
    expect(canAddBass([{ name: 'Bass', feet: [32, 16] }], [16, 8], 'Soft')).toBe(true)
  })

  it('is refused when another switch already sounds those feet', () => {
    expect(canAddBass([{ name: 'Bass', feet: [32, 16] }], [16, 32], 'Other')).toBe(false)
  })

  it('is refused when the name is taken', () => {
    expect(canAddBass([{ name: 'Bass', feet: [32, 16] }], [8, 4], 'Bass')).toBe(false)
  })

  it('is refused without feet, or without a name', () => {
    expect(canAddBass([], [], 'Soft')).toBe(false)
    expect(canAddBass([], [16, 8], '   ')).toBe(false)
  })

  it('is refused once the machine is full', () => {
    const full = Array.from({ length: MAX_BASS_REGISTERS }, (_, i) => ({
      name: `b${i}`, feet: [2 ** (i + 1)],
    }))
    expect(canAddBass(full, [16, 8], 'One more')).toBe(false)
  })
})
