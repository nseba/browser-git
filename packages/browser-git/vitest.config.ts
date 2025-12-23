import { defineConfig, mergeConfig } from "vitest/config";
import sharedConfig from "../../vitest.config.shared";
import path from "path";

export default mergeConfig(
  sharedConfig,
  defineConfig({
    resolve: {
      alias: {
        "@browser-git/storage-adapters": path.resolve(
          __dirname,
          "../storage-adapters/src/index.ts",
        ),
      },
    },
    test: {
      setupFiles: ["./test/setup.ts"],
    },
  }),
);
