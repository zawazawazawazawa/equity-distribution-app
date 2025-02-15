export type GameMode = 'hand-vs-hand' | 'hand-vs-range';

export interface GameModeState {
  mode: GameMode;
  heroHand: string;
  villainInput: string | string[]; // handの場合はstring、rangeの場合はstring[]
}