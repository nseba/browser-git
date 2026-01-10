import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebars: SidebarsConfig = {
  docsSidebar: [
    "intro",
    "getting-started",
    {
      type: "category",
      label: "Architecture",
      items: [
        "architecture/overview",
        "architecture/storage-layer",
        "architecture/wasm-bridge",
      ],
    },
    {
      type: "category",
      label: "Guides",
      items: [
        "guides/integration",
        "guides/cors-workarounds",
        "guides/authentication",
      ],
    },
    "browser-compatibility",
    "limitations",
    "migration",
  ],
  apiSidebar: [
    {
      type: "category",
      label: "API Reference",
      items: [
        "api/repository",
        "api/filesystem",
        "api/storage-adapters",
        "api/diff-engine",
      ],
    },
  ],
};

export default sidebars;
