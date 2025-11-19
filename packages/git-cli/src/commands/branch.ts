import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { success, error } from '../utils/output.js';
import chalk from 'chalk';

export const branchCommand = new Command('branch')
  .description('List, create, or delete branches')
  .argument('[branch-name]', 'name of branch to create')
  .option('-d, --delete <branch>', 'delete a branch')
  .option('-D, --force-delete <branch>', 'force delete a branch')
  .option('-m, --move <old-name> <new-name>', 'rename a branch')
  .option('-a, --all', 'list all branches (local and remote)')
  .option('-r, --remote', 'list remote branches')
  .action(async (branchName: string | undefined, options) => {
    try {
      const repo = await Repository.open(process.cwd());

      // Delete branch
      if (options.delete || options.forceDelete) {
        const toDelete = options.delete || options.forceDelete;
        await repo.deleteBranch(toDelete, !!options.forceDelete);
        success(`Deleted branch ${toDelete}`);
        return;
      }

      // Rename branch
      if (options.move) {
        const [oldName, newName] = options.move.split(' ');
        await repo.renameBranch(oldName, newName);
        success(`Renamed branch ${oldName} to ${newName}`);
        return;
      }

      // Create branch
      if (branchName) {
        await repo.createBranch(branchName);
        success(`Created branch ${branchName}`);
        return;
      }

      // List branches
      const branches = await repo.listBranches();

      const currentBranch = await repo.getCurrentBranch();

      branches.forEach(branch => {
        if (branch === currentBranch) {
          console.log(chalk.green(`* ${branch}`));
        } else {
          console.log(`  ${branch}`);
        }
      });
    } catch (err) {
      error(`Failed to manage branches: ${(err as Error).message}`);
      process.exit(1);
    }
  });
