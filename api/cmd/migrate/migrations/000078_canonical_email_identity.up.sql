BEGIN;

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM users
		WHERE OCTET_LENGTH(email) <> CHAR_LENGTH(email)
	) THEN
		RAISE EXCEPTION 'cannot enable ASCII-only email identity while non-ASCII addresses exist'
			USING HINT = 'Run the documented read-only inventory and resolve each non-ASCII identity before retrying.';
	END IF;

	IF EXISTS (
		SELECT 1
		FROM (
			SELECT TRANSLATE(
				BTRIM(email, E' \t\n\r\f\v'),
				'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
				'abcdefghijklmnopqrstuvwxyz'
			) AS canonical_email
			FROM users
		) canonicalized
		WHERE OCTET_LENGTH(canonical_email) > 254
			OR OCTET_LENGTH(SPLIT_PART(canonical_email, '@', 1)) > 64
			OR canonical_email !~ '^[a-z0-9!#$%&''*+/=?^_`{|}~-]+(\.[a-z0-9!#$%&''*+/=?^_`{|}~-]+)*@([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$'
	) THEN
		RAISE EXCEPTION 'cannot enable canonical email identity while unsupported ASCII addresses exist'
			USING HINT = 'Run the documented read-only malformed-identity report and resolve each listed user ID before retrying.';
	END IF;

  IF EXISTS (
    SELECT 1
    FROM users
    GROUP BY TRANSLATE(
		BTRIM(email, E' \t\n\r\f\v'),
		'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
		'abcdefghijklmnopqrstuvwxyz'
	)
    HAVING COUNT(*) > 1
  ) THEN
    RAISE EXCEPTION 'cannot enforce canonical email uniqueness while duplicate identities exist'
      USING HINT = 'Run the documented read-only collision report and resolve each collision group before retrying.';
  END IF;
END
$$;

CREATE UNIQUE INDEX users_email_unique_canonical
  ON users (TRANSLATE(
	BTRIM(email, E' \t\n\r\f\v'),
	'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
	'abcdefghijklmnopqrstuvwxyz'
  ));

DROP INDEX users_email_unique_lower;

COMMIT;
