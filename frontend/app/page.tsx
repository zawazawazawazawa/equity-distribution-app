"use client";
import { useState } from "react";
import { Bar } from "react-chartjs-2";
import { Card } from "./components/Card";
import { GameModeSelector } from "./components/GameModeSelector";
import { GameMode, GameModeState } from "../types/poker";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  TooltipItem,
  Legend,
} from "chart.js";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
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

type HandVsRangeResult = {
  opponentHand: string;
  equity: number;
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

  const formatHandString = (hand: string): string => {
    // @記号で分割し、最初の部分だけを取得
    const cleanHand = hand.split("@")[0];

    // 2文字ごとにカードを分割し、スペースで結合
    const cards = cleanHand.match(/.{2}/g) || [];
    return cards
      .map((card) => {
        const rank = card[0].toUpperCase();
        const suit = card[1].toLowerCase();
        return `${rank}${suit}`;
      })
      .join(" ");
  };

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

    if (selectedCards.filter((card) => card).length !== 3) {
      setValidationError("Please select all three flop cards");
      return;
    }

    const endpoint =
      gameState.mode === "hand-vs-range"
        ? "calculate-hand-vs-range"
        : "calculate-equity";

    try {
      const response = await fetch(`http://localhost:8080/${endpoint}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          yourHand: handRange.toUpperCase(),
          opponentsHands: opponentHandRange.toUpperCase(),
          flopCards: selectedCards.map((card) => `${card.rank}${card.suit}`),
        }),
      });

      if (gameState.mode === "hand-vs-range") {
        const data: HandVsRangeResult[] = await response.json();
        setEquityData(
          data.map((result) => [result.opponentHand, result.equity])
        );
      } else {
        const data: ApiResponse = await response.json();
        if (Array.isArray(data)) {
          setEquityData(data.sort((a, b) => b[1] - a[1]));
        } else if (
          data &&
          typeof data === "object" &&
          Array.isArray(data.equity)
        ) {
          setEquityData(data.equity.sort((a, b) => b[1] - a[1]));
        }
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
    labels: [], // ラベルは不要になるため空の配列に
    datasets: [
      {
        label:
          gameState.mode === "hand-vs-range"
            ? "Equity vs Opponent Hands"
            : "Equity Distribution",
        data: equityData.map((item, index, arr) => ({
          x: arr.length <= 1 ? 0 : (index / (arr.length - 1)) * 100,
          y: Number(item[1]),
        })),
        backgroundColor: "rgba(54, 162, 235, 0.5)",
        borderColor: "rgba(54, 162, 235, 1)",
        borderWidth: 1,
      },
    ],
  };

  // グラフオプションを更新
  const options = {
    responsive: true,
    scales: {
      x: {
        type: "linear" as const,
        min: 5,
        max: 95,
        display: true,
        // grid: {
        //   color: "rgba(255, 255, 255, 0.1)",
        // },
        ticks: {
          color: "#9CA3AF",
          stepSize: 10,
          callback: (tickValue: number | string) => {
            return tickValue + "%";
          },
        },
        title: {
          display: true,
          text: "Percentage of hand",
          color: "#9CA3AF",
        },
      },
      y: {
        beginAtZero: true,
        max: 100,
        grid: {
          color: "rgba(255, 255, 255, 0.1)",
        },
        ticks: {
          color: "#9CA3AF",
        },
        title: {
          display: true,
          text: "Equity %",
          color: "#9CA3AF",
        },
      },
    },
    plugins: {
      legend: {
        display: false, // レジェンドも非表示に
      },
      tooltip: {
        callbacks: {
          title: (tooltipItems: TooltipItem<"bar">[]) => {
            const index = tooltipItems[0]?.dataIndex ?? 0;
            const originalLabel = equityData[index]?.[0];
            return originalLabel ? `Opponent Hand: ${originalLabel}` : "";
          },
          label: (context: TooltipItem<"bar">) => {
            const y =
              typeof context.parsed.y === "number" ? context.parsed.y : 0;
            return `Equity: ${y.toFixed(2)}%`;
          },
        },
      },
    },
    maintainAspectRatio: false,
  };

  return (
    <div className="min-h-screen p-8">
      <main className="flex flex-col items-center gap-8 max-w-7xl mx-auto">
        <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent">
          PLO Equity Distribution Graph
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
              <label htmlFor="opponentHandRange" className="text-gray-300">
                Opponent range:
              </label>
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
                    className="p-2 border rounded bg-gray-800 text-gray-200 border-gray-700"
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
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-2xl font-semibold text-blue-400">
                Hand vs Range Equity
              </h2>
              {/* <div className="text-sm text-gray-400">
                {equityData.length} combinations
              </div> */}
            </div>
            <div style={{ height: "400px" }}>
              <Bar data={data} options={options} />
            </div>
            <div className="mt-6 grid grid-cols-3 gap-4 text-gray-300">
              <div>
                <p className="font-semibold">Your Hand</p>
                <p className="text-xl">{formatHandString(handRange)}</p>
              </div>
              <div>
                <p className="font-semibold">Flop</p>
                <p className="text-xl">
                  {selectedCards
                    .map((card) => `${card.rank}${card.suit}`)
                    .join(" ")}
                </p>
              </div>
              <div>
                <p className="font-semibold">Average Equity</p>
                <p className="text-xl">
                  {(
                    equityData.reduce((sum, pair) => sum + Number(pair[1]), 0) /
                    equityData.length
                  ).toFixed(2)}
                  %
                </p>
              </div>
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
