interface CardProps {
  value: string;
  isSelected: boolean;
  onClick: () => void;
}

export const Card: React.FC<CardProps> = ({ value, isSelected, onClick }) => {
  // valueが空の場合は何も表示しない
  if (!value) {
    return null;
  }

  // valueからランクとスートを抽出（例：Ac -> A, c）
  const rank = value[0];
  const suit = value[1].toLowerCase();

  // スートに基づいて画像パスを生成
  let suitName = "";
  switch (suit) {
    case "h":
      suitName = "heart";
      break;
    case "d":
      suitName = "diamond";
      break;
    case "c":
      suitName = "club";
      break;
    case "s":
      suitName = "spade";
      break;
  }

  // 10の場合はTを10に変換
  const displayRank = rank === "T" ? "10" : rank;

  // 画像パスを生成
  const imagePath = `/playing_cards/${suitName}/${suitName}_${displayRank}.png`;

  return (
    <div
      className={`
        w-16 h-24 
        bg-white
        border border-solid 
        cursor-pointer 
        transition-transform duration-200
        ${isSelected ? "border-blue-500 scale-110" : "border-gray-300"}
        hover:border-blue-300
        relative
        rounded-lg
        overflow-hidden
      `}
      onClick={onClick}
    >
      <img
        src={imagePath}
        alt={`${displayRank} of ${suitName}`}
        className="w-full h-full object-contain"
      />
    </div>
  );
};
