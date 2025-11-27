import { formatDateForInput } from './date-utils';

/**
 * Downloads content as a markdown file.
 * Creates a blob, triggers download, and cleans up resources.
 *
 * @throws Error if download fails (e.g., blob creation or DOM manipulation error)
 */
export function downloadMarkdown(content: string, filename: string): void {
  let url: string | null = null;
  let link: HTMLAnchorElement | null = null;

  try {
    const blob = new Blob([content], { type: 'text/markdown;charset=utf-8' });
    url = URL.createObjectURL(blob);
    link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
  } catch (error) {
    console.error('Failed to download markdown file:', error);
    throw error;
  } finally {
    // Cleanup: remove link from DOM and revoke object URL
    if (link && link.parentNode) {
      document.body.removeChild(link);
    }
    if (url) {
      URL.revokeObjectURL(url);
    }
  }
}

/**
 * Generates a filename for export based on the date range.
 * Format: journal-export-YYYY-MM-DD-to-YYYY-MM-DD.md
 */
export function generateExportFilename(startDate: Date, endDate: Date): string {
  return `journal-export-${formatDateForInput(startDate)}-to-${formatDateForInput(endDate)}.md`;
}
