import * as vscode from 'vscode';

export interface MarkerRange {
    range: vscode.Range;
    type: string;
    context?: string;
}

export class MarkerDetector {
    // Pattern matches: // dingo:s:1 or // dingo:start (for backward compatibility)
    private readonly startPattern = /\/\/\s*dingo:(?:s:\d+|start)(?:\s+(\w+))?(?:\s+(.+))?$/i;
    // Pattern matches: // dingo:e:1 or // dingo:end (for backward compatibility)
    private readonly endPattern = /\/\/\s*dingo:(?:e:\d+|end)\s*$/i;
    private readonly generatedVarPattern = /\b(__err\d+|__tmp\d+)\b/g;
    private readonly errVarPattern = /\b(err)\b/g;

    /**
     * Find all DINGO:GENERATED marker ranges in a document
     */
    public findMarkerRanges(document: vscode.TextDocument): MarkerRange[] {
        const markers: MarkerRange[] = [];
        let inBlock = false;
        let blockStart: number | null = null;
        let blockType = 'unknown';
        let blockContext: string | undefined;

        for (let i = 0; i < document.lineCount; i++) {
            const line = document.lineAt(i);
            const text = line.text.trim();

            // Check for block start
            const startMatch = text.match(this.startPattern);
            if (startMatch && !inBlock) {
                inBlock = true;
                blockStart = i;
                blockType = startMatch[1] || 'unknown';
                blockContext = startMatch[2]?.trim();
                continue;
            }

            // Check for block end
            const endMatch = text.match(this.endPattern);
            if (endMatch && inBlock) {
                if (blockStart !== null && blockStart + 1 < i) {
                    // Create range EXCLUDING the marker comment lines
                    // Start at line AFTER the START comment, end at line BEFORE the END comment
                    const startPos = document.lineAt(blockStart + 1).range.start;
                    const endPos = document.lineAt(i - 1).range.end;

                    markers.push({
                        range: new vscode.Range(startPos, endPos),
                        type: blockType,
                        context: blockContext
                    });
                }
                inBlock = false;
                blockStart = null;
                blockType = 'unknown';
                blockContext = undefined;
            }
        }

        // Handle unclosed blocks (shouldn't happen, but be defensive)
        if (inBlock && blockStart !== null) {
            console.warn(`Unclosed DINGO:GENERATED block starting at line ${blockStart + 1}`);
        }

        return markers;
    }

    /**
     * Check if a document contains any dingo markers
     */
    public hasMarkers(document: vscode.TextDocument): boolean {
        const text = document.getText();
        return text.includes('dingo:s:') || text.includes('dingo:start') || text.includes('DINGO:GENERATED:START');
    }

    /**
     * Find marker comment lines (just the START and END lines themselves)
     */
    public findMarkerLines(document: vscode.TextDocument): vscode.Range[] {
        const markerLines: vscode.Range[] = [];

        for (let i = 0; i < document.lineCount; i++) {
            const line = document.lineAt(i);
            const text = line.text.trim();

            // Check if this line is a marker comment
            if (text.match(this.startPattern) || text.match(this.endPattern)) {
                markerLines.push(line.range);
            }
        }

        return markerLines;
    }

    /**
     * Find all generated variable occurrences in a document
     * Returns ranges for variables like __err0, __tmp0, err, etc.
     * Suppresses these variables throughout the entire file
     */
    public findGeneratedVariables(document: vscode.TextDocument): vscode.Range[] {
        const variableRanges: vscode.Range[] = [];

        for (let i = 0; i < document.lineCount; i++) {
            const line = document.lineAt(i);
            const text = line.text;

            // Highlight __err0, __tmp0, etc. everywhere
            this.generatedVarPattern.lastIndex = 0;
            let match;
            while ((match = this.generatedVarPattern.exec(text)) !== null) {
                const startPos = new vscode.Position(i, match.index);
                const endPos = new vscode.Position(i, match.index + match[0].length);
                variableRanges.push(new vscode.Range(startPos, endPos));
            }

            // Highlight "err" everywhere (when reuse_err_variable = true, it's used throughout)
            this.errVarPattern.lastIndex = 0;
            while ((match = this.errVarPattern.exec(text)) !== null) {
                const startPos = new vscode.Position(i, match.index);
                const endPos = new vscode.Position(i, match.index + match[0].length);
                variableRanges.push(new vscode.Range(startPos, endPos));
            }
        }

        return variableRanges;
    }
}
