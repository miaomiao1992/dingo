// Simple OG image generator using SVG and sharp
import sharp from 'sharp';
import { readFileSync, writeFileSync } from 'fs';

const svg = `
<svg width="1200" height="630" xmlns="http://www.w3.org/2000/svg">
  <rect width="1200" height="630" fill="#1a1a1a"/>

  <!-- Dingo logo emoji -->
  <text x="600" y="200" font-size="120" text-anchor="middle" fill="#00d4aa">🦤</text>

  <!-- Title -->
  <text x="600" y="320" font-size="72" font-weight="bold" text-anchor="middle" fill="#ffffff" font-family="Arial, sans-serif">Dingo</text>

  <!-- Tagline -->
  <text x="600" y="380" font-size="36" text-anchor="middle" fill="#cccccc" font-family="Arial, sans-serif">The Meta-Language for Go</text>

  <!-- Features -->
  <text x="600" y="450" font-size="24" text-anchor="middle" fill="#00d4aa" font-family="Arial, sans-serif">Result types • Pattern matching • Error propagation • Sum types</text>

  <!-- URL -->
  <text x="600" y="550" font-size="28" text-anchor="middle" fill="#888888" font-family="Arial, sans-serif">dingolang.com</text>
</svg>
`;

try {
  // Check if sharp is available
  const sharpAvailable = await import('sharp').then(() => true).catch(() => false);

  if (sharpAvailable) {
    // Use sharp to convert SVG to PNG
    await sharp(Buffer.from(svg))
      .png()
      .toFile('public/og-image.png');
    console.log('✅ OG image created at public/og-image.png');
  } else {
    // Fallback: just write the SVG
    writeFileSync('public/og-image.svg', svg);
    console.log('⚠️  sharp not available, created SVG instead at public/og-image.svg');
  }
} catch (error) {
  console.error('Error generating OG image:', error.message);
  process.exit(1);
}
