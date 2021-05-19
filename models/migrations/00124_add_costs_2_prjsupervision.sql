-- +goose Up
UPDATE contracts SET tables = tables || (
       SELECT jsonb_insert(
              tables,
              '{project_supervision,columns,1}'::text[],
              '{"kind": 3,"name": "{\"en\": \"Costs\", \"pl\": \"Costs\", \"ro\": \"Costs\", \"au\": \"Costs\", \"lv\":\"Costs\", \"bg\": \"Costs\"}","headers": null}',
              TRUE
       )
);

UPDATE contracts SET tables = tables || (
       SELECT jsonb_insert(
              tables,
              '{project_supervision,rows,1,1}'::text[],
              '""',
              TRUE
      )
);

UPDATE contracts SET tables = tables || (
       SELECT jsonb_insert(
              tables,
              '{project_supervision,rows,0,1}'::text[],
              '""',
              TRUE
      )
);


-- +goose Down
UPDATE contracts SET tables = tables
       #- '{project_supervision,columns,2}';

UPDATE contracts SET tables = tables
       #- '{project_supervision,rows,1,2}';

UPDATE contracts SET tables = tables
       #- '{project_supervision,rows,0,2}';
