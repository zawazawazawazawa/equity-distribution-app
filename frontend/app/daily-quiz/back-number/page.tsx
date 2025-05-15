"use client";
import { useState, useEffect } from "react";
import Link from "next/link";

export default function BackNumber() {
  const [dates, setDates] = useState<string[]>([]);

  useEffect(() => {
    // 過去1週間分の日付を計算
    const pastDates: string[] = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date();
      date.setDate(date.getDate() - i);
      pastDates.push(date.toISOString().split("T")[0]); // YYYY-MM-DD形式
    }
    setDates(pastDates);
  }, []);

  // 日付を日本語表示用にフォーマット (YYYY年MM月DD日)
  const formatDateJP = (dateStr: string) => {
    const date = new Date(dateStr);
    return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`;
  };

  return (
    <div className="min-h-screen p-8">
      <main className="flex flex-col items-center gap-8 max-w-7xl mx-auto">
        <div className="w-full flex flex-col items-center mb-4">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent mb-4">
            Daily Quiz Back Number
          </h1>
          <nav className="flex gap-4">
            <Link href="/" className="text-blue-400 hover:text-blue-300">
              Home
            </Link>
            <Link
              href="/daily-quiz/back-number"
              className="text-blue-400 hover:text-blue-300"
            >
              Back Number
            </Link>
          </nav>
        </div>

        <section className="card w-full max-w-2xl">
          <h2 className="text-2xl font-semibold text-blue-400 mb-6">
            過去1週間のDaily Quiz
          </h2>
          <div className="flex flex-col gap-4">
            {dates.map((date) => (
              <Link
                key={date}
                href={`/daily-quiz/${date}`}
                className="p-4 bg-gray-800 hover:bg-gray-700 rounded-lg transition-colors"
              >
                <div className="text-xl text-gray-200">
                  {formatDateJP(date)}のクイズ
                </div>
              </Link>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
