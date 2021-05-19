-- +goose Up
CREATE TABLE meeting_guests (
	id UUID PRIMARY KEY DEFAULT PUBLIC.gen_random_uuid(),

	meeting_id UUID NOT NULL,
	name TEXT NOT NULL,
        type INTEGER,
        email TEXT,
        phone TEXT,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	deleted_at TIMESTAMP WITH TIME ZONE
);

INSERT INTO meeting_guests (meeting_id, name, email, phone)
SELECT meetings.ID, meetings.guest, meetings.guest_email, meetings.guest_phone
FROM meetings;

ALTER TABLE meetings DROP COLUMN guest;
ALTER TABLE meetings DROP COLUMN guest_phone;
ALTER TABLE meetings DROP COLUMN guest_email;

-- +goose Down
ALTER TABLE meetings ADD COLUMN guest TEXT NOT NULL DEFAULT 'nan';
ALTER TABLE meetings ADD COLUMN guest_phone TEXT;
ALTER TABLE meetings ADD COLUMN guest_email TEXT;

UPDATE meetings
SET guest = gst.name,
guest_email = gst.email,
guest_phone = gst.phone
FROM (SELECT meeting_id, name, email, phone from meeting_guests) AS gst
WHERE meetings.ID = gst.meeting_id;

ALTER TABLE meetings ALTER COLUMN guest DROP DEFAULT;

DROP TABLE meeting_guests;
