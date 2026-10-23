import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/cn'
import { useFormat } from '@/lib/format'
import type { AudioBitrate, AudioFormat, MediaInfo } from '@/lib/types'
import { useAnalysis } from '@/store/analysis'
import { Segmented, Switch } from '@/components/ui/Controls'
import { Row } from '@/components/ui/Feedback'
import { Select } from '@/components/ui/Select'

const FORMATS: AudioFormat[] = ['mp3', 'm4a', 'opus', 'flac', 'wav']
const BITRATES: AudioBitrate[] = ['best', '320', '256', '192', '128', '96']

export function AudioTab({ info }: { info: MediaInfo }) {
  const { t } = useTranslation()
  const fmt = useFormat()
  const { sel, select } = useAnalysis()
  const lossless = sel.audioFormat === 'flac' || sel.audioFormat === 'wav'

  return (
    <div className="p-4">
      <div role="radiogroup" aria-label={t('result.audio.format')} className="grid grid-cols-5 gap-2">
        {FORMATS.map((f) => {
          const active = sel.audioFormat === f
          return (
            <button
              key={f}
              type="button"
              role="radio"
              aria-checked={active}
              onClick={() => select({ audioFormat: f, tab: 'audio', preset: f === 'mp3' || f === 'm4a' ? f : null })}
              className={cn(
                'relative rounded-xl border px-3 py-2.5 text-start transition-colors',
                active ? 'border-transparent' : 'hover:bg-line/[0.04]',
              )}
            >
              {active && <motion.span layoutId="audio-fmt" className="absolute inset-0 rounded-xl border border-accent/45 bg-accent/[0.08]" transition={{ type: 'spring', stiffness: 500, damping: 38 }} />}
              <span className="relative block text-[13px] font-semibold uppercase text-fg">{f}</span>
              <span className="relative mt-0.5 block text-2xs leading-snug text-muted">{t(`result.audio.formats.${f}`)}</span>
            </button>
          )
        })}
      </div>

      <div className="mt-5">
        <Row label={t('result.audio.bitrate')} hint={lossless ? t('result.audio.lossless') : undefined}>
          <Segmented<AudioBitrate>
            size="sm"
            label={t('result.audio.bitrate')}
            value={sel.audioBitrate}
            onChange={(audioBitrate) => select({ audioBitrate })}
            options={BITRATES.map((b) => ({ value: b, label: b === 'best' ? t('result.audio.best') : fmt.number(Number(b)), disabled: lossless }))}
          />
        </Row>
        {!info.isPlaylist && info.audio.length > 0 && (
          <Row label={t('result.audio.source')}>
            <Select
              size="sm"
              className="w-[260px]"
              label={t('result.audio.source')}
              value={sel.audioFormatId}
              onChange={(audioFormatId) => select({ audioFormatId })}
              options={[
                { value: '', label: t('result.audio.sourceAuto') },
                ...info.audio.map((a) => ({
                  value: a.id,
                  dir: 'ltr' as const,
                  label: `${a.acodec || a.ext.toUpperCase()} · ${fmt.bitrate(a.abr || a.tbr)}${a.language ? ` · ${a.language}` : ''}`,
                  hint: fmt.bytes(a.filesize, a.sizeApprox),
                })),
              ]}
            />
          </Row>
        )}
        {!info.matched && (
          <>
            <Row label={t('result.audio.embedThumbnail')}>
              <Switch label={t('result.audio.embedThumbnail')} checked={sel.embedThumbnail && sel.audioFormat !== 'wav'} disabled={sel.audioFormat === 'wav'} onChange={(embedThumbnail) => select({ embedThumbnail })} />
            </Row>
            <Row label={t('result.audio.embedMetadata')}>
              <Switch label={t('result.audio.embedMetadata')} checked={sel.embedMetadata} onChange={(embedMetadata) => select({ embedMetadata })} />
            </Row>
          </>
        )}
      </div>
    </div>
  )
}
