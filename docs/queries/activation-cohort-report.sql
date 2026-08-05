\pset format aligned
\pset tuples_only off

WITH excluded_test_accounts AS (
  SELECT value::integer AS user_id
  FROM jsonb_array_elements_text(:'registry_json'::jsonb)
), eligible_users AS (
  SELECT u.id, u.created_at
  FROM users u
  LEFT JOIN excluded_test_accounts excluded ON excluded.user_id = u.id
  WHERE excluded.user_id IS NULL
), cohorts AS (
  SELECT 7 AS window_days
  UNION ALL SELECT 30
  UNION ALL SELECT 90
  UNION ALL SELECT 365
), cohort_users AS (
  SELECT
    cohorts.window_days,
    users.id,
    EXISTS (
      SELECT 1
      FROM wordlists
      WHERE wordlists.user_id = users.id
        AND wordlists.created_at >= users.created_at
    ) AS created_wordlist,
    EXISTS (
      SELECT 1
      FROM words
      WHERE words.user_id = users.id
        AND words.created_at >= users.created_at
    ) AS added_word,
    EXISTS (
      SELECT 1
      FROM quiz_performance
      WHERE quiz_performance.user_id = users.id
        AND quiz_performance.created_at >= users.created_at
    ) AS answered_quiz
  FROM cohorts
  LEFT JOIN eligible_users users
    ON users.created_at >= current_date - make_interval(days => cohorts.window_days)
)
SELECT
  window_days,
  count(id) AS signups,
  count(*) FILTER (WHERE created_wordlist) AS wordlist_creators,
  count(*) FILTER (WHERE added_word) AS word_adders,
  count(*) FILTER (WHERE answered_quiz) AS database_quiz_answerers,
  round(100.0 * count(*) FILTER (WHERE created_wordlist) / NULLIF(count(id), 0), 1)
    AS signup_to_wordlist_pct,
  round(100.0 * count(*) FILTER (WHERE added_word) / NULLIF(count(id), 0), 1)
    AS signup_to_word_pct,
  round(100.0 * count(*) FILTER (WHERE answered_quiz) / NULLIF(count(id), 0), 1)
    AS signup_to_database_quiz_answer_pct,
  'database_quiz_answerers is not quiz_session_completed' AS semantic_warning
FROM cohort_users
GROUP BY window_days
ORDER BY window_days;
