// Custom loader for golden test files
// Reads .dingo, .go.golden, and .reasoning.md files from ../tests/golden/

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import type { Loader } from 'astro/loaders';
import {
  extractFeatureType,
  extractOrder,
  extractBaseName,
  generateTitle,
  featureTypes,
} from './featureMapping';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export interface GoldenExample {
  id: string;
  title: string;
  featureType: string;
  featureDisplayName: string;
  dingoCode: string;
  goCode: string;
  reasoning?: string;
  order: number;
}

/**
 * Custom loader to read golden test files from ../tests/golden/
 */
export function goldenLoader(): Loader {
  return {
    name: 'golden-loader',
    load: async ({ store, logger }) => {
      // Resolve path to golden tests directory
      // From: /Users/jack/mag/dingo/langingpage/src/lib/
      // To:   /Users/jack/mag/dingo/tests/golden/
      const currentDir = path.dirname(fileURLToPath(import.meta.url));
      const goldenDir = path.resolve(currentDir, '../../../tests/golden');

      logger.info(`Loading golden examples from: ${goldenDir}`);

      // Check if directory exists
      if (!fs.existsSync(goldenDir)) {
        logger.warn(`Golden tests directory not found: ${goldenDir}`);
        return;
      }

      // Read all files from golden directory
      const allFiles = fs.readdirSync(goldenDir);

      // Group files by base name
      const fileGroups: Map<string, { dingo?: string; go?: string; reasoning?: string }> =
        new Map();

      for (const file of allFiles) {
        if (file.endsWith('.dingo')) {
          const baseName = extractBaseName(file);
          if (!fileGroups.has(baseName)) {
            fileGroups.set(baseName, {});
          }
          fileGroups.get(baseName)!.dingo = file;
        } else if (file.endsWith('.go.golden')) {
          const baseName = extractBaseName(file);
          if (!fileGroups.has(baseName)) {
            fileGroups.set(baseName, {});
          }
          fileGroups.get(baseName)!.go = file;
        } else if (file.endsWith('.reasoning.md')) {
          const baseName = extractBaseName(file);
          if (!fileGroups.has(baseName)) {
            fileGroups.set(baseName, {});
          }
          fileGroups.get(baseName)!.reasoning = file;
        }
      }

      logger.info(`Found ${fileGroups.size} example groups`);

      // Process each group and create collection entries
      for (const [baseName, files] of fileGroups) {
        try {
          // Skip if missing required files (.dingo or .go.golden)
          if (!files.dingo || !files.go) {
            logger.warn(`Skipping incomplete example: ${baseName} (missing .dingo or .go.golden)`);
            continue;
          }

          // Read file contents with error handling
          const dingoCode = fs.readFileSync(path.join(goldenDir, files.dingo), 'utf-8');
          const goCode = fs.readFileSync(path.join(goldenDir, files.go), 'utf-8');
          const reasoning = files.reasoning
            ? fs.readFileSync(path.join(goldenDir, files.reasoning), 'utf-8')
            : undefined;

          // Validate content is not empty
          if (!dingoCode.trim()) {
            logger.warn(`Empty Dingo code in: ${files.dingo} - skipping`);
            continue;
          }

          if (!goCode.trim()) {
            logger.warn(`Empty Go code in: ${files.go} - skipping`);
            continue;
          }

          // Validate content is valid text (not binary)
          if (dingoCode.includes('\0') || goCode.includes('\0')) {
            logger.error(
              `Binary content detected in: ${baseName} (expected text files) - skipping`,
            );
            continue;
          }

          // Extract metadata
          const featureType = extractFeatureType(baseName);
          const featureDisplayName = featureTypes[featureType] || 'Other';
          const order = extractOrder(baseName);
          const title = generateTitle(baseName);

          // Create entry
          const entry: GoldenExample = {
            id: baseName,
            title,
            featureType,
            featureDisplayName,
            dingoCode,
            goCode,
            reasoning,
            order,
          };

          // Store entry in collection
          store.set({
            id: baseName,
            data: entry,
          });

          logger.info(`Loaded example: ${baseName} (${featureDisplayName})`);
        } catch (error) {
          // Catch and log file read errors without crashing the build
          const errorMessage = error instanceof Error ? error.message : String(error);
          logger.error(`Failed to load example ${baseName}: ${errorMessage}`);
          continue;
        }
      }

      logger.info('Golden examples loaded successfully');
    },
  };
}
