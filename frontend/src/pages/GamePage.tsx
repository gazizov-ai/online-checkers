import { ArrowLeft, CircleAlert, Flag, Handshake, Search, Trophy, X } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getGame } from "../api/game";
import { getProfiles, profileName } from "../api/profile";
import { getRating } from "../api/rating";
import { Board } from "../components/Board";
import { MoveNotationPanel } from "../components/MoveNotationPanel";
import { StatusPill } from "../components/StatusPill";
import { friendlyGameError, isGameInputError } from "../game/messages";
import type {
  Color,
  GameRecord,
  GameSnapshot,
  GameStatePayload,
  Move,
  Position,
  Profile,
  Rating,
  User,
} from "../types/domain";
import { friendlyApiError, socketStatusLabel } from "../ui/labels";
import type { SocketStatus } from "../ws/gameSocket";
import {
  connectGameSocket,
  sendDrawOffer,
  sendDrawResponse,
  sendMove,
  sendResign,
} from "../ws/gameSocket";

type GamePageProps = {
  gameId: string;
  user: User;
  onLeave: () => void;
  onFindNewGame: () => void;
  onOpenProfile: (userId: string) => void;
};

function positionKey(position: Position): string {
  return `${position.row}:${position.col}`;
}

function samePosition(a?: Position | null, b?: Position | null): boolean {
  return Boolean(a && b && a.row === b.row && a.col === b.col);
}

function isCaptureMove(snapshot: GameSnapshot, move: Move): boolean {
  const rowStep = Math.sign(move.to.row - move.from.row);
  const colStep = Math.sign(move.to.col - move.from.col);
  let row = move.from.row + rowStep;
  let col = move.from.col + colStep;

  while (row !== move.to.row && col !== move.to.col) {
    if (snapshot.board.cells[row]?.[col]) {
      return true;
    }

    row += rowStep;
    col += colStep;
  }

  return false;
}

