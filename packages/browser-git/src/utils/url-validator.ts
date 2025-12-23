/**
 * URL validation utilities to prevent SSRF and other URL-based attacks
 */

/**
 * URL validation error
 */
export class URLValidationError extends Error {
  constructor(
    message: string,
    public readonly url: string,
  ) {
    super(message);
    this.name = "URLValidationError";
  }
}

/**
 * URL validation options
 */
export interface URLValidationOptions {
  /**
   * Allowed protocols (default: ['https:', 'http:'])
   */
  allowedProtocols?: string[];

  /**
   * Whether to allow HTTP (default: false, HTTPS only)
   */
  allowHttp?: boolean;

  /**
   * Allowed domains (if specified, only these domains are allowed)
   */
  allowedDomains?: string[];

  /**
   * Denied domains (these domains are always blocked)
   */
  deniedDomains?: string[];

  /**
   * Whether to block private/internal IP addresses (default: true)
   */
  blockPrivateIPs?: boolean;

  /**
   * Whether to block localhost (default: true)
   */
  blockLocalhost?: boolean;

  /**
   * Maximum URL length (default: 2048)
   */
  maxLength?: number;
}

/**
 * Default validation options
 */
const defaultOptions: Required<URLValidationOptions> = {
  allowedProtocols: ["https:", "http:"],
  allowHttp: false,
  allowedDomains: [],
  deniedDomains: ["localhost", "127.0.0.1", "::1", "0.0.0.0"],
  blockPrivateIPs: true,
  blockLocalhost: true,
  maxLength: 2048,
};

/**
 * Checks if a hostname is a private IP address
 */
function isPrivateIP(hostname: string): boolean {
  // IPv4 private ranges
  const privateIPv4Patterns = [
    /^127\./, // 127.0.0.0/8 (loopback)
    /^10\./, // 10.0.0.0/8
    /^172\.(1[6-9]|2\d|3[01])\./, // 172.16.0.0/12
    /^192\.168\./, // 192.168.0.0/16
    /^169\.254\./, // 169.254.0.0/16 (link-local)
  ];

  // Check IPv4 patterns
  if (privateIPv4Patterns.some((pattern) => pattern.test(hostname))) {
    return true;
  }

  // IPv6 private addresses
  if (hostname.includes(":")) {
    // Loopback
    if (hostname === "::1" || hostname === "0:0:0:0:0:0:0:1") {
      return true;
    }
    // Link-local (fe80::/10)
    if (hostname.startsWith("fe80:") || hostname.startsWith("fe80::")) {
      return true;
    }
    // Unique local (fc00::/7) - starts with fc or fd followed by hex digit
    if (/^f[cd][0-9a-f]{0,2}:/i.test(hostname)) {
      return true;
    }
  }

  return false;
}

/**
 * Checks if a hostname is localhost
 */
function isLocalhost(hostname: string): boolean {
  const localhostPatterns = [
    "localhost",
    "127.0.0.1",
    "::1",
    "0.0.0.0",
    "0:0:0:0:0:0:0:0",
    "0:0:0:0:0:0:0:1",
  ];

  return localhostPatterns.includes(hostname.toLowerCase());
}

/**
 * Validates a Git repository URL
 *
 * @param url - The URL to validate
 * @param options - Validation options
 * @throws {URLValidationError} If the URL is invalid or unsafe
 * @returns The parsed URL object if valid
 */
