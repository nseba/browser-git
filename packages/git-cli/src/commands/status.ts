import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { error, section, info, dim } from '../utils/output.js';
import chalk from 'chalk';

export const statusCommand = new Command('status')
  .description('Show the working tree status')
  .option('-s, --short', 'show short format')
  .option('-b, --branch', 'show branch information')
  .action(async (options) => {
    try {
      const repo = await Repository.open(process.cwd());
      const status = await repo.status();

      if (options.short) {
        // Short format
        status.modified.forEach((file: string) => console.log(chalk.red(' M'), file));
        status.added.forEach((file: string) => console.log(chalk.green('A '), file));
        status.deleted.forEach((file: string) => console.log(chalk.red(' D'), file));
        status.untracked.forEach((file: string) => console.log(chalk.red('??'), file));
      } else {
        // Long format
        section(`On branch ${status.branch || 'HEAD'}`);
        console.log();

        if (status.added.length > 0 || status.modified.length > 0 || status.deleted.length > 0) {
          console.log('Changes to be committed:');
          console.log(dim('  (use "bgit restore --staged <file>..." to unstage)'));
          console.log();
          status.added.forEach((file: string) => console.log(chalk.green(`  new file:   ${file}`)));
          status.modified.forEach((file: string) => console.log(chalk.green(`  modified:   ${file}`)));
          status.deleted.forEach((file: string) => console.log(chalk.green(`  deleted:    ${file}`)));
          console.log();
        }

        if (status.untracked.length > 0) {
          console.log('Untracked files:');
          console.log(dim('  (use "bgit add <file>..." to include in what will be committed)'));
          console.log();
          status.untracked.forEach((file: string) => console.log(chalk.red(`  ${file}`)));
          console.log();
        }

        if (status.modified.length === 0 && status.added.length === 0 &&
            status.deleted.length === 0 && status.untracked.length === 0) {
          info('nothing to commit, working tree clean');
        }
      }
    } catch (err) {
      error(`Failed to get status: ${(err as Error).message}`);
      process.exit(1);
    }
  });
