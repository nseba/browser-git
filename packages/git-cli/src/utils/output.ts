import chalk from 'chalk';

export interface OutputOptions {
  color?: boolean;
}

const defaultOptions: OutputOptions = {
  color: true,
};

let options = { ...defaultOptions };

export function setOutputOptions(newOptions: Partial<OutputOptions>) {
  options = { ...options, ...newOptions };
}

export function success(message: string): void {
  if (options.color) {
    console.log(chalk.green('✓'), message);
  } else {
    console.log('✓', message);
  }
}

export function error(message: string): void {
  if (options.color) {
    console.error(chalk.red('✗'), message);
  } else {
    console.error('✗', message);
  }
}

export function warning(message: string): void {
  if (options.color) {
    console.warn(chalk.yellow('⚠'), message);
  } else {
    console.warn('⚠', message);
  }
}

export function info(message: string): void {
  if (options.color) {
    console.log(chalk.blue('ℹ'), message);
  } else {
    console.log('ℹ', message);
  }
}

export function header(message: string): void {
  if (options.color) {
    console.log(chalk.bold.cyan(message));
  } else {
    console.log(message);
  }
}

export function section(title: string): void {
  if (options.color) {
    console.log(chalk.bold.underline(title));
  } else {
    console.log(title);
    console.log('='.repeat(title.length));
  }
}

export function dim(message: string): void {
  if (options.color) {
    console.log(chalk.dim(message));
  } else {
    console.log(message);
  }
}

export function highlight(message: string): void {
  if (options.color) {
    console.log(chalk.bold(message));
  } else {
    console.log(message);
  }
}

export function table(rows: string[][]): void {
  if (rows.length === 0) return;

  const columnWidths = rows[0]!.map((_, colIndex) =>
    Math.max(...rows.map(row => row[colIndex]?.length || 0))
  );

  rows.forEach(row => {
    const formattedRow = row
      .map((cell, index) => cell.padEnd(columnWidths[index] || 0))
      .join('  ');
    console.log(formattedRow);
  });
}

export function list(items: string[], prefix: string = '•'): void {
  items.forEach(item => {
    console.log(`${prefix} ${item}`);
  });
}

export function progress(current: number, total: number, message?: string): void {
  const percentage = Math.round((current / total) * 100);
  const bar = '█'.repeat(Math.round(percentage / 2)) + '░'.repeat(50 - Math.round(percentage / 2));
  const status = message ? ` ${message}` : '';

  if (options.color) {
    process.stdout.write(`\r${chalk.cyan(bar)} ${percentage}%${status}`);
  } else {
    process.stdout.write(`\r${bar} ${percentage}%${status}`);
  }

  if (current === total) {
    process.stdout.write('\n');
  }
}
