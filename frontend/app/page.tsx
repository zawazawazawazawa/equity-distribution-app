"use client";
import { useState } from "react";
import { Line } from "react-chartjs-2";
import { Card } from "./components/Card";
import { GameModeSelector } from "./components/GameModeSelector";
import { GameMode, GameModeState } from "../types/poker";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from "chart.js";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
);

type ApiResponse =
  | Array<[string, number]>
  | { equity: Array<[string, number]> };

type Card = {
  rank: string;
  suit: string;
};

const ranks = [
  "A",
  "K",
  "Q",
  "J",
  "10",
  "9",
  "8",
  "7",
  "6",
  "5",
  "4",
  "3",
  "2",
];
const suits = ["h", "d", "c", "s"];

export default function Home() {
  // 既存のstate
  const [selectedCards, setSelectedCards] = useState<Card[]>([]);
  const [handRange, setHandRange] = useState("");
  const [opponentHandRange, setOpponentHandRange] = useState("");
  const [equityData, setEquityData] = useState<[string, number][]>([]);
  const [validationError, setValidationError] = useState<string>("");

  // 新規追加のstate
  const [gameState, setGameState] = useState<GameModeState>({
    mode: "hand-vs-range", // デフォルトをhand-vs-rangeに変更
    heroHand: "",
    villainInput: "",
  });

  const getAvailableCards = (excludeCards: Card[]) => {
    const availableCards: Card[] = [];
    ranks.forEach((rank) => {
      suits.forEach((suit) => {
        if (
          !excludeCards.some((card) => card.rank === rank && card.suit === suit)
        ) {
          availableCards.push({ rank, suit });
        }
      });
    });
    return availableCards;
  };

  const handleCardSelect = (
    event: React.ChangeEvent<HTMLSelectElement>,
    index: number
  ) => {
    const [rank, suit] = event.target.value.split("");
    const newCard = { rank, suit };

    setSelectedCards((prev) => {
      const newCards = [...prev];
      newCards[index] = newCard;
      return newCards.slice(0, 3); // 最大3枚まで
    });
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setValidationError("");

    // Validate flop cards
    if (selectedCards.filter((card) => card).length !== 3) {
      setValidationError("Please select all three flop cards");
      return;
    }

    try {
      const response = await fetch("http://localhost:8080/calculate-equity", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          yourHands: handRange.toUpperCase(),
          opponentsHands: opponentHandRange.toUpperCase(),
          flopCards: selectedCards.map((card) => `${card.rank}${card.suit}`),
        }),
      });
      const data: ApiResponse = await response.json();
      if (Array.isArray(data)) {
        const sortedData = data.sort((a, b) => b[1] - a[1]);
        setEquityData(sortedData);
      } else if (
        data &&
        typeof data === "object" &&
        Array.isArray(data.equity)
      ) {
        const sortedData = data.equity.sort((a, b) => b[1] - a[1]);
        setEquityData(sortedData);
      } else {
        console.error("Unexpected data format:", data);
      }
    } catch (error) {
      console.error("Error calculating equity:", error);
    }
  };

  const handleModeChange = (newMode: GameMode) => {
    setGameState((prev) => ({
      ...prev,
      mode: newMode,
      villainInput: newMode === "hand-vs-range" ? "" : [],
    }));
    // モード切替時にフォームをリセット
    setHandRange("");
    setOpponentHandRange("");
    setEquityData([]);
  };

  const data = {
    labels: equityData.map((item) => item[0] as string),
    datasets: [
      {
        label: "Equity",
        data: equityData.map((item) => Number(item[1])),
        backgroundColor: "rgba(75, 192, 192, 0.2)",
        borderColor: "rgba(75, 192, 192, 1)",
        borderWidth: 1,
        pointStyle: "circle",
        radius: 1,
        hoverRadius: 10,
      },
    ],
  };

  const options = {
    scales: {
      x: {
        display: false,
      },
      y: {
        beginAtZero: true,
        grid: {
          color: "rgba(255, 255, 255, 0.1)",
        },
        ticks: {
          color: "#9CA3AF",
        },
      },
    },
    plugins: {
      legend: {
        labels: {
          color: "#9CA3AF",
        },
      },
    },
    maintainAspectRatio: false,
  };

  return (
    <div className="min-h-screen p-8">
      <main className="flex flex-col items-center gap-8 max-w-7xl mx-auto">
        <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent">
          Poker Equity Calculator
        </h1>

        <GameModeSelector
          currentMode={gameState.mode}
          onModeChange={handleModeChange}
        />

        <section className="card w-full max-w-2xl">
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            <div className="flex flex-col gap-2">
              <label htmlFor="handRange" className="text-gray-300">
                {gameState.mode === "hand-vs-range"
                  ? "Your Hand:"
                  : "Your Range:"}
              </label>
              <textarea
                id="handRange"
                value={handRange}
                onChange={(e) => setHandRange(e.target.value)}
                rows={gameState.mode === "hand-vs-range" ? 1 : 4}
                className="input"
                placeholder={
                  gameState.mode === "hand-vs-range"
                    ? "Enter your hand"
                    : "Enter your range"
                }
              />
            </div>
            <div className="flex flex-col gap-2">
              <textarea
                id="opponentHandRange"
                value={opponentHandRange}
                onChange={(e) => setOpponentHandRange(e.target.value)}
                rows={4}
                className="input"
                placeholder="Enter opponent's range:"
              />
            </div>

            <section className="border-t border-gray-800 pt-6">
              <div className="flex gap-4">
                {[0, 1, 2].map((index) => (
                  <select
                    key={index}
                    className="p-2 border rounded"
                    value={
                      selectedCards[index]
                        ? `${selectedCards[index].rank}${selectedCards[index].suit}`
                        : ""
                    }
                    onChange={(e) => handleCardSelect(e, index)}
                  >
                    <option value="">Select Card</option>
                    {getAvailableCards(
                      selectedCards.filter((_, i) => i !== index)
                    ).map((card) => (
                      <option
                        key={`${card.rank}${card.suit}`}
                        value={`${card.rank}${card.suit}`}
                      >
                        {card.rank}
                        {card.suit}
                      </option>
                    ))}
                  </select>
                ))}
              </div>
              {validationError && (
                <p className="text-red-400 mt-2">{validationError}</p>
              )}
              {selectedCards.length > 0 && (
                <p className="text-gray-400 mt-4">
                  Selected Cards:{" "}
                  {selectedCards
                    .map((card) => `${card.rank}${card.suit}`)
                    .join(", ")}
                </p>
              )}
            </section>

            <button type="submit" className="btn-primary">
              Calculate Equity
            </button>
          </form>
        </section>

        {equityData.length > 0 && (
          <section className="card w-full max-w-2xl">
            <h2 className="text-2xl font-semibold mb-6 text-blue-400">
              Equity Distribution
            </h2>
            <Line data={data} options={options} />
          </section>
        )}
      </main>
    </div>
  );
}
