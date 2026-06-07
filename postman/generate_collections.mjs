import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.dirname(fileURLToPath(import.meta.url));
const schema = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json";

const jsonHeaders = [{ key: "Content-Type", value: "application/json" }];
const bearer = (variable) => ({
  type: "bearer",
  bearer: [{ key: "token", value: `{{${variable}}}`, type: "string" }],
});

function script(exec) {
  return { listen: "test", script: { type: "text/javascript", exec } };
}

function request(name, method, url, options = {}) {
  const req = {
    method,
    header: options.headers ?? [],
    url,
  };
  if (options.auth) req.auth = options.auth;
  if (options.body !== undefined) {
    req.body = { mode: "raw", raw: options.body, options: { raw: { language: "json" } } };
  }
  return {
    name,
    request: req,
    event: [script(options.tests ?? [])],
  };
}

function folder(name, item) {
  return { name, item };
}

const common = {
  okJSON: [
    'pm.test("HTTP 200", () => pm.response.to.have.status(200));',
    'pm.test("JSON response", () => pm.expect(pm.response.headers.get("Content-Type")).to.include("application/json"));',
  ],
  error(status, code) {
    return [
      `pm.test("HTTP ${status}", () => pm.response.to.have.status(${status}));`,
      `pm.test("error code ${code}", () => { const body = pm.response.json(); pm.expect(body).to.have.nested.property("error.code", "${code}"); pm.expect(body.error.message).to.be.a("string").and.not.empty; });`,
    ];
  },
};

const initScript = [
  'if (!pm.environment.get("runId")) {',
  '  const runId = Date.now().toString(36) + Math.random().toString(36).slice(2, 8);',
  '  pm.environment.set("runId", runId);',
  '  pm.environment.set("usernameA", "e2e_a_" + runId);',
  '  pm.environment.set("usernameB", "e2e_b_" + runId);',
  '  pm.environment.set("usernameC", "e2e_c_" + runId);',
  '  pm.environment.set("emailA", "e2e_a_" + runId + "@example.test");',
  '}',
];

