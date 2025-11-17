import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { success, error, warning, progress } from '../utils/output.js';

export const pushCommand = new Command('push')
  .description('Update remote refs along with associated objects')
  .argument('[remote]', 'remote name', 'origin')
  .argument('[refspec]', 'refspec to push')
  .option('-f, --force', 'force push (may lose commits on remote)')
  .option('--all', 'push all branches')
  .option('--tags', 'push all tags')
  .option('--delete', 'delete remote branch')
  .option('--set-upstream', 'set upstream for current branch')
  .option('--username <username>', 'username for authentication')
  .option('--token <token>', 'token for authentication')
  .action(async (remote: string, refspec: string | undefined, options) => {
    try {
      const repo = await Repository.open(process.cwd());

      if (options.force) {
        warning('Force push can cause data loss on remote repository');
      }

      const pushOptions: any = {
        force: options.force,
        all: options.all,
        tags: options.tags,
        delete: options.delete,
        setUpstream: options.setUpstream,
      };

      if (refspec) {
        pushOptions.refspec = refspec;
      }

      if (options.username && options.token) {
        await repo.setAuth({
          type: 'basic',
          username: options.username,
          password: options.token,
        });
      }

      // Show progress
      let lastProgress = 0;
      pushOptions.onProgress = (current: number, total: number) => {
        progress(current, total, 'Pushing objects');
        lastProgress = current;
      };

      const result = await repo.push(remote, pushOptions);

      if (lastProgress > 0) console.log();

      if (result.rejected) {
        error('Push rejected. Use --force to force push (may cause data loss)');
        process.exit(1);
      } else if (options.delete) {
        success(`Deleted remote branch ${refspec}`);
      } else {
        success(`Pushed to ${remote}`);
      }
    } catch (err) {
      error(`Failed to push: ${(err as Error).message}`);
      process.exit(1);
    }
  });
