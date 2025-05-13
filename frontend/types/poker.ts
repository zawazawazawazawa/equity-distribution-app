export type GameMode = 'hand-vs-range';

export interface GameModeState {
  mode: GameMode;
  heroHand: string;
  villainInput: string;
}
