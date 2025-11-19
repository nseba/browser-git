/**
 * Parse glob patterns from command line arguments
 */
export function parseGlobPatterns(args: string[]): string[] {
  return args.map(arg => {
    // Handle escaped characters
    return arg.replace(/\\\*/g, '*');
  });
}

/**
 * Parse key=value pairs from command line
 */
export function parseKeyValuePairs(args: string[]): Record<string, string> {
  const result: Record<string, string> = {};

  args.forEach(arg => {
    const match = arg.match(/^([^=]+)=(.+)$/);
    if (match && match[1] && match[2]) {
      result[match[1]] = match[2];
    }
  });

  return result;
}

/**
 * Parse author/committer format: "Name <email>"
 */
export function parseAuthor(authorString: string): { name: string; email: string } | null {
  const match = authorString.match(/^([^<]+)\s*<([^>]+)>$/);
  if (!match || !match[1] || !match[2]) {
    return null;
  }

  return {
    name: match[1].trim(),
    email: match[2].trim(),
  };
}

/**
 * Format author for display
 */
export function formatAuthor(author: { name: string; email: string }): string {
  return `${author.name} <${author.email}>`;
}

/**
 * Parse date string to Date object
 */
export function parseDate(dateString: string): Date {
  return new Date(dateString);
}

/**
 * Format date for display
 */
export function formatDate(date: Date): string {
  return date.toLocaleString();
}

/**
 * Format relative date (e.g., "2 hours ago")
 */
export function formatRelativeDate(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);
  const diffWeeks = Math.floor(diffDays / 7);
  const diffMonths = Math.floor(diffDays / 30);
  const diffYears = Math.floor(diffDays / 365);

  if (diffSecs < 60) return `${diffSecs} second${diffSecs !== 1 ? 's' : ''} ago`;
  if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
  if (diffWeeks < 4) return `${diffWeeks} week${diffWeeks !== 1 ? 's' : ''} ago`;
  if (diffMonths < 12) return `${diffMonths} month${diffMonths !== 1 ? 's' : ''} ago`;
  return `${diffYears} year${diffYears !== 1 ? 's' : ''} ago`;
}

/**
 * Truncate hash to short form (7 characters)
 */
export function shortHash(hash: string): string {
  return hash.substring(0, 7);
}

/**
 * Truncate string to max length
 */
export function truncate(str: string, maxLength: number): string {
  if (str.length <= maxLength) return str;
  return str.substring(0, maxLength - 3) + '...';
}
