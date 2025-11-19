import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { success, error } from '../utils/output.js';

export const checkoutCommand = new Command('checkout')
  .description('Switch branches or restore working tree files')
  .argument('<target>', 'branch or commit to checkout')
  .argument('[paths...]', 'limit checkout to specific paths')
  .option('-b, --create', 'create and checkout new branch')
  .option('-B, --force-create', 'create/reset and checkout branch')
  .option('-f, --force', 'force checkout (discard local changes)')
  .action(async (target: string, paths: string[], options) => {
    try {
      const repo = await Repository.open(process.cwd());

      if (options.create || options.forceCreate) {
        await repo.createBranch(target);
        await repo.checkout(target, { force: options.force });
        success(`Switched to a new branch '${target}'`);
      } else if (paths && paths.length > 0) {
        await repo.checkout(target, { paths, force: options.force });
        success(`Restored ${paths.length} file(s) from ${target}`);
      } else {
        await repo.checkout(target, { force: options.force });
        success(`Switched to branch '${target}'`);
      }
    } catch (err) {
      error(`Failed to checkout: ${(err as Error).message}`);
      process.exit(1);
    }
  });
