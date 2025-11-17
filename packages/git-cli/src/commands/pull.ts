import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { success, error, warning, progress } from '../utils/output.js';

export const pullCommand = new Command('pull')
  .description('Fetch from and integrate with another repository or a local branch')
  .argument('[remote]', 'remote name', 'origin')
  .argument('[branch]', 'branch to pull')
  .option('--rebase', 'rebase instead of merge')
  .option('--ff-only', 'only update if fast-forward is possible')
  .option('--no-ff', 'create merge commit even if fast-forward is possible')
  .option('--username <username>', 'username for authentication')
  .option('--token <token>', 'token for authentication')
  .action(async (remote: string, branch: string | undefined, options) => {
    try {
      const repo = await Repository.open(process.cwd());

      const pullOptions: any = {
        rebase: options.rebase,
        fastForwardOnly: options.ffOnly,
        noFastForward: options.noFf,
      };

      if (branch) {
        pullOptions.branch = branch;
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
      pullOptions.onProgress = (current: number, total: number) => {
        progress(current, total, 'Pulling changes');
        lastProgress = current;
      };

      const result = await repo.pull(remote, pullOptions);

      if (lastProgress > 0) console.log();

      if (result.conflicts && result.conflicts.length > 0) {
        warning(`Automatic merge failed. Fix conflicts and commit the result.`);
        console.log('\nConflicts:');
        result.conflicts.forEach(conflict => {
          console.log(`  ${conflict.path}`);
        });
        process.exit(1);
      } else if (result.fastForward) {
        success(`Fast-forwarded to ${result.commitHash?.substring(0, 7)}`);
      } else if (result.upToDate) {
        success('Already up to date');
      } else {
        success(`Pull completed with commit ${result.commitHash?.substring(0, 7)}`);
      }
    } catch (err) {
      error(`Failed to pull: ${(err as Error).message}`);
      process.exit(1);
    }
  });