const contract = {
  info: {
    _postman_id: "d4fc8e20-103e-4944-bf85-867771e319a2",
    name: "Online Checkers - Identity and Profile Contracts",
    description: "Rigorous black-box checks for service readiness, OIDC metadata, auth failure modes, user registration, login, JWT usage, and asynchronous profile projection.",
    schema,
  },
  event: [{ listen: "prerequest", script: { type: "text/javascript", exec: initScript } }],
  item: [
    folder("Readiness and identity metadata", [
      ...["auth", "matchmaking", "game", "rating", "profile"].map((service) =>
        request(`${service} health`, "GET", `{{${service}Url}}/health`, {
          tests: [
            ...common.okJSON,
            `pm.test("service identity", () => { const body = pm.response.json(); pm.expect(body.status).to.eql("ok"); pm.expect(body.service).to.eql("${service}-service"); });`,
            'pm.test("bounded latency", () => pm.expect(pm.response.responseTime).to.be.below(2000));',
          ],
        }),
      ),
      request("OIDC discovery", "GET", "{{authUrl}}/.well-known/openid-configuration", {
        tests: [
          ...common.okJSON,
          'pm.test("issuer and algorithms", () => { const body = pm.response.json(); pm.expect(body.issuer).to.eql(pm.environment.get("authUrl")); pm.expect(body.jwks_uri).to.eql(pm.environment.get("authUrl") + "/.well-known/jwks.json"); pm.expect(body.id_token_signing_alg_values_supported).to.include("RS256"); });',
        ],
      }),
      request("JWKS", "GET", "{{authUrl}}/.well-known/jwks.json", {
        tests: [
          ...common.okJSON,
          'pm.test("usable RSA signing key", () => { const body = pm.response.json(); pm.expect(body.keys).to.have.lengthOf.at.least(1); const key = body.keys[0]; pm.expect(key.kty).to.eql("RSA"); pm.expect(key.alg).to.eql("RS256"); pm.expect(key.use).to.eql("sig"); pm.expect(key.n).to.be.a("string").and.not.empty; pm.expect(key.e).to.be.a("string").and.not.empty; });',
        ],
      }),
    ]),
    folder("Auth rejection paths", [
      request("Me without credentials", "GET", "{{authUrl}}/api/v1/me", {
        tests: common.error(401, "unauthorized"),
      }),
      request("Me with malformed bearer", "GET", "{{authUrl}}/api/v1/me", {
        auth: { type: "bearer", bearer: [{ key: "token", value: "not-a-jwt", type: "string" }] },
        tests: common.error(401, "invalid_token"),
      }),
      request("Register malformed JSON", "POST", "{{authUrl}}/api/v1/register", {
        headers: jsonHeaders,
        body: '{"username":',
        tests: common.error(400, "invalid_request"),
      }),
      request("Register unknown field", "POST", "{{authUrl}}/api/v1/register", {
        headers: jsonHeaders,
        body: '{"username":"unknown_field","password":"password-123","admin":true}',
        tests: common.error(400, "invalid_request"),
      }),
      request("Register empty username", "POST", "{{authUrl}}/api/v1/register", {
        headers: jsonHeaders,
        body: '{"username":"","password":"password-123"}',
        tests: common.error(400, "invalid_username"),
      }),
      request("Register short password", "POST", "{{authUrl}}/api/v1/register", {
        headers: jsonHeaders,
        body: '{"username":"short_{{runId}}","password":"short"}',
        tests: common.error(400, "invalid_password"),
      }),
    ]),
    folder("Registration constraints", [
      ...[
        ["A", "usernameA", ',"email":"{{emailA}}"'],
        ["B", "usernameB", ""],
        ["C", "usernameC", ""],
      ].map(([label, username, email]) =>
        request(`Register user ${label}`, "POST", "{{authUrl}}/api/v1/register", {
          headers: jsonHeaders,
          body: `{"username":"{{${username}}}","password":"password-123"${email}}`,
          tests: [
            'pm.test("HTTP 201", () => pm.response.to.have.status(201));',
            `pm.test("safe user ${label}", () => { const body = pm.response.json(); pm.expect(body.user.id).to.match(/^[0-9a-f-]{36}$/); pm.expect(body.user.username).to.eql(pm.environment.get("${username}")); pm.expect(body.user).not.to.have.property("password_hash"); pm.environment.set("user${label}Id", body.user.id); });`,
          ],
        }),
      ),
      request("Duplicate username", "POST", "{{authUrl}}/api/v1/register", {
        headers: jsonHeaders,
        body: '{"username":"{{usernameA}}","email":"other_{{runId}}@example.test","password":"password-123"}',
        tests: common.error(409, "username_taken"),
      }),
      request("Duplicate email", "POST", "{{authUrl}}/api/v1/register", {
        headers: jsonHeaders,
        body: '{"username":"other_{{runId}}","email":"{{emailA}}","password":"password-123"}',
        tests: common.error(409, "email_taken"),
      }),
    ]),
    folder("Login and token semantics", [
      request("Wrong password", "POST", "{{authUrl}}/api/v1/login", {
        headers: jsonHeaders,
        body: '{"username":"{{usernameA}}","password":"wrong-password"}',
        tests: common.error(401, "invalid_credentials"),
      }),
      request("Login unknown user", "POST", "{{authUrl}}/api/v1/login", {
        headers: jsonHeaders,
        body: '{"username":"missing_{{runId}}","password":"password-123"}',
        tests: common.error(401, "invalid_credentials"),
      }),
      ...["A", "B", "C"].map((label) =>
        request(`Login user ${label}`, "POST", "{{authUrl}}/api/v1/login", {
          headers: jsonHeaders,
          body: `{"username":"{{username${label}}}","password":"password-123"}`,
          tests: [
            ...common.okJSON,
            `pm.test("token pair ${label}", () => { const body = pm.response.json(); pm.expect(body.token_type).to.eql("Bearer"); pm.expect(body.expires_in).to.be.above(0); pm.expect(body.access_token.split(".")).to.have.lengthOf(3); pm.expect(body.id_token.split(".")).to.have.lengthOf(3); pm.expect(body.user.id).to.eql(pm.environment.get("user${label}Id")); pm.environment.set("token${label}", body.access_token); ${label === "A" ? 'pm.environment.set("idTokenA", body.id_token);' : ""} });`,
          ],
        }),
      ),
      request("ID token rejected as access token", "GET", "{{authUrl}}/api/v1/me", {
        auth: bearer("idTokenA"),
        tests: common.error(401, "invalid_token"),
      }),
      request("Access token resolves subject", "GET", "{{authUrl}}/api/v1/me", {
        auth: bearer("tokenA"),
        tests: [
          ...common.okJSON,
          'pm.test("correct subject and safe fields", () => { const body = pm.response.json(); pm.expect(body.user.id).to.eql(pm.environment.get("userAId")); pm.expect(body.user.username).to.eql(pm.environment.get("usernameA")); pm.expect(body.user).not.to.have.property("password_hash"); });',
        ],
      }),
    ]),
    folder("Profile projection and validation", [
      request("Wait for user A profile projection", "GET", "{{profileUrl}}/api/v1/profiles/{{userAId}}", {
        tests: [
          'const deadline = Date.now() + 20000;',
          'function poll() { pm.sendRequest(pm.environment.get("profileUrl") + "/api/v1/profiles/" + pm.environment.get("userAId"), (err, res) => { if (!err && res.code === 200) { const body = res.json(); pm.test("profile projection created", () => { pm.expect(body.user_id).to.eql(pm.environment.get("userAId")); pm.expect(body.username).to.eql(pm.environment.get("usernameA")); }); return; } if (Date.now() >= deadline) { pm.test("profile projection created", () => pm.expect.fail(err || "last HTTP status " + (res && res.code))); return; } setTimeout(poll, 250); }); }',
          'if (pm.response.code === 200) { const body = pm.response.json(); pm.test("profile projection created", () => pm.expect(body.user_id).to.eql(pm.environment.get("userAId"))); } else { poll(); }',
        ],
      }),
      request("Wait for all profiles via batch with duplicate IDs", "POST", "{{profileUrl}}/api/v1/profiles/batch", {
        headers: jsonHeaders,
        body: '{"user_ids":["{{userAId}}","{{userBId}}","{{userAId}}","{{userCId}}"]}',
        tests: [
          'const deadline = Date.now() + 20000; const url = pm.environment.get("profileUrl") + "/api/v1/profiles/batch"; const body = {user_ids:[pm.environment.get("userAId"),pm.environment.get("userBId"),pm.environment.get("userAId"),pm.environment.get("userCId")]};',
          'function verify(res) { const data = res.json(); pm.expect(data.profiles).to.have.lengthOf(3); pm.expect(new Set(data.profiles.map(p => p.user_id)).size).to.eql(3); }',
          'function poll() { pm.sendRequest({url, method:"POST", header:{"Content-Type":"application/json"}, body:{mode:"raw",raw:JSON.stringify(body)}}, (err,res) => { if (!err && res.code === 200 && res.json().profiles.length === 3) { pm.test("batch is eventually complete and deduplicated", () => verify(res)); return; } if (Date.now() >= deadline) { pm.test("batch is eventually complete and deduplicated", () => pm.expect.fail(err || "profiles missing")); return; } setTimeout(poll,250); }); }',
          'if (pm.response.code === 200 && pm.response.json().profiles.length === 3) { pm.test("batch is eventually complete and deduplicated", () => verify(pm.response)); } else { poll(); }',
        ],
      }),
      request("Profile invalid UUID", "GET", "{{profileUrl}}/api/v1/profiles/not-a-uuid", {
        tests: common.error(400, "invalid_request"),
      }),
      request("Profile missing UUID", "GET", "{{profileUrl}}/api/v1/profiles/00000000-0000-4000-8000-000000000099", {
        tests: common.error(404, "profile_not_found"),
      }),
      request("Batch over 100 IDs", "POST", "{{profileUrl}}/api/v1/profiles/batch", {
        headers: jsonHeaders,
        body: '{"user_ids":[' + Array.from({ length: 101 }, () => '"00000000-0000-4000-8000-000000000099"').join(",") + "]}",
        tests: common.error(400, "too_many_profiles"),
      }),
      request("Patch profile with invalid bearer", "PATCH", "{{profileUrl}}/api/v1/profiles/me", {
        auth: { type: "bearer", bearer: [{ key: "token", value: "not-a-jwt", type: "string" }] },
        headers: jsonHeaders,
        body: '{"display_name":"Injected"}',
        tests: common.error(401, "invalid_token"),
      }),
      request("Reject invalid country code", "PATCH", "{{profileUrl}}/api/v1/profiles/me", {
        auth: bearer("tokenA"),
        headers: jsonHeaders,
        body: '{"country_code":"USA"}',
        tests: common.error(400, "invalid_country_code"),
      }),
      request("Reject non-http avatar URL", "PATCH", "{{profileUrl}}/api/v1/profiles/me", {
        auth: bearer("tokenA"),
        headers: jsonHeaders,
        body: '{"avatar_url":"javascript:alert(1)"}',
        tests: common.error(400, "invalid_avatar_url"),
      }),
      request("Reject oversized Unicode display name", "PATCH", "{{profileUrl}}/api/v1/profiles/me", {
        auth: bearer("tokenA"),
        headers: jsonHeaders,
        body: '{"display_name":"' + "Я".repeat(41) + '"}',
        tests: common.error(400, "invalid_display_name"),
      }),
      request("Reject oversized bio", "PATCH", "{{profileUrl}}/api/v1/profiles/me", {
        auth: bearer("tokenA"),
        headers: jsonHeaders,
        body: '{"bio":"' + "x".repeat(301) + '"}',
        tests: common.error(400, "invalid_bio"),
      }),
      request("Normalize valid profile update", "PATCH", "{{profileUrl}}/api/v1/profiles/me", {
        auth: bearer("tokenA"),
        headers: jsonHeaders,
        body: '{"display_name":"  Alice CI  ","country_code":" us ","avatar_url":" https://example.test/avatar.png ","bio":"  edge-case bio  "}',
        tests: [
          ...common.okJSON,
          'pm.test("values normalized", () => { const body = pm.response.json(); pm.expect(body.display_name).to.eql("Alice CI"); pm.expect(body.country_code).to.eql("US"); pm.expect(body.avatar_url).to.eql("https://example.test/avatar.png"); pm.expect(body.bio).to.eql("edge-case bio"); });',
        ],
      }),
      request("Whitespace update preserves existing values", "PATCH", "{{profileUrl}}/api/v1/profiles/me", {
        auth: bearer("tokenA"),
        headers: jsonHeaders,
        body: '{"display_name":"   ","bio":"  "}',
        tests: [
          ...common.okJSON,
          'pm.test("blank optional fields do not erase persisted values", () => { const body = pm.response.json(); pm.expect(body.display_name).to.eql("Alice CI"); pm.expect(body.bio).to.eql("edge-case bio"); });',
        ],
      }),
    ]),
  ],
};

