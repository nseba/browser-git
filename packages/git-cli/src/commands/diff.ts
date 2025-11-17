import { Command } from 'commander';
import { Repository } from '@browser-git/browser-git';
import { error } from '../utils/output.js';
import chalk from 'chalk';

export const diffCommand = new Command('diff')
  .description('Show changes between commits, commit and working tree, etc')
  .argument('[paths...]', 'limit diff to specific paths')
  .option('--cached', 'show diff between index and HEAD')
  .option('--staged', 'same as --cached')
  .option('--stat', 'show diffstat instead of patch')
  .option('-U, --unified <lines>', 'number of context lines', '3')
  .action(async (paths: string[], options) => {
    try {
      const repo = await Repository.open(process.cwd());
      const diffs = await repo.diff({
        paths: paths.length > 0 ? paths : undefined,
        cached: options.cached || options.staged,
        unified: parseInt(options.unified),
      });

      if (diffs.length === 0) {
        console.log('No changes');
        return;
      }

      diffs.forEach(diff => {
        console.log(chalk.bold(`diff --git a/${diff.path} b/${diff.path}`));

        if (diff.oldMode !== diff.newMode) {
          console.log(`old mode ${diff.oldMode}`);
          console.log(`new mode ${diff.newMode}`);
        }

        if (options.stat) {
          const additions = diff.hunks.reduce((sum, h) => sum + h.additions, 0);
          const deletions = diff.hunks.reduce((sum, h) => sum + h.deletions, 0);
          console.log(` ${diff.path} | ${additions + deletions} ${chalk.green('+'.repeat(additions))}${chalk.red('-'.repeat(deletions))}`);
        } else {
          console.log(chalk.bold(`--- a/${diff.path}`));
          console.log(chalk.bold(`+++ b/${diff.path}`));

          diff.hunks.forEach(hunk => {
            console.log(chalk.cyan(`@@ -${hunk.oldStart},${hunk.oldLines} +${hunk.newStart},${hunk.newLines} @@`));

            hunk.lines.forEach(line => {
              if (line.type === 'add') {
                console.log(chalk.green(`+${line.content}`));
              } else if (line.type === 'delete') {
                console.log(chalk.red(`-${line.content}`));
              } else {
                console.log(` ${line.content}`);
              }
            });
          });
        }
        console.log();
      });
    } catch (err) {
      error(`Failed to show diff: ${(err as Error).message}`);
      process.exit(1);
    }
  });
