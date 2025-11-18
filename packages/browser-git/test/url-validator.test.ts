import { describe, it, expect } from 'vitest';
import {
  validateGitURL,
  validateGitHubURL,
  validateGitLabURL,
  tryValidateGitURL,
  URLValidationError,
  GitHostingPresets,
} from '../src/utils/url-validator';

describe('URL Validator', () => {
  describe('validateGitURL', () => {
    it('should accept valid HTTPS GitHub URL', () => {
      const url = 'https://github.com/user/repo.git';
      const result = validateGitURL(url);
      expect(result).toBeInstanceOf(URL);
      expect(result.hostname).toBe('github.com');
    });

    it('should reject HTTP by default', () => {
      const url = 'http://github.com/user/repo.git';
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
      expect(() => validateGitURL(url)).toThrow('HTTP protocol is not allowed');
    });

    it('should allow HTTP when allowHttp is true', () => {
      const url = 'http://github.com/user/repo.git';
      const result = validateGitURL(url, { allowHttp: true });
      expect(result).toBeInstanceOf(URL);
    });

    it('should reject localhost by default', () => {
      const url = 'https://localhost:3000/repo.git';
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
      expect(() => validateGitURL(url)).toThrow('Localhost URLs are not allowed');
    });

    it('should allow localhost in development mode', () => {
      const url = 'http://localhost:3000/repo.git';
      const result = validateGitURL(url, GitHostingPresets.development());
      expect(result).toBeInstanceOf(URL);
    });

    it('should reject private IP addresses', () => {
      const privateIPs = [
        'https://10.0.0.1/repo.git',
        'https://172.16.0.1/repo.git',
        'https://192.168.1.1/repo.git',
        'https://127.0.0.1/repo.git',
        'https://169.254.1.1/repo.git',
      ];

      privateIPs.forEach(url => {
        expect(() => validateGitURL(url)).toThrow(URLValidationError);
        expect(() => validateGitURL(url)).toThrow(/Private IP|Localhost/);
      });
    });

    it('should allow private IPs in development mode', () => {
      const url = 'http://192.168.1.1/repo.git';
      const result = validateGitURL(url, GitHostingPresets.development());
      expect(result).toBeInstanceOf(URL);
    });

    it('should reject URLs with path traversal', () => {
      const url = 'https://github.com/user/../admin/repo.git';
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
      expect(() => validateGitURL(url)).toThrow('directory traversal');
    });

    it('should enforce allowed domains', () => {
      const url = 'https://evil.com/repo.git';
      const options = { allowedDomains: ['github.com', 'gitlab.com'] };
      expect(() => validateGitURL(url, options)).toThrow(URLValidationError);
      expect(() => validateGitURL(url, options)).toThrow('not in the allow list');
    });

    it('should accept URLs from allowed domains', () => {
      const url = 'https://github.com/user/repo.git';
      const options = { allowedDomains: ['github.com', 'gitlab.com'] };
      const result = validateGitURL(url, options);
      expect(result).toBeInstanceOf(URL);
    });

    it('should enforce denied domains', () => {
      const url = 'https://evil.com/repo.git';
      const options = { deniedDomains: ['evil.com', 'malicious.org'] };
      expect(() => validateGitURL(url, options)).toThrow(URLValidationError);
      expect(() => validateGitURL(url, options)).toThrow('in the deny list');
    });

    it('should reject too long URLs', () => {
      const longPath = 'a'.repeat(3000);
      const url = `https://github.com/${longPath}`;
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
      expect(() => validateGitURL(url)).toThrow('exceeds maximum length');
    });

    it('should reject invalid URL format', () => {
      const url = 'not a valid url';
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
      expect(() => validateGitURL(url)).toThrow('Invalid URL format');
    });

    it('should reject file:// protocol', () => {
      const url = 'file:///etc/passwd';
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
      expect(() => validateGitURL(url)).toThrow('not allowed');
    });

    it('should reject javascript: protocol', () => {
      const url = 'javascript:alert(1)';
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
    });

    it('should accept subdomains of allowed domains', () => {
      const url = 'https://api.github.com/repos/user/repo.git';
      const options = { allowedDomains: ['github.com'] };
      const result = validateGitURL(url, options);
      expect(result).toBeInstanceOf(URL);
    });
  });

  describe('GitHub preset', () => {
    it('should only accept GitHub URLs', () => {
      const validURL = 'https://github.com/user/repo.git';
      const result = validateGitHubURL(validURL);
      expect(result).toBeInstanceOf(URL);

      const invalidURL = 'https://gitlab.com/user/repo.git';
      expect(() => validateGitHubURL(invalidURL)).toThrow(URLValidationError);
    });

    it('should accept github.com subdomains', () => {
      const url = 'https://raw.github.com/user/repo/main/file.txt';
      const result = validateGitHubURL(url);
      expect(result).toBeInstanceOf(URL);
    });
  });

  describe('GitLab preset', () => {
    it('should only accept GitLab URLs', () => {
      const validURL = 'https://gitlab.com/user/repo.git';
      const result = validateGitLabURL(validURL);
      expect(result).toBeInstanceOf(URL);

      const invalidURL = 'https://github.com/user/repo.git';
      expect(() => validateGitLabURL(invalidURL)).toThrow(URLValidationError);
    });
  });

  describe('tryValidateGitURL', () => {
    it('should return URL for valid URLs', () => {
      const url = 'https://github.com/user/repo.git';
      const result = tryValidateGitURL(url);
      expect(result).toBeInstanceOf(URL);
    });

    it('should return null for invalid URLs', () => {
      const url = 'https://localhost/repo.git';
      const result = tryValidateGitURL(url);
      expect(result).toBeNull();
    });

    it('should not throw for invalid URLs', () => {
      const url = 'invalid url';
      expect(() => tryValidateGitURL(url)).not.toThrow();
      expect(tryValidateGitURL(url)).toBeNull();
    });
  });

  describe('Strict preset', () => {
    it('should enforce HTTPS only', () => {
      const url = 'http://github.com/user/repo.git';
      expect(() => validateGitURL(url, GitHostingPresets.strict())).toThrow();
    });

    it('should block private IPs', () => {
      const url = 'https://192.168.1.1/repo.git';
      expect(() => validateGitURL(url, GitHostingPresets.strict())).toThrow();
    });

    it('should block localhost', () => {
      const url = 'https://localhost/repo.git';
      expect(() => validateGitURL(url, GitHostingPresets.strict())).toThrow();
    });
  });

  describe('Public hosts preset', () => {
    it('should accept GitHub, GitLab, and Bitbucket', () => {
      const urls = [
        'https://github.com/user/repo.git',
        'https://gitlab.com/user/repo.git',
        'https://bitbucket.org/user/repo.git',
      ];

      urls.forEach(url => {
        const result = validateGitURL(url, GitHostingPresets.publicHosts());
        expect(result).toBeInstanceOf(URL);
      });
    });

    it('should reject other domains', () => {
      const url = 'https://example.com/repo.git';
      expect(() => validateGitURL(url, GitHostingPresets.publicHosts())).toThrow();
    });
  });

  describe('IPv6 addresses', () => {
    it('should reject IPv6 localhost', () => {
      const url = 'https://[::1]/repo.git';
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
    });

    it('should reject IPv6 link-local addresses', () => {
      const url = 'https://[fe80::1]/repo.git';
      expect(() => validateGitURL(url)).toThrow(URLValidationError);
    });

    it('should reject IPv6 unique local addresses', () => {
      const urls = [
        'https://[fc00::1]/repo.git',
        'https://[fd00::1]/repo.git',
      ];

      urls.forEach(url => {
        expect(() => validateGitURL(url)).toThrow(URLValidationError);
      });
    });
  });
});
