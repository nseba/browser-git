import { defineConfig, mergeConfig } from 'vitest/config';
import sharedConfig from '../../vitest.config.shared';

export default mergeConfig(
  sharedConfig,
  defineConfig({
    test: {
      passWithNoTests: true,
    },
  })
);
