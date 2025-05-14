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

export default function DailyQuiz() {
  const [quizResults, setQuizResults] = useState<QuizResult[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [userAnswers, setUserAnswers] = useState<number[]>([]);
  const [results, setResults] = useState<boolean[]>([]);
  const [showResults, setShowResults] = useState<boolean[]>([]);

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

        // ユーザー回答の初期化
        setUserAnswers(new Array(data.length).fill(50)); // デフォルト値50%
        setResults(new Array(data.length).fill(false));
        setShowResults(new Array(data.length).fill(false));
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

  // 解答を提出する
  const handleSubmit = (index: number) => {
    if (!quizResults[index]) return;

    // 正誤判定（±5%の範囲内なら正解）
    const userAnswer = userAnswers[index];
    const correctAnswer = quizResults[index].average_equity;
    const isCorrect = Math.abs(userAnswer - correctAnswer) <= 5;

    // 結果を更新
    const newResults = [...results];
    newResults[index] = isCorrect;
    setResults(newResults);

    // 結果表示を更新
    const newShowResults = [...showResults];
    newShowResults[index] = true;
    setShowResults(newShowResults);
  };

  // スライダーの値を更新
  const handleSliderChange = (index: number, value: number) => {
    const newAnswers = [...userAnswers];
    newAnswers[index] = value;
    setUserAnswers(newAnswers);
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
              href="/daily-quiz"
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
              今日のクイズはまだありません。
            </p>
          </div>
        ) : (
          quizResults.map((result, index) => (
            <section key={result.id} className="card w-full max-w-2xl">
              <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-semibold text-blue-400">
                  問題 {index + 1}: {result.scenario}
                </h2>
              </div>

              <div className="mb-6">
                <p className="text-gray-300 mb-4">
                  以下のシチュエーションで、heroのEquityはいくつでしょう？
                  <br />
                  誤差±5%の範囲まで正解とする
                </p>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-gray-300">
                  <div>
                    <p className="font-semibold">Situation:</p>
                    <p className="text-xl">{result.scenario}</p>
                  </div>
                  <div>
                    <p className="font-semibold">Hero:</p>
                    <p className="text-xl">
                      {formatHandString(result.hero_hand)}
                    </p>
                  </div>
                </div>
              </div>

              <div className="mb-6">
                <label
                  htmlFor={`equity-slider-${index}`}
                  className="block text-gray-300 mb-2"
                >
                  あなたの回答: {userAnswers[index]}%
                </label>
                <input
                  id={`equity-slider-${index}`}
                  type="range"
                  min="0"
                  max="100"
                  step="1"
                  value={userAnswers[index]}
                  onChange={(e) =>
                    handleSliderChange(index, parseInt(e.target.value))
                  }
                  className="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer"
                  disabled={showResults[index]}
                />
                <div className="flex justify-between text-xs text-gray-400 mt-1">
                  <span>0%</span>
                  <span>25%</span>
                  <span>50%</span>
                  <span>75%</span>
                  <span>100%</span>
                </div>
              </div>

              {!showResults[index] ? (
                <button
                  onClick={() => handleSubmit(index)}
                  className="btn-primary w-full"
                >
                  解答する
                </button>
              ) : (
                <div className="mt-4">
                  <div
                    className={`p-4 rounded-lg mb-4 ${
                      results[index]
                        ? "bg-green-900/30 text-green-400"
                        : "bg-red-900/30 text-red-400"
                    }`}
                  >
                    <p className="font-bold text-lg">
                      {results[index] ? "正解！" : "不正解"}
                    </p>
                    <p>
                      あなたの回答: {userAnswers[index]}%<br />
                      正解: {result.average_equity.toFixed(2)}%
                    </p>
                  </div>

                  {result.result && Array.isArray(result.result) && (
                    <div style={{ height: "400px" }}>
                      <Bar
                        data={
                          prepareChartData(result) || {
                            labels: [],
                            datasets: [],
                          }
                        }
                        options={options}
                      />
                    </div>
                  )}
                </div>
              )}
            </section>
          ))
        )}
      </main>
    </div>
  );
}
