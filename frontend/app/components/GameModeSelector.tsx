import React from "react";
import { GameMode } from "../../types/poker";

interface GameModeSelectorProps {
  currentMode: GameMode;
  onModeChange: (mode: GameMode) => void;
}

export const GameModeSelector: React.FC<GameModeSelectorProps> = ({
  currentMode,
  onModeChange,
}) => {
  return (
    <div className="flex flex-col gap-4 mb-6">
      <h2 className="text-lg font-semibold">Game Mode</h2>
      <div className="flex gap-4">
        <button
          className={`px-4 py-2 rounded-lg transition-colors ${
            currentMode === "hand-vs-range"
              ? "bg-blue-600 text-white"
              : "bg-gray-200 hover:bg-gray-300"
          }`}
          onClick={() => onModeChange("hand-vs-range")}
        >
          Hand vs Range
        </button>
        <button
          className={`px-4 py-2 rounded-lg transition-colors ${
            currentMode === "range-vs-range"
              ? "bg-blue-600 text-white"
              : "bg-gray-200 hover:bg-gray-300"
          }`}
          onClick={() => onModeChange("range-vs-range")}
        >
          Range vs Range
        </button>
      </div>
    </div>
  );
};
