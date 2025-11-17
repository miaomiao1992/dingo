# UI Debugging Strategy Document

## Context

**Reference App**: React app at http://localhost:5174/ (in `reference/` folder)
- Uses inline Tailwind classes
- Working and looks correct
- Example-driven UI (code comparison examples)

**Target App**: Astro app at http://localhost:4321/ (current folder)
- Uses extracted components and global styling
- Looks different from reference (broken)
- Same content but different visual implementation

**Goal**: Identify and fix visual differences between the apps

---

## Debugging Approaches

### Approach 1: Visual Screenshot Comparison (Pixel-Level)
**Strategy**: Take full-page and section screenshots of both apps and compare them visually

**Steps**:
1. Take full-page screenshots of both apps at same viewport size
2. Take section-specific screenshots (header, sidebar, main content, footer)
3. Use visual diff tools to identify pixel-level differences
4. Document differences with annotations

**Tools**:
- Chrome DevTools screenshot API
- Image comparison (manual or programmatic)

**Pros**:
- Quick visual overview of what's different
- Easy to communicate issues to developers
- Can catch subtle visual bugs

**Cons**:
- Doesn't tell you WHY things are different
- Hard to automate fixes
- Content differences can create false positives

**Expected Output**:
- Screenshots of both apps
- Annotated images showing differences
- List of visual discrepancies

---

### Approach 2: DOM Structure Comparison (Element-Level)
**Strategy**: Extract and compare the DOM structure of both apps

**Steps**:
1. Take DOM snapshot of both apps
2. Compare element hierarchies
3. Identify structural differences (missing elements, wrong nesting, etc.)
4. Compare element attributes (classes, IDs, data attributes)

**Tools**:
- Chrome DevTools `take_snapshot` tool
- Custom DOM traversal scripts

**Pros**:
- Reveals structural issues
- Shows missing/extra elements
- Can identify class name differences

**Cons**:
- Large DOM trees are hard to compare manually
- Doesn't show computed styles (just declared classes)
- Can miss CSS-only issues

**Expected Output**:
- DOM snapshots of both apps
- List of structural differences
- Missing/extra elements report

---

### Approach 3: Computed CSS Comparison (Style-Level)
**Strategy**: Compare computed styles for corresponding elements

**Steps**:
1. Identify key elements in both apps (assign matching IDs/classes)
2. Extract computed CSS for each element
3. Compare CSS properties side-by-side
4. Identify which styles differ and by how much

**Tools**:
- Chrome DevTools `evaluate_script` with `getComputedStyle()`
- Custom comparison scripts

**Pros**:
- Shows exact CSS values being applied
- Reveals cascade/specificity issues
- Can identify conflicting styles

**Cons**:
- Requires element mapping (which element corresponds to which)
- Lots of noise (many CSS properties)
- Doesn't explain source of styles (inline, class, global)

**Expected Output**:
- Computed style reports for key elements
- Side-by-side CSS comparison
- List of differing properties with values

---

### Approach 4: Layout Box Model Analysis (Spacing-Level)
**Strategy**: Compare layout measurements (dimensions, spacing, positioning)

**Steps**:
1. Identify key layout elements (containers, flex/grid parents, etc.)
2. Extract box model data: width, height, margin, padding, border
3. Compare measurements between apps
4. Identify spacing/sizing discrepancies

**Tools**:
- Chrome DevTools `getBoundingClientRect()`
- Custom measurement scripts

**Pros**:
- Focuses on layout issues (common source of visual differences)
- Easy to quantify differences (e.g., "10px extra margin")
- Helps identify responsive breakpoint issues

**Cons**:
- Doesn't explain WHY sizes differ
- Can miss color/font/border style issues
- Requires element mapping

**Expected Output**:
- Box model measurements for both apps
- Spacing comparison report
- List of sizing/spacing differences

---

### Approach 5: Tailwind Class Audit (Framework-Level)
**Strategy**: Compare Tailwind classes used in both apps

**Steps**:
1. Extract all Tailwind classes from reference app (inline in JSX)
2. Extract all Tailwind classes from Astro app (in components/global CSS)
3. Create class usage comparison
4. Identify missing or different Tailwind utilities

**Tools**:
- DOM traversal to collect classes
- Class parsing and comparison

**Pros**:
- Specific to Tailwind (our styling framework)
- Shows exactly which utilities are missing
- Easy to apply fixes (add missing classes)

**Cons**:
- Assumes both apps use Tailwind correctly
- Doesn't account for global CSS overrides
- Can miss CSS cascade issues

**Expected Output**:
- List of all Tailwind classes in each app
- Missing classes report
- Different class usage report

