/** How the card is exported. `pdf` isn't a third renderer: it's the HTML sheet rendered to A4 by a headless browser on the backend. */
export type ReportFormat = 'html' | 'pdf' | 'csv'

/** What the report dialog emits, the genuinely per-export fields. The letterhead is not here: it's read from Settings when the sheet is written, so it can't go stale. */
export interface ReportOptions {
  format: ReportFormat
  /** ISO yyyy-mm-dd. */
  date: string
}
