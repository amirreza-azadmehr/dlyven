import { Fragment } from 'react'

/**
 * Renders "a · b · c" where each segment is a bidi-isolated run that keeps the UI direction.
 * (Not <bdi>: its dir=auto would flip a translated RTL phrase that starts with "393 MB".)
 * Latin values inside segments are isolated by the formatter (FSI…PDI), so sizes, dates and
 * translated words never reorder each other in RTL UI.
 */
export function BidiList({ items, separator = ' · ' }: { items: string[]; separator?: string }) {
  const parts = items.filter(Boolean)
  return (
    <>
      {parts.map((p, i) => (
        <Fragment key={i}>
          {i > 0 && separator}
          <span className="[unicode-bidi:isolate]">{p}</span>
        </Fragment>
      ))}
    </>
  )
}