---

### Approach 6: Incremental Component Isolation (Section-Level)
**Strategy**: Break page into sections and compare each independently

**Steps**:
1. Identify logical sections (header, sidebar, main content, footer, etc.)
2. For each section:
   - Compare DOM structure
   - Compare computed styles
   - Take screenshots
   - Document differences
3. Prioritize fixing by section

**Tools**:
- Chrome DevTools element selection
- Custom section identification scripts

**Pros**:
- Breaks large problem into smaller pieces
- Helps prioritize fixes (fix header first, then sidebar, etc.)
- Easier to identify which component is broken

**Cons**:
- More time-consuming (analyze each section)
- Can miss cross-section interactions (e.g., global styles affecting multiple sections)

**Expected Output**:
- Per-section analysis reports
- Prioritized list of broken components
- Component-specific fix recommendations

---

### Approach 7: CSS Specificity & Cascade Analysis (Source-Level)
**Strategy**: Analyze CSS rule sources and specificity to understand style conflicts

**Steps**:
1. For key elements, identify all CSS rules that apply
2. Check specificity of each rule
3. Identify which rules are being overridden
4. Find global styles that might conflict with Tailwind utilities

**Tools**:
- Chrome DevTools "Computed" tab → "Show All" (see all rules)
- Custom specificity calculator

**Pros**:
- Reveals CSS cascade issues (common in Tailwind + global CSS)
- Shows which styles are being overridden
- Helps understand "why" styles don't match

**Cons**:
- Complex to automate
- Requires CSS expertise to interpret
- Can be tedious for many elements

**Expected Output**:
- CSS rule cascade reports
- Specificity conflict list
- Override sources (global CSS vs Tailwind)

---

### Approach 8: Responsive Breakpoint Testing (Viewport-Level)
**Strategy**: Test both apps at different viewport sizes to identify responsive issues

**Steps**:
1. Define test viewports (mobile, tablet, desktop)
2. Take screenshots at each viewport size
3. Compare layouts at each breakpoint
4. Identify responsive-specific issues

**Tools**:
- Chrome DevTools `resize_page` tool
- Screenshot comparison

**Pros**:
- Catches responsive layout bugs
- Tests mobile vs desktop experiences
- Identifies breakpoint configuration issues

**Cons**:
- Multiplies work (test at multiple sizes)
- May not be the root issue if desktop is already broken
- More relevant if layout is correct at one size but not another

**Expected Output**:
- Screenshots at multiple viewport sizes
- Responsive layout comparison
- Breakpoint-specific issues list

---

## Recommended Workflow (Hybrid Approach)

Based on experience with similar debugging tasks, I recommend a **hybrid approach**:

### Phase 1: Quick Visual Assessment
1. **Screenshot Comparison (Approach 1)** - Get quick visual overview
2. Identify which sections look different

### Phase 2: Structural Analysis
3. **DOM Structure Comparison (Approach 2)** - Check for structural issues
4. **Component Isolation (Approach 6)** - Focus on broken sections

### Phase 3: Deep Style Analysis
5. **Computed CSS Comparison (Approach 3)** - Compare styles for key elements
6. **Tailwind Class Audit (Approach 5)** - Check for missing/wrong utilities
7. **CSS Cascade Analysis (Approach 7)** - Identify override conflicts

### Phase 4: Fix and Validate
8. Apply fixes based on findings
9. Re-run visual comparison to validate

---

## Testing Framework Design

We'll create a testing framework that:
1. Runs each approach automatically
2. Generates reports for each approach
3. Scores each approach based on:
   - **Time to complete**
   - **Clarity of results** (how easy to understand findings)
   - **Actionability** (how easy to apply fixes)
   - **Accuracy** (does it find real issues)
4. Recommends best approach(es) for this specific problem

### Success Metrics

For each approach, we'll measure:
- **Execution Time**: How long it takes to run
- **Findings Count**: Number of issues identified
- **Fix Success Rate**: How many findings lead to actual fixes
- **Residual Issues**: How many visual differences remain after applying fixes

---

## Next Steps

1. ✅ Create this strategy document
2. ⏳ Implement testing framework
3. ⏳ Run all approaches on both apps
4. ⏳ Collect and analyze results
5. ⏳ Identify best approach(es)
6. ⏳ Apply fixes to Astro app
7. ⏳ Validate fixes with visual comparison

---

## Notes

- Both apps should be running before testing
- Reference: http://localhost:5174/
- Target: http://localhost:4321/
- Framework: Chrome DevTools MCP for browser automation
- Output: Markdown reports for each approach
