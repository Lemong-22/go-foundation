"use client";

import { useState } from "react";
import { Button } from "@entropy-course/ui/components/button";
import { Card } from "@entropy-course/ui/components/card";
import type { CodeSnippet } from "@/lib/mock-data";

interface Props {
  snippets: CodeSnippet[];
}

export function CodeModule({ snippets }: Props) {
  const [copied, setCopied] = useState<string | null>(null);

  const handleCopy = async (code: string, id: string) => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(id);
      setTimeout(() => setCopied(null), 2000);
    } catch {
      // Fallback for environments without clipboard API
      const textarea = document.createElement("textarea");
      textarea.value = code;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
      setCopied(id);
      setTimeout(() => setCopied(null), 2000);
    }
  };

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold flex items-center gap-2">
        <span className="text-green-500">💻</span> Copy-Paste Code
      </h2>
      <div className="space-y-4">
        {snippets.map((snippet) => (
          <Card key={snippet.id} className="overflow-hidden">
            <div className="p-3 bg-muted/50 border-b flex items-center justify-between">
              <div>
                <p className="font-medium text-sm">{snippet.title}</p>
                <p className="text-xs text-muted-foreground">{snippet.description}</p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleCopy(snippet.code, snippet.id)}
              >
                {copied === snippet.id ? "✓ Copied!" : "Copy"}
              </Button>
            </div>
            <pre className="p-4 text-sm font-mono overflow-x-auto bg-card max-h-80 overflow-y-auto">
              <code>{snippet.code}</code>
            </pre>
          </Card>
        ))}
      </div>
    </div>
  );
}
