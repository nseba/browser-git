import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { success, error } from '../utils/output.js';

export const initCommand = new Command('init')
  .description('Initialize a new git repository')
  .argument('[path]', 'path to initialize repository', '.')
  .option('--bare', 'create a bare repository')
  .option('--initial-branch <name>', 'initial branch name', 'main')
  .option('--hash <algorithm>', 'hash algorithm (sha1 or sha256)', 'sha1')
  .option('--storage <backend>', 'storage backend (indexeddb, opfs, localstorage, memory)', 'indexeddb')
  .action(async (path: string, options) => {
    try {
      const repo = await Repository.init(path, {
        bare: options.bare,
        initialBranch: options.initialBranch,
        hashAlgorithm: options.hash,
        storage: options.storage,
      });

      success(`Initialized empty Git repository in ${path}/.git/`);
    } catch (err) {
      error(`Failed to initialize repository: ${(err as Error).message}`);
      process.exit(1);
    }
  });
