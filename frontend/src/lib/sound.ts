// A soft two-note chime generated with WebAudio (no audio asset needed).
let ctx: AudioContext | null = null

export function playChime() {
  try {
    ctx ??= new AudioContext()
    const now = ctx.currentTime
    for (const [i, freq] of [880, 1318.5].entries()) {
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()
      osc.type = 'sine'
      osc.frequency.value = freq
      const t = now + i * 0.11
      gain.gain.setValueAtTime(0, t)
      gain.gain.linearRampToValueAtTime(0.12, t + 0.015)
      gain.gain.exponentialRampToValueAtTime(0.0001, t + 0.45)
      osc.connect(gain).connect(ctx.destination)
      osc.start(t)
      osc.stop(t + 0.5)
    }
  } catch {
    /* audio unavailable */
  }
}
