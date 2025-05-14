"use client";
import { useState, useEffect } from "react";
import { Bar } from "react-chartjs-2";
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

// APIから返されるデータの型定義
type HandVsRangeResult = {
  opponentHand: string;
  equity: number;
};

type QuizResult = {
  id: number;
  date: string;
  scenario: string;
  hero_hand: string;
  flop: string;
  result: HandVsRangeResult[];
  average_equity: number;
  created_at: string;
};

export default function DailyQuizResults() {
  const [quizResults, setQuizResults] = useState<QuizResult[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // 翌日の日付を計算（YYYY-MM-DD形式）
  // const getTomorrowDate = () => {
  //   const tomorrow = new Date();
  //   tomorrow.setDate(tomorrow.getDate() + 1);
  //   return tomorrow.toISOString().split("T")[0];
  // };

  useEffect(() => {
    const fetchQuizResults = async () => {
      setLoading(true);
      setError(null);

      try {
        // const date = getTomorrowDate();
        const response = await fetch(
          `http://localhost:8080/api/daily-quiz-results?date=2025-05-16`
        );

        if (!response.ok) {
          throw new Error(`APIリクエストが失敗しました: ${response.status}`);
        }

        const data = await response.json();
        setQuizResults(data);
      } catch (err) {
        console.error("クイズ結果の取得中にエラーが発生しました:", err);
        setError(
          "データの取得中にエラーが発生しました。後でもう一度お試しください。"
        );
      } finally {
        setLoading(false);
      }
    };

    fetchQuizResults();
  }, []);

  // 結果データをグラフ用に変換
  const prepareChartData = (result: QuizResult) => {
    if (!result || !result.result) return null;

    // resultはHandVsRangeResultの配列
    const equityData = Array.isArray(result.result)
      ? result.result.map(
          (item: HandVsRangeResult) =>
            [item.opponentHand, item.equity] as [string, number]
        )
      : [];

    // equityの値で降順（高い順）にソート
    equityData.sort((a, b) => b[1] - a[1]);

    return {
      labels: [], // ラベルは不要になるため空の配列に
      datasets: [
        {
          label: "Equity vs Opponent Hands",
          data: equityData.map(
            (
              item: [string, number],
              index: number,
              arr: [string, number][]
            ) => ({
              x: arr.length <= 1 ? 0 : (index / (arr.length - 1)) * 100,
              y: Number(item[1]),
              // opponentHandの情報を保持
              opponentHand: item[0],
              equity: Number(item[1]),
            })
          ),
          backgroundColor: "rgba(54, 162, 235, 0.5)",
          borderColor: "rgba(54, 162, 235, 1)",
          borderWidth: 1,
        },
      ],
    };
  };

  // グラフオプション
  const options = {
    responsive: true,
    scales: {
      x: {
        type: "linear" as const,
        min: 5,
        max: 95,
        display: true,
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
            // データセットから直接opponentHandを取得
            const dataset = tooltipItems[0]?.dataset;
            if (!dataset || !dataset.data || !Array.isArray(dataset.data))
              return "";

            // 型キャストを避け、安全にプロパティにアクセス
            const dataPoint = dataset.data[index];
            // dataPointがオブジェクトであることを確認してからプロパティにアクセス
            const opponentHand =
              dataPoint &&
              typeof dataPoint === "object" &&
              dataPoint !== null &&
              "opponentHand" in dataPoint
                ? (dataPoint as { opponentHand: string }).opponentHand
                : "";
            return opponentHand ? `Opponent Hand: ${opponentHand}` : "";
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

  // 手札の表示形式を整える
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

  return (
    <div className="min-h-screen p-8">
      <main className="flex flex-col items-center gap-8 max-w-7xl mx-auto">
        <div className="w-full flex flex-col items-center mb-4">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent mb-4">
            Daily Quiz
          </h1>
          <nav className="flex gap-4">
            <Link href="/" className="text-blue-400 hover:text-blue-300">
              ホーム
            </Link>
            <Link
              href="/daily-quiz-results"
              className="text-blue-400 hover:text-blue-300"
            >
              Daily Quiz
            </Link>
          </nav>
        </div>

        {loading ? (
          <div className="card w-full max-w-2xl flex justify-center items-center p-12">
            <p className="text-xl text-gray-300">データを読み込み中...</p>
          </div>
        ) : error ? (
          <div className="card w-full max-w-2xl">
            <p className="text-red-500">{error}</p>
          </div>
        ) : !quizResults || quizResults.length === 0 ? (
          <div className="card w-full max-w-2xl">
            <p className="text-xl text-gray-300">
              この日付のクイズ結果はありません。
            </p>
          </div>
        ) : (
          quizResults.map((result) => (
            <section key={result.id} className="card w-full max-w-2xl">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-semibold text-blue-400">
                  {result.scenario}
                </h2>
                <div className="text-sm text-gray-400">{result.date}</div>
              </div>

              {result.result && Array.isArray(result.result) && (
                <div style={{ height: "400px" }}>
                  <Bar
                    data={
                      prepareChartData(result) || { labels: [], datasets: [] }
                    }
                    options={options}
                  />
                </div>
              )}

              <div className="mt-6 grid grid-cols-3 gap-4 text-gray-300">
                <div>
                  <p className="font-semibold">Your Hand</p>
                  <p className="text-xl">
                    {formatHandString(result.hero_hand)}
                  </p>
                </div>
                <div>
                  <p className="font-semibold">Flop</p>
                  <p className="text-xl">{formatHandString(result.flop)}</p>
                </div>
                <div>
                  <p className="font-semibold">Average Equity</p>
                  <p className="text-xl">{result.average_equity.toFixed(2)}%</p>
                </div>
              </div>
            </section>
          ))
        )}
      </main>
    </div>
  );
}
