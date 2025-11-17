// Content collections configuration

import { defineCollection, z } from 'astro:content';
import { goldenLoader } from './lib/goldenLoader';

// Define schema for golden examples
const goldenExamples = defineCollection({
  loader: goldenLoader(),
  schema: z.object({
    id: z.string(),
    title: z.string(),
    featureType: z.string(),
    featureDisplayName: z.string(),
    dingoCode: z.string(),
    goCode: z.string(),
    reasoning: z.string().optional(),
    order: z.number(),
  }),
});

// Export collections
export const collections = {
  'golden-examples': goldenExamples,
};
