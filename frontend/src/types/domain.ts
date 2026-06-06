export type User = {
  id: string;
  username: string;
  email?: string;
  created_at: string;
};

export type Color = "white" | "black";
export type GameStatus = "active" | "finished";
export type SearchStatus = "waiting" | "matching" | "matched";

export type Position = {
  row: number;
  col: number;
};

export type Piece = {
  color: Color;
  king: boolean;
};

export type Board = {
  cells: Array<Array<Piece | null>>;
};

export type GameSnapshot = {
  board: Board;
  turn: Color;
  status: GameStatus;
  winner?: Color;
  forced_piece?: Position;
};

export type GameStatePayload = {
  game_id: string;
  board_state: GameSnapshot;
  status: GameStatus;
  current_turn: Color;
  winner_id?: string;
};

export type GameRecord = {
  id: string;
  white_player_id: string;
  black_player_id: string;
  status: GameStatus;
  winner_id?: string;
  board_state: GameSnapshot;
  current_turn: Color;
};

export type MatchmakingResponse = {
  status: SearchStatus;
  game_id?: string;
};

export type Rating = {
  user_id: string;
  rating: number;
  games_played: number;
  wins: number;
  losses: number;
};
