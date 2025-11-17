import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { success, error, progress } from '../utils/output.js';

export const cloneCommand = new Command('clone')
  .description('Clone a repository into a new directory')
  .argument('<url>', 'repository URL to clone')
  .argument('[directory]', 'directory to clone into')
  .option('--depth <depth>', 'create shallow clone with specified depth')
  .option('--branch <branch>', 'checkout specific branch instead of HEAD')
  .option('--bare', 'create bare repository')
  .option('--storage <backend>', 'storage backend (indexeddb, opfs, localstorage, memory)', 'indexeddb')
  .option('--username <username>', 'username for authentication')
  .option('--token <token>', 'token for authentication')
  .action(async (url: string, directory: string | undefined, options) => {
    try {
      const targetDir = directory || url.split('/').pop()?.replace('.git', '') || 'repository';

      const cloneOptions: any = {
        depth: options.depth ? parseInt(options.depth) : undefined,
        branch: options.branch,
        bare: options.bare,
        storage: options.storage,
      };

      if (options.username && options.token) {
        cloneOptions.auth = {
          type: 'basic',
          username: options.username,
          password: options.token,
        };
      }

      // Show progress
      let lastProgress = 0;
      cloneOptions.onProgress = (current: number, total: number) => {
        progress(current, total, 'Cloning repository');
        lastProgress = current;
      };

      const repo = await Repository.clone(url, targetDir, cloneOptions);

      if (lastProgress > 0) {
        console.log(); // New line after progress
      }

      success(`Cloned repository to ${targetDir}`);
    } catch (err) {
      error(`Failed to clone: ${(err as Error).message}`);
      process.exit(1);
    }
  });
