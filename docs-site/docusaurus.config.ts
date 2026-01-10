import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";

const config: Config = {
  title: "BrowserGit",
  tagline:
    "Full-featured Git implementation for browsers using Go + WebAssembly and TypeScript",
  favicon: "img/favicon.ico",

  future: {
    v4: true,
  },

  url: "https://nseba.github.io",
  baseUrl: "/browser-git/",

  organizationName: "nseba",
  projectName: "browser-git",

  onBrokenLinks: "throw",

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      {
        docs: {
          sidebarPath: "./sidebars.ts",
          editUrl: "https://github.com/nseba/browser-git/tree/main/docs-site/",
        },
        blog: false,
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: "img/browser-git-social-card.png",
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: "BrowserGit",
      logo: {
        alt: "BrowserGit Logo",
        src: "img/logo.svg",
      },
      items: [
        {
          type: "docSidebar",
          sidebarId: "docsSidebar",
          position: "left",
          label: "Documentation",
        },
        {
          type: "docSidebar",
          sidebarId: "apiSidebar",
          position: "left",
          label: "API Reference",
        },
        {
          href: "https://nseba.github.io/browser-git/examples/",
          label: "Examples",
          position: "left",
        },
        {
          href: "https://github.com/nseba/browser-git",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Documentation",
          items: [
            {
              label: "Getting Started",
              to: "/docs/getting-started",
            },
            {
              label: "API Reference",
              to: "/docs/api/repository",
            },
            {
              label: "Architecture",
              to: "/docs/architecture/overview",
            },
          ],
        },
        {
          title: "Guides",
          items: [
            {
              label: "Integration Guide",
              to: "/docs/guides/integration",
            },
            {
              label: "CORS Workarounds",
              to: "/docs/guides/cors-workarounds",
            },
            {
              label: "Authentication",
              to: "/docs/guides/authentication",
            },
          ],
        },
        {
          title: "Community",
          items: [
            {
              label: "GitHub Issues",
              href: "https://github.com/nseba/browser-git/issues",
            },
            {
              label: "Discussions",
              href: "https://github.com/nseba/browser-git/discussions",
            },
          ],
        },
        {
          title: "More",
          items: [
            {
              label: "Examples",
              href: "https://nseba.github.io/browser-git/examples/",
            },
            {
              label: "GitHub",
              href: "https://github.com/nseba/browser-git",
            },
            {
              label: "npm",
              href: "https://www.npmjs.com/package/@browser-git/browser-git",
            },
          ],
        },
      ],
      copyright: `Copyright ${new Date().getFullYear()} Sebastian Negomireanu. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ["bash", "typescript", "go"],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
