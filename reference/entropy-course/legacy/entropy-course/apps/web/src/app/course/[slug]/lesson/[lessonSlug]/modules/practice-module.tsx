"use client";

import { useState } from "react";
import { Button } from "@entropy-course/ui/components/button";

export function PracticeModule() {
  const [code, setCode] = useState("// Write your code here\n");
  const [output, setOutput] = useState<string | null>(null);
  const [running, setRunning] = useState(false);

  const handleRun = async () => {
    setRunning(true);
    setOutput(null);
    // Simulate code execution
    await new Promise((r) => setTimeout(r, 1500));
    setOutput(
      "✓ All test cases passed!\n\nTest 1: passed\nTest 2: passed\nTest 3: passed"
    );
    setRunning(false);
  };

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold flex items-center gap-2">
        <span className="text-purple-500">✏</span> Coding Practice
      </h2>
      <p className="text-sm text-muted-foreground">
        Write and run code against hidden test cases. Get instant feedback.
      </p>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div>
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium">Code Editor</span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setCode("// Write your code here\n")}
              >
                Reset
              </Button>
              <Button size="sm" onClick={handleRun} disabled={running}>
                {running ? "Running..." : "▶ Run"}
              </Button>
            </div>
          </div>
          <textarea
            value={code}
            onChange={(e) => setCode(e.target.value)}
            className="w-full h-64 p-4 font-mono text-sm bg-card border rounded-xl resize-none focus:outline-none focus:ring-2 focus:ring-primary/50"
            spellCheck={false}
          />
        </div>
        <div>
          <div className="mb-2">
            <span className="text-sm font-medium">Output</span>
          </div>
          <div className="h-64 p-4 font-mono text-sm bg-card border rounded-xl overflow-auto">
            {output ? (
              <pre className="text-green-600 dark:text-green-400 whitespace-pre-wrap">
                {output}
              </pre>
            ) : (
              <p className="text-muted-foreground">
                Run your code to see output here...
              </p>
            )}
          </div>
        </div>
      </div>

      <div className="flex gap-3">
        <Button variant="outline" size="sm">
          💡 Show Hint
        </Button>
        <Button variant="outline" size="sm">
          📖 Show Solution
        </Button>
        <Button variant="outline" size="sm">
          🔁 Try Harder Version
        </Button>
      </div>
    </div>
  );
}
