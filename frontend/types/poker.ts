export type GameMode = 'hand-vs-range' | 'range-vs-range';

export interface GameModeState {
  mode: GameMode;
  heroHand: string;
  villainInput: string | string[];
}