const workflows = {
  info: {
    _postman_id: "468d8024-ad09-4c53-95db-a43e293985f7",
    name: "Online Checkers - Rating, Matchmaking and Game Workflows",
    description: "Stateful cross-service checks. Requires the identity/profile collection and the game.finished fixture publisher to run first.",
    schema,
  },
  item: [
    folder("Rating choreography and boundary cases", [
      request("Wait for winner rating", "GET", "{{ratingUrl}}/api/v1/ratings/{{userAId}}", {
        tests: [
          'const deadline = Date.now() + 20000; const url = pm.environment.get("ratingUrl") + "/api/v1/ratings/" + pm.environment.get("userAId");',
          'function verify(res) { const body = res.json(); pm.expect(body.rating).to.eql(1025); pm.expect(body.games_played).to.eql(1); pm.expect(body.wins).to.eql(1); pm.expect(body.losses).to.eql(0); }',
          'function poll() { pm.sendRequest(url, (err,res) => { if (!err && res.code === 200) { pm.test("winner projection is consistent", () => verify(res)); return; } if (Date.now() >= deadline) { pm.test("winner projection is consistent", () => pm.expect.fail(err || "rating missing")); return; } setTimeout(poll,250); }); }',
          'if (pm.response.code === 200) { pm.test("winner projection is consistent", () => verify(pm.response)); } else { poll(); }',
        ],
      }),
      request("Loser rating", "GET", "{{ratingUrl}}/api/v1/ratings/{{userBId}}", {
        tests: [
          ...common.okJSON,
          'pm.test("loser projection is consistent", () => { const body = pm.response.json(); pm.expect(body.rating).to.eql(975); pm.expect(body.games_played).to.eql(1); pm.expect(body.wins).to.eql(0); pm.expect(body.losses).to.eql(1); });',
        ],
      }),
      request("Unrated user is not synthesized", "GET", "{{ratingUrl}}/api/v1/ratings/{{userCId}}", {
        tests: common.error(404, "not_found"),
      }),
      request("Rating invalid UUID", "GET", "{{ratingUrl}}/api/v1/ratings/not-a-uuid", {
        tests: common.error(400, "invalid_request"),
      }),
      request("Leaderboard rejects zero limit", "GET", "{{ratingUrl}}/api/v1/leaderboard?limit=0", {
        tests: common.error(400, "invalid_request"),
      }),
      request("Leaderboard rejects non-numeric limit", "GET", "{{ratingUrl}}/api/v1/leaderboard?limit=NaN", {
        tests: common.error(400, "invalid_request"),
      }),
      request("Leaderboard caps huge limit and sorts ratings", "GET", "{{ratingUrl}}/api/v1/leaderboard?limit=100000", {
        tests: [
          ...common.okJSON,
          'pm.test("ordered and bounded", () => { const items = pm.response.json().items; pm.expect(items.length).to.be.at.most(100); pm.expect(items[0].user_id).to.eql(pm.environment.get("userAId")); pm.expect(items[0].rating).to.be.at.least(items[1].rating); });',
        ],
      }),
    ]),
    folder("Matchmaking state machine", [
      request("Status rejects invalid token", "GET", "{{matchmakingUrl}}/api/v1/matchmaking/status", {
        auth: { type: "bearer", bearer: [{ key: "token", value: "not-a-jwt", type: "string" }] },
        tests: common.error(401, "invalid_token"),
      }),
      request("Initial status is waiting without queue row", "GET", "{{matchmakingUrl}}/api/v1/matchmaking/status", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("virtual waiting state", () => { const body = pm.response.json(); pm.expect(body.status).to.eql("waiting"); pm.expect(body).not.to.have.property("game_id"); });'],
      }),
      request("Cancel absent search is idempotent", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/cancel", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("cancelled", () => pm.expect(pm.response.json().status).to.eql("cancelled"));'],
      }),
      request("User A starts waiting", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/search", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("waiting without game", () => { const body = pm.response.json(); pm.expect(body.status).to.eql("waiting"); pm.expect(body).not.to.have.property("game_id"); });'],
      }),
      request("Repeated search does not duplicate queue entry", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/search", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("still waiting", () => pm.expect(pm.response.json().status).to.eql("waiting"));'],
      }),
      request("User B atomically creates match", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/search", {
        auth: bearer("tokenB"),
        tests: [
          ...common.okJSON,
          'pm.test("matched with game id", () => { const body = pm.response.json(); pm.expect(body.status).to.eql("matched"); pm.expect(body.game_id).to.match(/^[0-9a-f-]{36}$/); pm.environment.set("gameId", body.game_id); });',
        ],
      }),
      request("User B repeated search starts a fresh cycle", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/search", {
        auth: bearer("tokenB"),
        tests: [...common.okJSON, 'pm.test("completed match is cleared for this user", () => { const body = pm.response.json(); pm.expect(body.status).to.eql("waiting"); pm.expect(body).not.to.have.property("game_id"); });'],
      }),
      request("User B cancels fresh search", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/cancel", {
        auth: bearer("tokenB"),
        tests: [...common.okJSON, 'pm.test("fresh search cancelled", () => pm.expect(pm.response.json().status).to.eql("cancelled"));'],
      }),
      request("User A consumes matched result", "GET", "{{matchmakingUrl}}/api/v1/matchmaking/status", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("consume-on-read returns the shared game once", () => { const body = pm.response.json(); pm.expect(body.status).to.eql("matched"); pm.expect(body.game_id).to.eql(pm.environment.get("gameId")); });'],
      }),
      request("Consumed match cannot be replayed by status", "GET", "{{matchmakingUrl}}/api/v1/matchmaking/status", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("second read returns virtual waiting state", () => { const body = pm.response.json(); pm.expect(body.status).to.eql("waiting"); pm.expect(body).not.to.have.property("game_id"); });'],
      }),
      request("Cancel after matched result was consumed is idempotent", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/cancel", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("nothing remains to cancel", () => pm.expect(pm.response.json().status).to.eql("cancelled"));'],
      }),
      request("User C waits independently", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/search", {
        auth: bearer("tokenC"),
        tests: [...common.okJSON, 'pm.test("waiting", () => pm.expect(pm.response.json().status).to.eql("waiting"));'],
      }),
      request("User C cancels waiting search", "POST", "{{matchmakingUrl}}/api/v1/matchmaking/cancel", {
        auth: bearer("tokenC"),
        tests: [...common.okJSON, 'pm.test("cancelled", () => pm.expect(pm.response.json().status).to.eql("cancelled"));'],
      }),
      request("Cancelled user returns to virtual waiting state", "GET", "{{matchmakingUrl}}/api/v1/matchmaking/status", {
        auth: bearer("tokenC"),
        tests: [...common.okJSON, 'pm.test("no stale game id", () => { const body = pm.response.json(); pm.expect(body.status).to.eql("waiting"); pm.expect(body).not.to.have.property("game_id"); });'],
      }),
    ]),
    folder("Game resource and history contracts", [
      request("Game invalid UUID", "GET", "{{gameUrl}}/api/v1/games/not-a-uuid", {
        tests: common.error(400, "invalid_game_id"),
      }),
      request("Game missing UUID", "GET", "{{gameUrl}}/api/v1/games/00000000-0000-4000-8000-000000000099", {
        tests: common.error(404, "game_not_found"),
      }),
      request("Matched game has valid initial state", "GET", "{{gameUrl}}/api/v1/games/{{gameId}}", {
        tests: [
          ...common.okJSON,
          'pm.test("players and state", () => { const body = pm.response.json(); pm.expect(body.id).to.eql(pm.environment.get("gameId")); pm.expect(body.white_player_id).to.eql(pm.environment.get("userAId")); pm.expect(body.black_player_id).to.eql(pm.environment.get("userBId")); pm.expect(body.status).to.eql("active"); pm.expect(body.current_turn).to.eql("white"); pm.expect(body.winner_id).to.eql(undefined); pm.expect(body.move_history).to.be.an("array").that.is.empty; pm.expect(body.legal_moves).to.be.an("array").that.is.not.empty; });',
          'pm.test("initial board has exactly 12 pieces per color on dark squares", () => { const cells = pm.response.json().board_state.board.cells; let white=0, black=0, invalid=0; cells.forEach((row,r) => row.forEach((piece,c) => { if (!piece) return; if ((r+c)%2 !== 1) invalid++; if (piece.color === "white") white++; if (piece.color === "black") black++; })); pm.expect(white).to.eql(12); pm.expect(black).to.eql(12); pm.expect(invalid).to.eql(0); });',
        ],
      }),
      request("Create game rejects unknown field", "POST", "{{gameUrl}}/api/v1/games", {
        headers: jsonHeaders,
        body: '{"white_player_id":"{{userAId}}","black_player_id":"{{userCId}}","unexpected":true}',
        tests: common.error(400, "invalid_request"),
      }),
      request("Create game rejects invalid white UUID", "POST", "{{gameUrl}}/api/v1/games", {
        headers: jsonHeaders,
        body: '{"white_player_id":"bad","black_player_id":"{{userCId}}"}',
        tests: common.error(400, "invalid_white_id"),
      }),
      request("Create game rejects same player", "POST", "{{gameUrl}}/api/v1/games", {
        headers: jsonHeaders,
        body: '{"white_player_id":"{{userAId}}","black_player_id":"{{userAId}}"}',
        tests: common.error(400, "invalid_request"),
      }),
      request("History rejects invalid token", "GET", "{{gameUrl}}/api/v1/users/{{userAId}}/games", {
        auth: { type: "bearer", bearer: [{ key: "token", value: "not-a-jwt", type: "string" }] },
        tests: common.error(401, "invalid_token"),
      }),
      request("Owner sees active game", "GET", "{{gameUrl}}/api/v1/users/{{userAId}}/games?limit=1", {
        auth: bearer("tokenA"),
        tests: [
          ...common.okJSON,
          'pm.test("active game visible to owner", () => { const body = pm.response.json(); pm.expect(body.items).to.have.lengthOf(1); pm.expect(body.items[0].game_id).to.eql(pm.environment.get("gameId")); pm.expect(body.items[0].status).to.eql("active"); pm.expect(body.items[0].user_color).to.eql("white"); });',
        ],
      }),
      request("Other user cannot enumerate active games", "GET", "{{gameUrl}}/api/v1/users/{{userAId}}/games", {
        auth: bearer("tokenC"),
        tests: [...common.okJSON, 'pm.test("active games hidden", () => pm.expect(pm.response.json().items).to.be.an("array").that.is.empty);'],
      }),
      request("History rejects malformed target UUID", "GET", "{{gameUrl}}/api/v1/users/not-a-uuid/games", {
        auth: bearer("tokenA"),
        tests: common.error(400, "invalid_request"),
      }),
      request("History rejects malformed cursor", "GET", "{{gameUrl}}/api/v1/users/{{userAId}}/games?cursor=-1", {
        auth: bearer("tokenA"),
        tests: common.error(400, "invalid_cursor"),
      }),
      request("History rejects non-numeric limit", "GET", "{{gameUrl}}/api/v1/users/{{userAId}}/games?limit=NaN", {
        auth: bearer("tokenA"),
        tests: common.error(400, "invalid_request"),
      }),
      request("Negative limit is normalized and bounded", "GET", "{{gameUrl}}/api/v1/users/{{userAId}}/games?limit=-10", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("service default applied", () => pm.expect(pm.response.json().items.length).to.be.at.most(20));'],
      }),
      request("Huge limit is capped", "GET", "{{gameUrl}}/api/v1/users/{{userAId}}/games?limit=999999", {
        auth: bearer("tokenA"),
        tests: [...common.okJSON, 'pm.test("hard cap applied", () => pm.expect(pm.response.json().items.length).to.be.at.most(50));'],
      }),
    ]),
  ],
};

for (const [filename, collection] of [
  ["identity-profile.postman_collection.json", contract],
  ["rating-matchmaking-game.postman_collection.json", workflows],
]) {
  fs.writeFileSync(path.join(root, filename), JSON.stringify(collection, null, 2) + "\n");
}
