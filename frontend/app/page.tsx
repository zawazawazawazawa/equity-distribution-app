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
        display: false, // Hide x-axis labels
      },
      y: {
        beginAtZero: true,
      },
    },
  };

  return (
    <div className="min-h-screen p-8">
      <main className="flex flex-col items-center gap-8">
        <h1 className="text-2xl font-bold">Poker Equity Calculator</h1>

        <GameModeSelector
          currentMode={gameState.mode}
          onModeChange={handleModeChange}
        />

        <section className="w-full max-w-2xl">
          <h2 className="text-xl mb-4">
            {gameState.mode === "hand-vs-range"
              ? "Hand vs Range"
              : "Range vs Range"}
          </h2>
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="flex flex-col">
              <label htmlFor="handRange">
                {gameState.mode === "hand-vs-range"
                  ? "Your Single Hand:"
                  : "Your Range:"}
              </label>
              <textarea
                id="handRange"
                value={handRange}
                onChange={(e) => setHandRange(e.target.value)}
                rows={gameState.mode === "hand-vs-range" ? 1 : 4}
                className="border p-2 rounded"
                placeholder={
                  gameState.mode === "hand-vs-range"
                    ? "Enter your hand:"
                    : "Enter your range:"
                }
              />
            </div>
            <div className="flex flex-col">
              <label htmlFor="opponentHandRange">Opponents Range:</label>
              <textarea
                id="opponentHandRange"
                value={opponentHandRange}
                onChange={(e) => setOpponentHandRange(e.target.value)}
                rows={4}
                className="border p-2 rounded"
                placeholder="Enter opponent's range:"
              />
            </div>

            {/* Card Selection Section */}
            <section>
              <h2 className="text-xl mb-4">Select Flop Cards</h2>
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
                <p className="text-red-500 mt-2">{validationError}</p>
              )}
              {selectedCards.length > 0 && (
                <p className="mt-4">
                  Selected Cards:{" "}
                  {selectedCards
                    .map((card) => `${card.rank}${card.suit}`)
                    .join(", ")}
                </p>
              )}
            </section>
            <button
              type="submit"
              className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
            >
              Draw Graph
            </button>
          </form>
        </section>

        {/* Equity Graph */}
        {equityData.length > 0 && (
          <section className="w-full max-w-2xl">
            <h2 className="text-xl mb-4">Equity Distribution</h2>
            <Line data={data} options={options} />
          </section>
        )}
      </main>
    </div>
  );
}
