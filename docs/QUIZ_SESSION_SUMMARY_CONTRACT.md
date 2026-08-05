# Quiz session summary contract

Decision: the quiz-answer command does not return a session summary. `PATCH /wordlists/:wordlistId/quizzes` remains a per-answer command with an empty `204 No Content` success response.

## Why the server does not own this summary

The API receives one answer containing a wordlist, word, definition, tracking row, quiz type, correctness, and response time. It does not receive a session identifier, session start, or completion signal. The mobile screen defines when a session begins and whether leaving it is completion or cancellation, so returning cumulative counts from each answer would invent a server session boundary, add an analytics/due-count read to every write, and still race another device.

The answer transaction remains authoritative only for that answer: it validates ownership, updates the Leitner row, records quiz analytics atomically, and commits before returning. Analytics-write or commit failure rolls back the Leitner transition and returns an error. A `204` therefore means that answer committed, but never represents a completed session.

## Client aggregation

One mounted quiz screen owns one process-local session ID and these values:

- `answeredCount`: answer, written-answer, or skip interactions accepted by the screen, counted once per displayed quiz.
- `correctCount`: the subset evaluated as correct by the same quiz payload.
- `durationMs`: elapsed client time from the first loaded quiz until explicit completion or unmount.
- `outcome`: `success` for explicit completion, `cancelled` for navigation/unmount before completion, and `failure` only for a future terminal session-level failure—not for an individual answer request.

The client emits the versioned `quiz_session_started`, `quiz_answered`, and `quiz_session_completed` analytics contract. These interaction counts are not represented as durable server totals and must not be shown as proof that every answer synced. An individual PATCH failure retains its own request error/retry semantics and must not be folded into a fabricated server summary.

## Progress and due data

Box transitions and durable totals come from explicit analytics/progress reads after the answer mutation invalidates their cache. The next-due count belongs to the progress-summary contract in the following roadmap item. The client must not infer it from reminder payloads, local answer counts, or the answer command.

The derived box-distribution snapshot and premium cache invalidation currently run after commit in detached work, so an immediate progress read may briefly return the previous snapshot. Completion UI must represent that state as refreshing/eventually consistent rather than rewriting the answer result or deriving a local box transition.

If a completion screen later needs box transitions, durable counts, or next-due data, it should issue one versioned summary/progress query when the session closes. Changing the answer PATCH from `204` requires a separately versioned API contract and coordinated mobile migration; clients must not conditionally parse a body from the current endpoint.

## Error semantics

- Invalid answer payload: `400` with an error body.
- Tracking row absent or not owned by the authenticated user: `404` with the stable `tracking not found` error.
- Storage/processing failure: `500` with an error body.
- Success: `204` with no response body.

The mobile loading/error-state roadmap item owns the visual retry treatment for per-answer failures. This contract does not silently turn those failures into a successful server session.
