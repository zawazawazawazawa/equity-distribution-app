export type Suit = "h" | "d" | "c" | "s";
export type Rank = "A" | "K" | "Q" | "J" | "T" | "9" | "8" | "7" | "6" | "5" | "4" | "3" | "2";
export type Card = `${Rank}${Suit}`;

export const SUITS: Suit[] = ["h", "d", "c", "s"];