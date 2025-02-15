"use client";
import { useState } from "react";
import { Line } from "react-chartjs-2";
import { Card } from "./components/Card";
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

export default function Home() {
  const [selectedCard, setSelectedCard] = useState<string | null>(null);
  const cards = ["A", "K", "Q"];
  const [handRange, setHandRange] = useState("");
  const [opponentHandRange, setOpponentHandRange] = useState("");
  const [equityData, setEquityData] = useState<[string, number][]>([]);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    try {
      const response = await fetch("http://localhost:8080/calculate-equity", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          yourHands: handRange.toUpperCase(),
          opponentsHands: opponentHandRange.toUpperCase(),
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
        <h1 className="text-2xl font-bold">ポーカーエクイティ計算</h1>

        {/* カード選択セクション */}
        <section>
          <h2 className="text-xl mb-4">フロップカード選択</h2>
          <div className="flex gap-4">
            {cards.map((card) => (
              <Card
                key={card}
                value={card}
                isSelected={selectedCard === card}
                onClick={() => setSelectedCard(card)}
              />
            ))}
          </div>
          {selectedCard && (
            <p className="mt-4">選択されたカード: {selectedCard}</p>
          )}
        </section>

        {/* エクイティ計算フォーム */}
        <section className="w-full max-w-2xl">
          <h2 className="text-xl mb-4">ハンドレンジ入力</h2>
          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="flex flex-col">
              <label htmlFor="opponentHandRange">
                相手のハンドレンジ (例: AhKdQsJc,AsKdQhJc):
              </label>
              <textarea
                id="opponentHandRange"
                value={opponentHandRange}
                onChange={(e) => setOpponentHandRange(e.target.value)}
                rows={4}
                className="border p-2 rounded"
              />
            </div>
            <div className="flex flex-col">
              <label htmlFor="handRange">
                自分のハンドレンジ (例: AhKdQsJc,AsKdQhJc):
              </label>
              <textarea
                id="handRange"
                value={handRange}
                onChange={(e) => setHandRange(e.target.value)}
                rows={4}
                className="border p-2 rounded"
              />
            </div>
            <button
              type="submit"
              className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
            >
              グラフを描画
            </button>
          </form>
        </section>

        {/* エクイティグラフ */}
        {equityData.length > 0 && (
          <section className="w-full max-w-2xl">
            <h2 className="text-xl mb-4">エクイティ分布</h2>
            <Line data={data} options={options} />
          </section>
        )}
      </main>
    </div>
  );
}
