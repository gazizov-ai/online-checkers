import { ArrowLeft, CircleAlert, Flag, Handshake, Search, Trophy } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getGame } from "../api/game";
import { getRating } from "../api/rating";
import { Board } from "../components/Board";
import { MoveNotationPanel } from "../components/MoveNotationPanel";
import { StatusPill } from "../components/StatusPill";
import {
  getLegalMovesFrom,
  getSelectablePositions,
  positionKey,
  samePosition,
  type LegalMove,
} from "../game/legalMoves";
import { friendlyGameError } from "../game/messages";
import type { Color, GameRecord, GameStatePayload, Position, Rating, User } from "../types/domain";
import { friendlyApiError, socketStatusLabel } from "../ui/labels";
import type { SocketStatus } from "../ws/gameSocket";
import { connectGameSocket, sendMove } from "../ws/gameSocket";

type GamePageProps = {
  gameId: string;
  user: User;
  onLeave: () => void;
  onFindNewGame: () => void;
};

export function GamePage({ gameId, user, onLeave, onFindNewGame }: GamePageProps) {
  const [game, setGame] = useState<GameRecord | null>(null);
  const [state, setState] = useState<GameStatePayload | null>(null);
  const [playerRatings, setPlayerRatings] = useState<Record<string, Rating | null>>({});
  const [selected, setSelected] = useState<Position | null>(null);
  const [socketStatus, setSocketStatus] = useState<SocketStatus>("closed");
  const [error, setError] = useState<string | null>(null);
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let cancelled = false;

    void getGame(gameId)
      .then((record) => {
        if (!cancelled) {
          setGame(record);
          setState({
            game_id: record.id,
            board_state: record.board_state,
            status: record.status,
            current_turn: record.current_turn,
            winner_id: record.winner_id,
          });
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(friendlyApiError(err, "Не удалось загрузить партию."));
        }
      });

    return () => {
      cancelled = true;
    };
  }, [gameId]);

  useEffect(() => {
    if (!game) {
      return;
    }

    let cancelled = false;
    const playerIds = [game.white_player_id, game.black_player_id];

    void Promise.all(
      playerIds.map(async (playerId) => {
        try {
          return [playerId, await getRating(playerId)] as const;
        } catch {
          return [playerId, null] as const;
        }
      }),
    ).then((ratings) => {
      if (!cancelled) {
        setPlayerRatings(Object.fromEntries(ratings));
      }
    });

    return () => {
      cancelled = true;
    };
  }, [game]);

  useEffect(() => {
    const socket = connectGameSocket(gameId, {
      onState: (payload) => {
        setState(payload);
        setSelected(null);
        setError(null);
      },
      onFinished: (payload) => {
        setState((current) =>
          current
            ? {
                ...current,
                status: "finished",
                winner_id: payload.winner_id,
              }
            : current,
        );
      },
      onError: (message) => setError(friendlyGameError(message)),
      onStatusChange: setSocketStatus,
    });

    socketRef.current = socket;

    return () => {
      socket.close();
      socketRef.current = null;
    };
  }, [gameId]);

  const playerColor = useMemo<Color | undefined>(() => {
    if (!game) {
      return undefined;
    }

    if (game.white_player_id === user.id) {
      return "white";
    }

    if (game.black_player_id === user.id) {
      return "black";
    }

    return undefined;
  }, [game, user.id]);

  const snapshot = state?.board_state;
  const isMyTurn = Boolean(playerColor && state?.current_turn === playerColor);
  const finished = state?.status === "finished";
  const selectablePieces = useMemo(() => {
    if (!snapshot || !playerColor || !isMyTurn || finished) {
      return new Set<string>();
    }

    return getSelectablePositions(snapshot, playerColor);
  }, [finished, isMyTurn, playerColor, snapshot]);

  const selectedMoves = useMemo<LegalMove[]>(() => {
    if (!snapshot || !selected) {
      return [];
    }

    return getLegalMovesFrom(snapshot, selected);
  }, [selected, snapshot]);

  const legalTargets = useMemo(() => {
    return new Set(selectedMoves.map((move) => positionKey(move.to)));
  }, [selectedMoves]);
  const captureTargets = useMemo(() => {
    return new Set(
      selectedMoves.filter((move) => move.capture).map((move) => positionKey(move.to)),
    );
  }, [selectedMoves]);

  const connectionLabel = finished ? "Партия завершена" : socketStatusLabel(socketStatus);
  const connectionTone = socketStatus === "open" || finished ? "good" : "warn";
  const whitePlayer = useMemo(
    () =>
      game
        ? {
            color: "white" as const,
            id: game.white_player_id,
            rating: playerRatings[game.white_player_id]?.rating ?? 1000,
          }
        : null,
    [game, playerRatings],
  );
  const blackPlayer = useMemo(
    () =>
      game
        ? {
            color: "black" as const,
            id: game.black_player_id,
            rating: playerRatings[game.black_player_id]?.rating ?? 1000,
          }
        : null,
    [game, playerRatings],
  );
  const topPlayer = playerColor === "black" ? whitePlayer : blackPlayer;
  const bottomPlayer = playerColor === "black" ? blackPlayer : whitePlayer;

  const handleSquareClick = useCallback(
    (position: Position) => {
      if (!snapshot || !isMyTurn || finished) {
        if (!finished) {
          setError("Сейчас ход соперника.");
        }
        return;
      }

      const piece = snapshot.board.cells[position.row]?.[position.col] ?? null;
      const clickedKey = positionKey(position);

      if (!selected) {
        if (piece?.color === playerColor && selectablePieces.has(clickedKey)) {
          setSelected(position);
          setError(null);
          return;
        }

        if (piece?.color === playerColor) {
          setError(
            snapshot.forced_piece
              ? "Нужно продолжить взятие той же шашкой."
              : "У этой шашки сейчас нет доступного хода.",
          );
          return;
        }

        setError("Выберите свою шашку.");
        return;
      }

      if (samePosition(selected, position)) {
        setSelected(null);
        setError(null);
        return;
      }

      if (piece?.color === playerColor) {
        if (!selectablePieces.has(clickedKey)) {
          setError(
            snapshot.forced_piece
              ? "Нужно продолжить взятие той же шашкой."
              : "Эта шашка сейчас не может ходить.",
          );
          return;
        }

        setSelected(position);
        setError(null);
        return;
      }

      const selectedMove = selectedMoves.find((move) => samePosition(move.to, position));
      if (!selectedMove) {
        setError("Эта клетка недоступна для выбранной шашки.");
        return;
      }

      try {
        sendMove(socketRef.current, selected, position);
        setSelected(null);
        setError(null);
      } catch (err) {
        setError(friendlyGameError(err instanceof Error ? err.message : "failed to send move"));
      }
    },
    [
      finished,
      isMyTurn,
      playerColor,
      selectablePieces,
      selected,
      selectedMoves,
      snapshot,
    ],
  );

  return (
    <section className="game-layout">
      <div className="game-topbar">
        <button className="ghost-button" onClick={onLeave} type="button">
          <ArrowLeft aria-hidden size={18} />
          Лобби
        </button>

        <div className="game-meta">
          <span>{gameId.slice(0, 8)}</span>
          <StatusPill tone={connectionTone}>{connectionLabel}</StatusPill>
        </div>
      </div>

      <div className="game-content">
        <div className="game-table">
          <PlayerStrip
            active={state?.current_turn === topPlayer?.color && !finished}
            player={topPlayer}
            winner={state?.winner_id === topPlayer?.id}
          />

          <div className="game-board-area">
            {snapshot ? (
              <Board
                disabled={finished}
                captureTargets={captureTargets}
                legalTargets={legalTargets}
                onSquareClick={handleSquareClick}
                playerColor={playerColor}
                snapshot={snapshot}
              />
            ) : (
              <div className="board-placeholder">Загрузка</div>
            )}
          </div>

          <div className="game-side-stack">
            <MoveNotationPanel />

            <div className="game-actions">
              {finished ? (
                <button className="primary-button game-actions__new" onClick={onFindNewGame} type="button">
                  <Search aria-hidden size={18} />
                  Найти новую партию
                </button>
              ) : (
                <>
                  <button className="ghost-button" disabled title="Пока не реализовано на бэке" type="button">
                    <Handshake aria-hidden size={18} />
                    Ничья
                  </button>
                  <button className="ghost-button" disabled title="Пока не реализовано на бэке" type="button">
                    <Flag aria-hidden size={18} />
                    Сдаться
                  </button>
                </>
              )}
            </div>
          </div>

          <PlayerStrip
            active={state?.current_turn === bottomPlayer?.color && !finished}
            player={bottomPlayer}
            winner={state?.winner_id === bottomPlayer?.id}
          />
        </div>

        {error ? (
          <p className="inline-error inline-error--with-icon game-error">
            <CircleAlert aria-hidden size={17} />
            {error}
          </p>
        ) : null}
      </div>
    </section>
  );
}

type PlayerStripProps = {
  active?: boolean;
  player: {
    color: Color;
    id: string;
    rating: number;
  } | null;
  winner?: boolean;
};

function PlayerStrip({ active = false, player, winner = false }: PlayerStripProps) {
  return (
    <div className={["player-strip", active ? "player-strip--active" : ""].join(" ")}>
      {player ? (
        <>
          <span className={`player-color-dot player-color-dot--${player.color}`} />
          <span className="player-strip__id" title={player.id}>
            {player.id}
          </span>
          <strong>{player.rating}</strong>
          {winner ? (
            <span className="winner-mark" title="Победитель">
              <Trophy aria-label="Победитель" size={17} strokeWidth={2.4} />
            </span>
          ) : null}
        </>
      ) : (
        <span className="player-strip__id">Загрузка игрока</span>
      )}
    </div>
  );
}
