"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function DailyQuizRedirect() {
  const router = useRouter();

  useEffect(() => {
    // 現在の日付を取得（YYYY-MM-DD形式）
    const today = new Date();
    const formattedDate = today.toISOString().split("T")[0];

    // 日付付きのページにリダイレクト
    router.push(`/daily-quiz/${formattedDate}`);
  }, [router]);

  // リダイレクト中の表示
  return (
    <div className="min-h-screen flex items-center justify-center">
      <p className="text-xl text-gray-300">リダイレクト中...</p>
    </div>
  );
}
