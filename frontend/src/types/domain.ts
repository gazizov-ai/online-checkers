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

export type Move = {
  from: Position;
  to: Position;
};

export type MoveHistoryItem = {
  turn_number: number;
  player_id: string;
  notation: string;
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
  legal_moves: Move[] | null;
  move_history: MoveHistoryItem[] | null;
  status: GameStatus;
  current_turn: Color;
  winner_id?: string;
  result?: "white_win" | "black_win" | "draw";
  finish_reason?: "checkers_rules" | "resignation" | "draw_agreement";
  draw_offer_by?: string;
};

export type GameRecord = {
  id: string;
  white_player_id: string;
  black_player_id: string;
  status: GameStatus;
  winner_id?: string;
  result?: "white_win" | "black_win" | "draw";
  finish_reason?: "checkers_rules" | "resignation" | "draw_agreement";
  draw_offer_by?: string;
  board_state: GameSnapshot;
  legal_moves: Move[] | null;
  move_history: MoveHistoryItem[] | null;
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

export type Profile = {
  user_id: string;
  username: string;
  display_name?: string;
  country_code?: string;
  avatar_url?: string;
  bio?: string;
  created_at: string;
  updated_at: string;
};

export type UpdateProfileInput = {
  display_name: string;
  country_code: string;
  avatar_url: string;
  bio: string;
};

export type UserGameHistoryItem = {
  game_id: string;
  white_player_id: string;
  black_player_id: string;
  user_color: Color;
  status: GameStatus;
  result?: "white_win" | "black_win" | "draw";
  finish_reason?: "checkers_rules" | "resignation" | "draw_agreement";
  winner_id?: string;
  created_at: string;
  finished_at?: string;
};

export type UserGameHistoryPage = {
  items: UserGameHistoryItem[];
  next_cursor?: string;
};