export function validateGitURL(
  url: string,
  options: URLValidationOptions = {},
): URL {
  const opts = { ...defaultOptions, ...options };

  // Check URL length
  if (url.length > opts.maxLength) {
    throw new URLValidationError(
      `URL exceeds maximum length of ${opts.maxLength} characters`,
      url,
    );
  }

  // Check for directory traversal in the raw URL (before URL parsing normalizes it)
  // This catches patterns like /user/../admin which the URL parser would normalize
  if (url.includes("/..") || url.includes("\\..")) {
    throw new URLValidationError(
      "URL contains suspicious directory traversal pattern (..)",
      url,
    );
  }

  // Parse the URL
  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch (error) {
    throw new URLValidationError(
      `Invalid URL format: ${(error as Error).message}`,
      url,
    );
  }

  // Check protocol
  if (!opts.allowedProtocols.includes(parsed.protocol)) {
    throw new URLValidationError(
      `Protocol '${parsed.protocol}' is not allowed. Allowed protocols: ${opts.allowedProtocols.join(", ")}`,
      url,
    );
  }

  // Check HTTP vs HTTPS
  if (!opts.allowHttp && parsed.protocol === "http:") {
    throw new URLValidationError(
      "HTTP protocol is not allowed. Use HTTPS for secure connections.",
      url,
    );
  }

  // Get hostname and strip IPv6 brackets if present
  let hostname = parsed.hostname.toLowerCase();
  // IPv6 addresses in URLs are wrapped in brackets, e.g., [::1]
  // Strip them for validation
  if (hostname.startsWith("[") && hostname.endsWith("]")) {
    hostname = hostname.slice(1, -1);
  }

  // Check localhost
  if (opts.blockLocalhost && isLocalhost(hostname)) {
    throw new URLValidationError(
      "Localhost URLs are not allowed for security reasons",
      url,
    );
  }

  // Check private IPs
  if (opts.blockPrivateIPs && isPrivateIP(hostname)) {
    throw new URLValidationError(
      "Private IP addresses are not allowed for security reasons (SSRF protection)",
      url,
    );
  }

  // Check denied domains
  if (opts.deniedDomains.length > 0) {
    const isDenied = opts.deniedDomains.some(
      (denied) =>
        hostname === denied.toLowerCase() ||
        hostname.endsWith("." + denied.toLowerCase()),
    );
    if (isDenied) {
      throw new URLValidationError(
        `Domain '${hostname}' is in the deny list`,
        url,
      );
    }
  }

  // Check allowed domains (if specified)
  if (opts.allowedDomains.length > 0) {
    const isAllowed = opts.allowedDomains.some(
      (allowed) =>
        hostname === allowed.toLowerCase() ||
        hostname.endsWith("." + allowed.toLowerCase()),
    );
    if (!isAllowed) {
      throw new URLValidationError(
        `Domain '${hostname}' is not in the allow list`,
        url,
      );
    }
  }

  // Additional checks for git URLs
  const path = parsed.pathname;

  // Validate that URL looks like a git repository
  // Git URLs typically end with .git or have a path
  if (
    !path.endsWith(".git") &&
    !path.endsWith("/") &&
    path.split("/").length < 2
  ) {
    console.warn(`URL may not be a valid git repository: ${url}`);
  }

  return parsed;
}

/**
 * Preset validation options for common Git hosting services
 */
export const GitHostingPresets = {
  /**
   * GitHub only (github.com and subdomains)
   */
  github: (): URLValidationOptions => ({
    allowHttp: false,
    allowedDomains: ["github.com"],
    blockPrivateIPs: true,
    blockLocalhost: true,
  }),

  /**
   * GitLab only (gitlab.com and subdomains)
   */
  gitlab: (): URLValidationOptions => ({
    allowHttp: false,
    allowedDomains: ["gitlab.com"],
    blockPrivateIPs: true,
    blockLocalhost: true,
  }),

  /**
   * Bitbucket only (bitbucket.org and subdomains)
   */
  bitbucket: (): URLValidationOptions => ({
    allowHttp: false,
    allowedDomains: ["bitbucket.org"],
    blockPrivateIPs: true,
    blockLocalhost: true,
  }),

  /**
   * Common public Git hosts
   */
  publicHosts: (): URLValidationOptions => ({
    allowHttp: false,
    allowedDomains: ["github.com", "gitlab.com", "bitbucket.org"],
    blockPrivateIPs: true,
    blockLocalhost: true,
  }),

  /**
   * Strict security (HTTPS only, no private IPs, no localhost)
   */
  strict: (): URLValidationOptions => ({
    allowHttp: false,
    blockPrivateIPs: true,
    blockLocalhost: true,
  }),

  /**
   * Development mode (allows localhost and HTTP)
   */
  development: (): URLValidationOptions => ({
    allowHttp: true,
    blockPrivateIPs: false,
    blockLocalhost: false,
    deniedDomains: [],
  }),
};

/**
 * Convenience function to validate a GitHub repository URL
 */
export function validateGitHubURL(url: string): URL {
  return validateGitURL(url, GitHostingPresets.github());
}

/**
 * Convenience function to validate a GitLab repository URL
 */
export function validateGitLabURL(url: string): URL {
  return validateGitURL(url, GitHostingPresets.gitlab());
}

/**
 * Convenience function to validate a Bitbucket repository URL
 */
export function validateBitbucketURL(url: string): URL {
  return validateGitURL(url, GitHostingPresets.bitbucket());
}

/**
 * Validates and normalizes a Git URL
 * Returns null if invalid, URL object if valid
 */
export function tryValidateGitURL(
  url: string,
  options?: URLValidationOptions,
): URL | null {
  try {
    return validateGitURL(url, options);
  } catch {
    return null;
  }
}
