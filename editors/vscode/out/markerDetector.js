"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.MarkerDetector = void 0;
const vscode = __importStar(require("vscode"));
class MarkerDetector {
    constructor() {
        // Pattern matches: // dingo:s:1 or // dingo:start (for backward compatibility)
        this.startPattern = /\/\/\s*dingo:(?:s:\d+|start)(?:\s+(\w+))?(?:\s+(.+))?$/i;
        // Pattern matches: // dingo:e:1 or // dingo:end (for backward compatibility)
        this.endPattern = /\/\/\s*dingo:(?:e:\d+|end)\s*$/i;
        this.generatedVarPattern = /\b(__err\d+|__tmp\d+)\b/g;
        this.errVarPattern = /\b(err)\b/g;
    }
    /**
     * Find all DINGO:GENERATED marker ranges in a document
     */
    findMarkerRanges(document) {
        const markers = [];
        let inBlock = false;
        let blockStart = null;
        let blockType = 'unknown';
        let blockContext;
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
    hasMarkers(document) {
        const text = document.getText();
        return text.includes('dingo:s:') || text.includes('dingo:start') || text.includes('DINGO:GENERATED:START');
    }
    /**
     * Find marker comment lines (just the START and END lines themselves)
     */
    findMarkerLines(document) {
        const markerLines = [];
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
    findGeneratedVariables(document) {
        const variableRanges = [];
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
exports.MarkerDetector = MarkerDetector;
//# sourceMappingURL=markerDetector.js.map