#!/usr/bin/env node

/**
 * Build script for Dingo TextMate grammar
 * Converts YAML to JSON for VS Code compatibility
 */

const fs = require('fs');
const path = require('path');
const yaml = require('js-yaml');

const YAML_PATH = path.join(__dirname, '../syntaxes/dingo.tmLanguage.yaml');
const JSON_PATH = path.join(__dirname, '../syntaxes/dingo.tmLanguage.json');

try {
  console.log('Building Dingo grammar...');

  // Read YAML file
  const yamlContent = fs.readFileSync(YAML_PATH, 'utf8');

  // Parse YAML
  const grammar = yaml.load(yamlContent);

  // Add schema for better editor support
  grammar.$schema = 'https://raw.githubusercontent.com/martinring/tmlanguage/master/tmlanguage.json';

  // Convert to JSON with pretty printing
  const jsonContent = JSON.stringify(grammar, null, 2);

  // Write JSON file
  fs.writeFileSync(JSON_PATH, jsonContent);

  console.log('✓ Grammar built successfully!');
  console.log(`  Input:  ${YAML_PATH}`);
  console.log(`  Output: ${JSON_PATH}`);

} catch (error) {
  console.error('✗ Build failed:', error.message);
  process.exit(1);
}
