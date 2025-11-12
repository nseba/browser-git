#!/usr/bin/env node

/**
 * Bundle size analysis script
 * Analyzes the WASM and TypeScript bundle sizes
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.join(__dirname, '..');

// ANSI color codes
const colors = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  red: '\x1b[31m',
};

function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
}

function getFileSize(filePath) {
  try {
    const stats = fs.statSync(filePath);
    return stats.size;
  } catch (error) {
    return 0;
  }
}

function analyzeDirectory(dirPath, label) {
  console.log(`\n${colors.bold}${colors.blue}${label}${colors.reset}`);
  console.log('─'.repeat(60));

  if (!fs.existsSync(dirPath)) {
    console.log(`${colors.yellow}  Directory not found${colors.reset}`);
    return { totalSize: 0, files: 0 };
  }

  let totalSize = 0;
  let fileCount = 0;
  const files = [];

  function walkDir(dir) {
    const items = fs.readdirSync(dir);

    for (const item of items) {
      const fullPath = path.join(dir, item);
      const stats = fs.statSync(fullPath);

      if (stats.isDirectory()) {
        walkDir(fullPath);
      } else {
        const size = stats.size;
        totalSize += size;
        fileCount++;

        const ext = path.extname(item);
        const relativePath = path.relative(dirPath, fullPath);
        files.push({ path: relativePath, size, ext });
      }
    }
  }

  walkDir(dirPath);

  // Sort files by size (largest first)
  files.sort((a, b) => b.size - a.size);

  // Show top 10 largest files
  const topFiles = files.slice(0, 10);
  if (topFiles.length > 0) {
    console.log(`\n  ${colors.bold}Top Files:${colors.reset}`);
    topFiles.forEach((file, index) => {
      const sizeStr = formatBytes(file.size).padStart(10);
      console.log(`  ${(index + 1).toString().padStart(2)}. ${sizeStr}  ${file.path}`);
    });
  }

  // Show breakdown by file type
  const byType = {};
  files.forEach(file => {
    const ext = file.ext || 'no-ext';
    byType[ext] = (byType[ext] || 0) + file.size;
  });

  if (Object.keys(byType).length > 0) {
    console.log(`\n  ${colors.bold}By Type:${colors.reset}`);
    Object.entries(byType)
      .sort(([, a], [, b]) => b - a)
      .forEach(([ext, size]) => {
        const sizeStr = formatBytes(size).padStart(10);
        const percentage = ((size / totalSize) * 100).toFixed(1);
        console.log(`     ${sizeStr}  ${ext.padEnd(10)}  (${percentage}%)`);
      });
  }

  console.log(`\n  ${colors.bold}Total:${colors.reset} ${formatBytes(totalSize)} (${fileCount} files)`);

  return { totalSize, files: fileCount };
}

function main() {
  console.log(`${colors.bold}${colors.green}`);
  console.log('┌────────────────────────────────────────────────────────────┐');
  console.log('│           Browser-Git Bundle Size Analysis                │');
  console.log('└────────────────────────────────────────────────────────────┘');
  console.log(colors.reset);

  // Analyze WASM
  const wasmPath = path.join(rootDir, 'packages/browser-git/dist');
  const wasmStats = analyzeDirectory(wasmPath, '📦 WASM Bundle (packages/browser-git/dist)');

  // Analyze each TypeScript package
  const packages = [
    { name: 'storage-adapters', path: 'packages/storage-adapters/dist' },
    { name: 'browser-git', path: 'packages/browser-git/dist' },
    { name: 'diff-engine', path: 'packages/diff-engine/dist' },
    { name: 'git-cli', path: 'packages/git-cli/dist' },
  ];

  let totalPackageSize = 0;
  let totalPackageFiles = 0;

  packages.forEach(pkg => {
    const pkgPath = path.join(rootDir, pkg.path);
    const stats = analyzeDirectory(pkgPath, `📦 ${pkg.name}`);
    totalPackageSize += stats.totalSize;
    totalPackageFiles += stats.files;
  });

  // Summary
  console.log(`\n${colors.bold}${colors.green}Summary${colors.reset}`);
  console.log('─'.repeat(60));
  console.log(`  ${colors.bold}Total Size:${colors.reset}  ${formatBytes(totalPackageSize + wasmStats.totalSize)}`);
  console.log(`  ${colors.bold}Total Files:${colors.reset} ${totalPackageFiles + wasmStats.files}`);
  console.log();

  // Size warnings
  const maxRecommendedSize = 5 * 1024 * 1024; // 5MB
  if (totalPackageSize + wasmStats.totalSize > maxRecommendedSize) {
    console.log(`${colors.yellow}⚠  Warning: Total bundle size exceeds recommended 5MB${colors.reset}`);
  }

  // Check for specific files
  const wasmFile = path.join(rootDir, 'packages/browser-git/dist/git-core.wasm');
  const wasmSize = getFileSize(wasmFile);
  if (wasmSize > 2 * 1024 * 1024) {
    console.log(`${colors.yellow}⚠  Warning: WASM file is larger than 2MB (${formatBytes(wasmSize)})${colors.reset}`);
  }
}

main();
