-- +goose Up
ALTER TABLE games
    ADD COLUMN result TEXT,
    ADD COLUMN finish_reason TEXT,
    ADD COLUMN draw_offer_by UUID;

ALTER TABLE games
    ADD CONSTRAINT games_result_check
    CHECK (
        result IS NULL
        OR result in ('white_win', 'black_win', 'draw')
    );

ALTER TABLE games 
    ADD CONSTRAINT games_finish_reason_check
    CHECK (
        finish_reason IS NULL
        OR finish_reason IN ('checkers_rules', 'resignation', 'draw_agreement')
    );

ALTER TABLE games
    ADD CONSTRAINT games_draw_offer_by_player_check
    CHECK (
        draw_offer_by IS NULL
        OR draw_offer_by = white_player_id
        OR draw_offer_by = black_player_id 
    );

ALTER TABLE games
    ADD CONSTRAINT games_finished_consistency_check
    CHECK (
        (
            status = 'active'
            AND result IS NULL
            AND finish_reason IS NULL
            AND winner_id IS NULL
            AND finished_at IS NULL
        )
        OR
        (
            status = 'finished'
            AND result IS NOT NULL
            AND finish_reason IS NOT NULL
            AND finished_at IS NOT NULL
            AND draw_offer_by IS NULL
            AND (
                (result = 'draw' AND winner_id IS NULL)
                OR
                (result = 'white_win' AND winner_id = white_player_id)
                OR
                (result = 'black_win' AND winner_id = black_player_id)
            )
        )
    );

-- +goose Down
ALTER TABLE games
    DROP CONSTRAINT games_finished_consistency_check,
    DROP CONSTRAINT games_draw_offer_by_player_check,
    DROP CONSTRAINT games_finish_reason_check,
    DROP CONSTRAINT games_result_check;

ALTER TABLE games
    DROP COLUMN draw_offer_by,
    DROP COLUMN finish_reason,
    DROP COLUMN result;
