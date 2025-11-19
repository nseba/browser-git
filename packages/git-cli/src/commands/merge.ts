import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { success, error, warning } from '../utils/output.js';

export const mergeCommand = new Command('merge')
  .description('Join two or more development histories together')
  .argument('<branch>', 'branch to merge into current branch')
  .option('--no-ff', 'create merge commit even if fast-forward is possible')
  .option('--ff-only', 'refuse to merge unless fast-forward is possible')
  .option('-m, --message <message>', 'merge commit message')
  .option('--abort', 'abort the current merge')
  .action(async (branch: string, options) => {
    try {
      const repo = await Repository.open(process.cwd());

      if (options.abort) {
        await repo.mergeAbort();
        success('Merge aborted');
        return;
      }

      const result = await repo.merge(branch, {
        noFastForward: options.noFf,
        fastForwardOnly: options.ffOnly,
        message: options.message,
      });

      if (result.conflicts && result.conflicts.length > 0) {
        warning(`Automatic merge failed. Fix conflicts and commit the result.`);
        console.log('\nConflicts:');
        result.conflicts.forEach((conflict: any) => {
          console.log(`  ${conflict.path}`);
        });
        process.exit(1);
      } else if (result.fastForward) {
        success(`Fast-forward merge to ${result.commitHash?.substring(0, 7)}`);
      } else {
        success(`Merge completed with commit ${result.commitHash?.substring(0, 7)}`);
      }
    } catch (err) {
      error(`Failed to merge: ${(err as Error).message}`);
      process.exit(1);
    }
  });
