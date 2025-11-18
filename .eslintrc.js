module.exports = {
  root: true,
  env: {
    browser: true,
    es2020: true,
    node: true,
  },
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2020,
    sourceType: 'module',
    ecmaFeatures: {
      jsx: true,
    },
  },
  plugins: ['@typescript-eslint'],
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'prettier',
  ],
  rules: {
    // Allow unused vars if they start with underscore
    '@typescript-eslint/no-unused-vars': [
      'warn',
      {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
        caughtErrorsIgnorePattern: '^_',
      },
    ],
    // Allow explicit any in some cases
    '@typescript-eslint/no-explicit-any': 'warn',
    // Allow empty functions for stubs
    '@typescript-eslint/no-empty-function': 'warn',
    // Allow this aliasing in tests
    '@typescript-eslint/no-this-alias': 'warn',
    // Allow ts-ignore comments (warn instead of error)
    '@typescript-eslint/ban-ts-comment': 'warn',
    // Allow constant conditions (used in tests)
    'no-constant-condition': 'warn',
  },
  ignorePatterns: [
    'node_modules/',
    'dist/',
    'build/',
    'coverage/',
    '*.min.js',
    '*.config.js',
    'playwright-report/',
    'test-results/',
    '*.wasm',
    '*.go',
    'packages/git-core/wasm/',
  ],
};
