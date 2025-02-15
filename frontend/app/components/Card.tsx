interface CardProps {
  value: string;
  isSelected: boolean;
  onClick: () => void;
}

export const Card: React.FC<CardProps> = ({ value, isSelected, onClick }) => {
  return (
    <div
      className={`
        w-24 h-36 
        border border-solid 
        flex items-center justify-center 
        cursor-pointer 
        transition-transform duration-200
        ${isSelected ? "border-blue-500 scale-110" : "border-gray-300"}
        hover:border-blue-300
      `}
      onClick={onClick}
    >
      {value}
    </div>
  );
};
