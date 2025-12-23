# Security Policy

## Supported Versions

Currently, BrowserGit is in pre-release development. Security updates will be provided for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of BrowserGit seriously. If you discover a security vulnerability, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **[security@your-domain.com]** (replace with actual contact)

Include the following information:

- Type of vulnerability
- Full paths of affected source files
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within 48 hours
- **Updates**: We will send you regular updates about our progress
- **Fix Timeline**: We aim to release a fix within 30 days for critical vulnerabilities
- **Credit**: We will credit you in the security advisory (unless you prefer to remain anonymous)

## Security Considerations for Users

### Credential Storage

**WARNING**: BrowserGit stores authentication credentials in browser storage mechanisms. This has inherent security limitations:

- **LocalStorage/SessionStorage**: Credentials stored here are accessible to any JavaScript code running on the same origin
- **Not encrypted**: Passwords and tokens are stored in plaintext (base64 is NOT encryption)
- **XSS Risk**: Any XSS vulnerability in your application could expose these credentials

**Best Practices**:

1. **Use token-based authentication** instead of passwords
2. **Use short-lived tokens** with minimal permissions
3. **Never store credentials in localStorage** for production applications
4. **Use 'memory' storage** (default) which clears on page reload
5. **Implement Content Security Policy (CSP)** to mitigate XSS attacks

### CORS and Privacy

BrowserGit makes HTTP requests directly from the browser to Git servers. This has implications:

- **CORS Required**: The Git server must send appropriate CORS headers
- **Credentials Exposed**: Authentication tokens are sent in requests that could be intercepted
- **CORS Proxies**: Using CORS proxies means trusting a third party with your credentials

**Best Practices**:

1. Only connect to trusted Git servers
2. Use HTTPS exclusively (never HTTP)
3. Implement proper CORS configuration on your Git server
4. Avoid CORS proxies for sensitive repositories

### Content Security Policy (CSP)

BrowserGit uses WebAssembly and requires specific CSP directives:

```
Content-Security-Policy:
  script-src 'self' 'wasm-unsafe-eval';
  worker-src 'self' blob:;
```

**Note**: `'wasm-unsafe-eval'` is required for WASM execution and may not be acceptable in highly restrictive environments.

### Path Traversal Protection

BrowserGit normalizes paths to prevent directory traversal attacks. However:

- Paths are constrained within the virtual filesystem
- The `normalize()` function resolves `..` segments
- Applications should validate user-provided paths before passing to BrowserGit

**Best Practices**:

1. Validate and sanitize all user-provided file paths
2. Use absolute paths when possible
3. Implement application-level access controls

### Remote URL Validation

When cloning or fetching from remote repositories:

- **SSRF Risk**: BrowserGit does not currently validate remote URLs against internal networks
- **Trust**: Only clone from trusted sources
- **Validation**: URLs are parsed but not validated against allow/deny lists

**Best Practices**:

1. Validate repository URLs before passing to BrowserGit
2. Implement URL allow-lists in your application
3. Never pass user-provided URLs directly to clone/fetch operations without validation
4. Consider implementing network boundary checks

Example validation:

```typescript
function validateRepositoryURL(url: string): boolean {
  try {
    const parsed = new URL(url);

    // Only allow HTTPS
    if (parsed.protocol !== "https:") {
      return false;
    }

    // Deny internal/private IP ranges
    const hostname = parsed.hostname;
    if (
      hostname === "localhost" ||
      hostname === "127.0.0.1" ||
      hostname.startsWith("192.168.") ||
      hostname.startsWith("10.") ||
      hostname.startsWith("172.")
    ) {
      return false;
    }

    // Allow-list of trusted domains
    const allowedDomains = ["github.com", "gitlab.com", "bitbucket.org"];
    if (
      !allowedDomains.some(
        (domain) => hostname === domain || hostname.endsWith("." + domain),
      )
    ) {
      return false;
    }

    return true;
  } catch {
    return false;
  }
}
```

### XSS Protection in Example Applications

The example applications use `innerHTML` and `dangerouslySetInnerHTML` for rendering:

- **Risk**: These can introduce XSS vulnerabilities if not properly sanitized
- **Mitigation**: Content is either controlled or sanitized through libraries (e.g., marked)
- **Warning**: Do not copy example code to production without proper security review

**Best Practices**:

1. Sanitize all user-generated content before rendering
2. Use React's built-in XSS protection (avoid `dangerouslySetInnerHTML`)
3. Implement CSP headers
4. Use DOMPurify or similar sanitization libraries for markdown/HTML content

## Known Limitations

### Security Limitations

1. **Browser Storage**: All data is stored in browser storage, which is accessible to JavaScript
2. **No Encryption at Rest**: Git objects and credentials are not encrypted in storage
3. **CORS Dependency**: Requires CORS support from Git servers
4. **No GPG Signing**: Commit signing with GPG keys is not supported
5. **No SSH Authentication**: SSH keys cannot be used for authentication in browsers
6. **Token Exposure**: Tokens are sent in HTTP headers and could be logged/cached

### Not Suitable For

- **Highly sensitive repositories**: Use native Git with proper encryption
- **Compliance requirements**: May not meet SOC2, HIPAA, or similar standards
- **Production credential storage**: Use proper secrets management solutions
- **Unattended operations**: Browser tabs can be closed, storage can be cleared

## Security Features

### Implemented

- ✅ Path normalization to prevent directory traversal
- ✅ Input validation for authentication configuration
- ✅ CORS error detection and helpful error messages
- ✅ Authentication validation
- ✅ No eval() or Function() constructor usage
- ✅ Secure random number generation using Web Crypto API
- ✅ HTTPS enforcement recommendations

### Planned

- 🔄 URL validation and SSRF protection (v0.2.0)
- 🔄 Enhanced credential storage with Web Crypto API encryption (v0.3.0)
- 🔄 Optional GPG signing support (v0.4.0)
- 🔄 Security audit log (v0.5.0)

## Security Testing

BrowserGit undergoes regular security testing:

- Static code analysis
- Dependency vulnerability scanning
- Manual security review
- Fuzzing of parsers and protocol handlers

## Dependencies

BrowserGit has minimal dependencies to reduce supply chain risk. All dependencies are:

- Actively maintained
- Scanned for known vulnerabilities
- Pinned to specific versions
- Reviewed before updates

## Updates and Patches

Security patches will be released as soon as possible after discovery and verification. Subscribe to GitHub releases or watch the repository for notifications.

## Questions?

If you have questions about security that are not covered here, please open a public discussion in GitHub Discussions or contact the maintainers.

## Responsible Disclosure

We follow responsible disclosure practices:

1. Vulnerabilities are fixed before public disclosure
2. Security advisories are published after fixes are available
3. Credit is given to reporters (with permission)
4. CVE IDs are obtained for significant vulnerabilities

---

**Last Updated**: 2025-11-18
