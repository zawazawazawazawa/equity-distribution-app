"use client";
import { useState } from "react";
import { Line } from "react-chartjs-2";
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
    <div>
      <h1>Equity Distribution Graph</h1>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="opponentHandRange">
            Opponent Hand Range (Format: AhKdQsJc,AsKdQhJc):
          </label>
          <textarea
            id="opponentHandRange"
            value={opponentHandRange}
            onChange={(e) => setOpponentHandRange(e.target.value as string)}
            rows={4}
            cols={50}
          />
        </div>
        <div>
          <label htmlFor="handRange">
            Your Hand Range (Format: AhKdQsJc,AsKdQhJc):
          </label>
          <textarea
            id="handRange"
            value={handRange}
            onChange={(e) => setHandRange(e.target.value as string)}
            rows={4}
            cols={50}
          />
        </div>
        <button type="submit">Draw Graph</button>
      </form>
      {equityData.length > 0 && (
        <div>
          <h2>Equity Distribution</h2>
          <Line data={data} options={options} />
        </div>
      )}
    </div>
  );
}