export function GamePage({
  gameId,
  user,
  onLeave,
  onFindNewGame,
  onOpenProfile,
}: GamePageProps) {
  const [game, setGame] = useState<GameRecord | null>(null);
  const [state, setState] = useState<GameStatePayload | null>(null);
  const [playerRatings, setPlayerRatings] = useState<Record<string, Rating | null>>({});
  const [playerProfiles, setPlayerProfiles] = useState<Record<string, Profile>>({});
  const [selected, setSelected] = useState<Position | null>(null);
  const [socketStatus, setSocketStatus] = useState<SocketStatus>("closed");
  const [error, setError] = useState<string | null>(null);
  const [confirmingResign, setConfirmingResign] = useState(false);
  const [drawActionPending, setDrawActionPending] = useState(false);
  const [notationHeight, setNotationHeight] = useState(320);
  const boardAreaRef = useRef<HTMLDivElement | null>(null);
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let cancelled = false;

    void getGame(gameId)
      .then((record) => {
        if (!cancelled) {
          setGame(record);
          setState((current) =>
            current ?? {
              game_id: record.id,
              board_state: record.board_state,
              legal_moves: record.legal_moves ?? [],
              move_history: record.move_history ?? [],
              status: record.status,
              current_turn: record.current_turn,
              winner_id: record.winner_id,
              result: record.result,
              finish_reason: record.finish_reason,
              draw_offer_by: record.draw_offer_by,
            },
          );
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

    void getProfiles(playerIds)
      .then((profiles) => {
        if (!cancelled) {
          setPlayerProfiles(
            Object.fromEntries(profiles.map((profile) => [profile.user_id, profile])),
          );
        }
      })
      .catch(() => {
        if (!cancelled) {
          setPlayerProfiles({});
        }
      });

    return () => {
      cancelled = true;
    };
  }, [game]);

  useEffect(() => {
    const socket = connectGameSocket(gameId, {
      onState: (payload) => {
        setDrawActionPending(false);
        setState({
          ...payload,
          legal_moves: payload.legal_moves ?? [],
          move_history: payload.move_history ?? [],
        });
        setSelected(null);
        setError(null);
      },
      onFinished: (payload) => {
        setConfirmingResign(false);
        setState((current) =>
          current
            ? {
                ...current,
                status: "finished",
                winner_id: payload.winner_id,
                result: payload.result,
                finish_reason: payload.finish_reason,
                draw_offer_by: undefined,
              }
            : current,
        );
      },
      onDrawOffered: (payload) => {
        setConfirmingResign(false);
        void payload;
      },
      onDrawDeclined: () => {
        setDrawActionPending(false);
      },
      onError: (message) => {
        setDrawActionPending(false);
        if (!isGameInputError(message)) {
          setError(friendlyGameError(message));
        }
      },
      onStatusChange: setSocketStatus,
    });

    socketRef.current = socket;

    return () => {
      socket.close();
      socketRef.current = null;
    };
  }, [gameId]);

  useEffect(() => {
    const boardArea = boardAreaRef.current;
    if (!boardArea) {
      return;
    }

    const media = window.matchMedia("(max-width: 720px)");
    const updateHeight = () => {
      setNotationHeight(media.matches ? 320 : Math.round(boardArea.getBoundingClientRect().height + 52));
    };
    const observer = new ResizeObserver(updateHeight);

    observer.observe(boardArea);
    media.addEventListener("change", updateHeight);
    updateHeight();

    return () => {
      observer.disconnect();
      media.removeEventListener("change", updateHeight);
    };
  }, []);

  useEffect(() => {
    if (!selected) {
      return;
    }

    const clearSelectionOutsidePiece = (event: MouseEvent) => {
      const target = event.target;
      if (target instanceof Element && target.closest(".piece")) {
        return;
      }

      setSelected(null);
    };

    document.addEventListener("click", clearSelectionOutsidePiece);
    return () => document.removeEventListener("click", clearSelectionOutsidePiece);
  }, [selected]);

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
    if (!state || !playerColor || !isMyTurn || finished) {
      return new Set<string>();
    }

    return new Set((state.legal_moves ?? []).map((move) => positionKey(move.from)));
  }, [finished, isMyTurn, playerColor, state]);

  const selectedMoves = useMemo<Move[]>(() => {
    if (!state || !selected) {
      return [];
    }

    return (state.legal_moves ?? []).filter((move) => samePosition(move.from, selected));
  }, [selected, state]);

  const legalTargets = useMemo(() => {
    return new Set(selectedMoves.map((move) => positionKey(move.to)));
  }, [selectedMoves]);
  const captureTargets = useMemo(() => {
    return new Set(
      snapshot
        ? selectedMoves
            .filter((move) => isCaptureMove(snapshot, move))
            .map((move) => positionKey(move.to))
        : [],
    );
  }, [selectedMoves, snapshot]);

  const connectionLabel = finished ? "Партия завершена" : socketStatusLabel(socketStatus);
  const connectionTone = socketStatus === "open" || finished ? "good" : "warn";
  const whitePlayer = useMemo(
    () =>
      game
        ? {
            color: "white" as const,
            id: game.white_player_id,
            name: profileName(playerProfiles[game.white_player_id]),
            rating: playerRatings[game.white_player_id]?.rating ?? 1000,
          }
        : null,
    [game, playerProfiles, playerRatings],
  );
  const blackPlayer = useMemo(
    () =>
      game
        ? {
            color: "black" as const,
            id: game.black_player_id,
            name: profileName(playerProfiles[game.black_player_id]),
            rating: playerRatings[game.black_player_id]?.rating ?? 1000,
          }
        : null,
    [game, playerProfiles, playerRatings],
  );
  const topPlayer = playerColor === "black" ? whitePlayer : blackPlayer;
  const bottomPlayer = playerColor === "black" ? blackPlayer : whitePlayer;

  const handleSquareClick = useCallback(
    (position: Position) => {
      if (!snapshot || !isMyTurn || finished) {
        return;
      }

      const piece = snapshot.board.cells[position.row]?.[position.col] ?? null;
      const clickedKey = positionKey(position);

      if (!selected) {
        if (piece?.color === playerColor && selectablePieces.has(clickedKey)) {
          setSelected(position);
          setError(null);
        }
        return;
      }

      if (samePosition(selected, position)) {
        setSelected(null);
        setError(null);
        return;
      }

      if (piece?.color === playerColor) {
        if (selectablePieces.has(clickedKey)) {
          setSelected(position);
          setError(null);
        }
        return;
      }

      const selectedMove = selectedMoves.find((move) => samePosition(move.to, position));
      if (!selectedMove) {
        setSelected(null);
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

  const handleResign = useCallback(() => {
    try {
      sendResign(socketRef.current);
      setConfirmingResign(false);
      setError(null);
    } catch (err) {
      setError(friendlyGameError(err instanceof Error ? err.message : "failed to resign"));
    }
  }, []);

  const handleDraw = useCallback(() => {
    try {
      setDrawActionPending(true);
      if (state?.draw_offer_by && state.draw_offer_by !== user.id) {
        sendDrawResponse(socketRef.current, true);
      } else {
        sendDrawOffer(socketRef.current);
      }
      setError(null);
    } catch (err) {
      setDrawActionPending(false);
      setError(friendlyGameError(err instanceof Error ? err.message : "failed to offer draw"));
    }
  }, [state?.draw_offer_by, user.id]);

  const handleDeclineDraw = useCallback(() => {
    try {
      setDrawActionPending(true);
      sendDrawResponse(socketRef.current, false);
      setError(null);
    } catch (err) {
      setDrawActionPending(false);
      setError(friendlyGameError(err instanceof Error ? err.message : "failed to decline draw"));
    }
  }, []);

  const drawOfferedByMe = state?.draw_offer_by === user.id;
  const drawOfferedByOpponent = Boolean(
    state?.draw_offer_by && state.draw_offer_by !== user.id,
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
            draw={state?.result === "draw"}
            onOpenProfile={onOpenProfile}
            player={topPlayer}
            winner={state?.winner_id === topPlayer?.id}
          />

          <div className="game-board-area" ref={boardAreaRef}>
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
            <MoveNotationPanel
              blackPlayerId={game?.black_player_id}
              height={notationHeight}
              history={state?.move_history ?? []}
              whitePlayerId={game?.white_player_id}
            />

            <div className="game-actions">
              {finished ? (
                <button className="primary-button game-actions__new" onClick={onFindNewGame} type="button">
                  <Search aria-hidden size={18} />
                  Найти новую партию
                </button>
              ) : (
                <>
                  {confirmingResign ? (
                    <>
                      <button
                        className="primary-button"
                        disabled={socketStatus !== "open"}
                        onClick={handleResign}
                        type="button"
                      >
                        <Flag aria-hidden size={18} />
                        Сдаться
                      </button>
                      <button
                        className="ghost-button resign-cancel-button"
                        onClick={() => setConfirmingResign(false)}
                        type="button"
                      >
                        <X aria-hidden size={18} />
                        Отмена
                      </button>
                    </>
                  ) : (
                    <>
                      <button
                        className={[
                          "ghost-button",
                          "draw-button",
                          drawOfferedByOpponent ? "draw-button--attention" : "",
                        ].join(" ")}
                        disabled={socketStatus !== "open" || drawOfferedByMe || drawActionPending}
                        onClick={handleDraw}
                        type="button"
                      >
                        <Handshake aria-hidden size={18} />
                        <span>
                          {drawOfferedByMe || drawActionPending ? "Ожидание..." : "Ничья"}
                        </span>
                      </button>
                      {drawOfferedByOpponent ? (
                        <button
                          className="ghost-button draw-decline-button"
                          disabled={socketStatus !== "open" || drawActionPending}
                          onClick={handleDeclineDraw}
                          type="button"
                        >
                          <X aria-hidden size={18} />
                          Отклонить
                        </button>
                      ) : (
                        <button
                          className="ghost-button"
                          disabled={socketStatus !== "open"}
                          onClick={() => setConfirmingResign(true)}
                          type="button"
                        >
                          <Flag aria-hidden size={18} />
                          Сдаться
                        </button>
                      )}
                    </>
                  )}
                </>
              )}
            </div>
          </div>

          <PlayerStrip
            active={state?.current_turn === bottomPlayer?.color && !finished}
            draw={state?.result === "draw"}
            onOpenProfile={onOpenProfile}
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
  draw?: boolean;
  player: {
    color: Color;
    id: string;
    name: string;
    rating: number;
  } | null;
  onOpenProfile: (userId: string) => void;
  winner?: boolean;
};

function PlayerStrip({
  active = false,
  draw = false,
  onOpenProfile,
  player,
  winner = false,
}: PlayerStripProps) {
  return (
    <div className={["player-strip", active ? "player-strip--active" : ""].join(" ")}>
      {player ? (
        <>
          <span className={`player-color-dot player-color-dot--${player.color}`} />
          <button
            className="player-strip__name"
            onClick={() => onOpenProfile(player.id)}
            title={player.id}
            type="button"
          >
            {player.name}
          </button>
          <strong>{player.rating}</strong>
          {winner ? (
            <span className="winner-mark" title="Победитель">
              <Trophy aria-label="Победитель" size={17} strokeWidth={2.4} />
            </span>
          ) : draw ? (
            <span className="draw-mark" title="Ничья">
              <Handshake aria-label="Ничья" size={17} strokeWidth={2.4} />
            </span>
          ) : null}
        </>
      ) : (
        <span className="player-strip__id">Загрузка игрока</span>
      )}
    </div>
  );
}
