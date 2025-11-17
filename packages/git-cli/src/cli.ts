#!/usr/bin/env node

import { Command } from 'commander';
import { initCommand } from './commands/init.js';
import { addCommand } from './commands/add.js';
import { commitCommand } from './commands/commit.js';
import { statusCommand } from './commands/status.js';
import { logCommand } from './commands/log.js';
import { diffCommand } from './commands/diff.js';
import { branchCommand } from './commands/branch.js';
import { checkoutCommand } from './commands/checkout.js';
import { mergeCommand } from './commands/merge.js';
import { cloneCommand } from './commands/clone.js';
import { fetchCommand } from './commands/fetch.js';
import { pullCommand } from './commands/pull.js';
import { pushCommand } from './commands/push.js';

const program = new Command();

program
  .name('bgit')
  .description('BrowserGit - A full-featured Git implementation for browsers')
  .version('0.1.0');

// Register commands
program.addCommand(initCommand);
program.addCommand(addCommand);
program.addCommand(commitCommand);
program.addCommand(statusCommand);
program.addCommand(logCommand);
program.addCommand(diffCommand);
program.addCommand(branchCommand);
program.addCommand(checkoutCommand);
program.addCommand(mergeCommand);
program.addCommand(cloneCommand);
program.addCommand(fetchCommand);
program.addCommand(pullCommand);
program.addCommand(pushCommand);

program.parse(process.argv);
