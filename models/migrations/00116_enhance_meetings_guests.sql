-- +goose Up

ALTER TABLE meeting_guests ADD COLUMN organization TEXT;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migrate_meeting_participants() RETURNS VOID AS $$
DECLARE
	m_row record;
	mg_row record;
	part_i text;
	p_name text;
	p_email text;
	parts text[];

BEGIN
	FOR m_row IN
		SELECT * FROM meetings WHERE participant IS NOT NULL
	LOOP
		FOREACH part_i IN ARRAY m_row.participant
		LOOP
			parts=regexp_matches(part_i, '(.*)<(.*)>', 'g');
			p_name=parts[1];
			p_email=parts[2];

			INSERT INTO meeting_guests(meeting_id, name, email, type) SELECT m_row.id, p_name, p_email, 12
				WHERE NOT EXISTS (
					SELECT id FROM meeting_guests WHERE meeting_id=m_row.id AND email=p_email
				)
        AND p_name IS NOT NULL
			RETURNING * INTO mg_row;
		END LOOP;
	END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migrate_meeting_participants();
DROP FUNCTION migrate_meeting_participants;

ALTER TABLE meetings DROP COLUMN participant;

-- +goose Down
ALTER TABLE meeting_guests DROP COLUMN organization;
ALTER TABLE meetings ADD COLUMN participant TEXT[];
