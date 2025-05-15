"use client";
import { useState } from "react";
import { Bar } from "react-chartjs-2";
import { Card } from "./components/Card";
import Link from "next/link";
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
  "T", // 10をTに変更
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
  const [selectedPreset, setSelectedPreset] = useState(
    "SRP BB call vs UTG open"
  ); // デフォルト値を設定
  const [equityData, setEquityData] = useState<[string, number][]>([]);
  const [validationError, setValidationError] = useState<string>("");

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

    const endpoint = "calculate-hand-vs-range";

    try {
      const response = await fetch(`http://localhost:8080/${endpoint}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          yourHand: handRange.toUpperCase(),
          selectedPreset: selectedPreset,
          flopCards: selectedCards.map((card) => {
            // Tを10に変換
            const rank = card.rank === "T" ? "10" : card.rank;
            return `${rank}${card.suit}`;
          }),
        }),
      });

      const data: HandVsRangeResult[] = await response.json();
      setEquityData(data.map((result) => [result.opponentHand, result.equity]));
    } catch (error) {
      console.error("Error calculating equity:", error);
    }
  };

  const data = {
    labels: [], // ラベルは不要になるため空の配列に
    datasets: [
      {
        label: "Equity vs Opponent Hands",
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
        <div className="w-full flex flex-col items-center mb-4">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent mb-4">
            PLO Equity Distribution Graph
          </h1>
          <nav className="flex gap-4">
            <Link
              href={`/daily-quiz/${new Date().toISOString().split('T')[0]}`}
              className="text-blue-400 hover:text-blue-300"
            >
              Today&apos;s Quiz
            </Link>
          </nav>
        </div>

        <section className="card w-full max-w-2xl">
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            <div className="flex flex-col gap-2">
              <label htmlFor="handRange" className="text-gray-300">
                Your Hand:
              </label>
              <input
                type="text"
                id="handRange"
                value={handRange}
                onChange={(e) => setHandRange(e.target.value)}
                className="input"
                placeholder="Enter your hand"
              />
            </div>
            <div className="flex flex-col gap-2">
              <label htmlFor="opponentPreset" className="text-gray-300">
                Opponent range:
              </label>
              <select
                id="opponentPreset"
                value={selectedPreset}
                onChange={(e) => setSelectedPreset(e.target.value)}
                className="input p-2"
              >
                <option value="SRP BB call vs UTG open">
                  SRP BB call vs UTG open
                </option>
                <option value="SRP BB call vs BTN open">
                  SRP BB call vs BTN open
                </option>
                <option value="SRP BTN call vs UTG open">
                  SRP BTN call vs UTG open
                </option>
                <option value="3BP UTG call vs BB 3bet">
                  3BP UTG call vs BB 3bet
                </option>
                <option value="3BP UTG call vs BTN 3bet">
                  3BP UTG call vs BTN 3bet
                </option>
                <option value="3BP BTN call vs BB 3bet">
                  3BP BTN call vs BB 3bet
                </option>
              </select>
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
