#!/usr/bin/env node
/**
 * Synchronize package versions across the monorepo
 *
 * This script ensures all packages have the same version number as the root package.json
 */

const fs = require("fs");
const path = require("path");

const ROOT_PACKAGE = path.join(__dirname, "../package.json");
const PACKAGES_DIR = path.join(__dirname, "../packages");

function readJSON(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf-8"));
}

function writeJSON(filePath, data) {
  fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + "\n", "utf-8");
}

function main() {
  // Read root version
  const rootPackage = readJSON(ROOT_PACKAGE);
  const targetVersion = rootPackage.version;

  console.log(`📦 Syncing all packages to version ${targetVersion}`);

  // Get all package directories
  const packageDirs = fs
    .readdirSync(PACKAGES_DIR, { withFileTypes: true })
    .filter((dirent) => dirent.isDirectory())
    .map((dirent) => path.join(PACKAGES_DIR, dirent.name));

  let updatedCount = 0;

  // Update each package
  for (const packageDir of packageDirs) {
    const packageJsonPath = path.join(packageDir, "package.json");

    if (!fs.existsSync(packageJsonPath)) {
      console.warn(`⚠️  No package.json found in ${packageDir}`);
      continue;
    }

    const packageData = readJSON(packageJsonPath);
    const currentVersion = packageData.version;

    if (currentVersion !== targetVersion) {
      console.log(
        `  Updating ${packageData.name}: ${currentVersion} → ${targetVersion}`,
      );
      packageData.version = targetVersion;

      // Update internal dependencies to use workspace protocol
      if (packageData.dependencies) {
        for (const [depName, depVersion] of Object.entries(
          packageData.dependencies,
        )) {
          if (depName.startsWith("@browser-git/")) {
            packageData.dependencies[depName] = "workspace:*";
          }
        }
      }

      if (packageData.devDependencies) {
        for (const [depName, depVersion] of Object.entries(
          packageData.devDependencies,
        )) {
          if (depName.startsWith("@browser-git/")) {
            packageData.devDependencies[depName] = "workspace:*";
          }
        }
      }

      writeJSON(packageJsonPath, packageData);
      updatedCount++;
    } else {
      console.log(
        `  ✓ ${packageData.name} already at version ${targetVersion}`,
      );
    }
  }

  if (updatedCount > 0) {
    console.log(
      `\n✅ Updated ${updatedCount} package(s) to version ${targetVersion}`,
    );
  } else {
    console.log(`\n✅ All packages already at version ${targetVersion}`);
  }
}

try {
  main();
} catch (error) {
  console.error("❌ Error syncing versions:", error.message);
  process.exit(1);
}
