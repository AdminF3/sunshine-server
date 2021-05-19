-- +goose Up
UPDATE contracts SET fields = fields || (
       SELECT jsonb_set(
		fields,
		'{"calculations_qietg"}',
		'""',
		true
       )
);

UPDATE contracts SET fields = fields || (
       SELECT jsonb_set(
		fields,
		'{"calculations_qapkczg"}',
		'""',
		true
       )
);

UPDATE contracts SET fields = fields || (
       SELECT jsonb_set(
		fields,
		'{"calculations_om1"}',
		'""',
		true
       )
);

-- +goose Down
UPDATE contracts SET fields = fields #- '{calculations_qietg}';
UPDATE contracts SET fields = fields #- '{calculations_qapkczg}';
UPDATE contracts SET fields = fields #- '{calculations_om1}';
