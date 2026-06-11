"use client";

import { useState } from "react";

export function VideoModule() {
  const [playing, setPlaying] = useState(false);
  const [speed, setSpeed] = useState(1);

  return (
    <div className="space-y-3">
      <h2 className="text-lg font-semibold flex items-center gap-2">
        <span className="text-red-500">▶</span> Animated Explainer
      </h2>
      <div
        className="relative aspect-video bg-black rounded-xl overflow-hidden flex items-center justify-center cursor-pointer"
        onClick={() => setPlaying(!playing)}
      >
        {!playing ? (
          <div className="flex flex-col items-center gap-3 text-white">
            <div className="w-16 h-16 rounded-full bg-white/20 flex items-center justify-center backdrop-blur-sm">
              <span className="text-2xl ml-1">▶</span>
            </div>
            <p className="text-sm text-white/70">Click to play</p>
          </div>
        ) : (
          <div className="text-white text-center w-full h-full flex flex-col items-center justify-center bg-gradient-to-b from-black/80 to-transparent p-8">
            <p className="text-lg">[ Animated video plays here ]</p>
            <p className="text-sm text-white/60 mt-2">
              Concept explanation with visual diagrams and narration
            </p>
            <div className="flex gap-4 mt-6">
              {([1, 1.5, 2] as const).map((s) => (
                <button
                  key={s}
                  onClick={(e) => {
                    e.stopPropagation();
                    setSpeed(s);
                  }}
                  className={`text-sm px-3 py-1 rounded-full ${
                    speed === s ? "bg-white text-black" : "bg-white/20"
                  }`}
                >
                  {s}x
                </button>
              ))}
            </div>
          </div>
        )}
      </div>
      <div className="flex items-center justify-between text-sm text-muted-foreground">
        <p>Learn the core concept visually. Plays muted with captions by default — click to unmute.</p>
        <span className="text-xs bg-muted px-2 py-1 rounded">2:34 / 3:12</span>
      </div>
    </div>
  );
}
