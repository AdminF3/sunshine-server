-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migrate_baseline_years(baseTableName text) RETURNS VOID AS $$
DECLARE
  tableName text;
  y numeric;
BEGIN
  FOR j IN 0..2 LOOP
    tableName := baseTableName;
    y := j + 1;


    IF j > 0 THEN
      tableName := concat(baseTableName, '_', j::text);
    END IF;

    FOR i IN 0..11 LOOP
      UPDATE contracts SET tables = tables || jsonb_set(
        tables,
        concat('{', tableName, ',rows,', i::text, ',0}')::text[],
        concat('"', (i + 1)::text, '"')::jsonb,
        FALSE
      );
    END LOOP;

    UPDATE contracts SET tables = tables || jsonb_set(
      tables,
      concat('{', tableName, ',columns,0,name}')::text[],
      '"{\"en\": \"Month\", \"pl\": \"Miesiąc\", \"ro\": \"Luna\", \"au\": \"Monat\", \"lv\": \"Mēnesis1\", \"bg\": \"Месец\"}"'::jsonb,
      FALSE
    );

    IF baseTableName = 'baseyear_n' THEN
      UPDATE contracts SET tables = tables || jsonb_set(
        tables || jsonb_set(
          tables,
          concat('{', tableName, ',columns,2,headers,0}')::text[],
          '"$Q_{t}$"'::jsonb,
          FALSE
        ),
        concat('{', tableName, ',columns,1,headers,0}')::text[],
        '"$D_{Apk}$"'::jsonb,
        FALSE
      );
    ELSIF baseTableName = 'baseconditions_n' THEN
      UPDATE contracts SET tables = tables || jsonb_set(
        tables || jsonb_set(
          tables || jsonb_set(
            tables,
            concat('{', tableName, ',columns,3,headers,0}')::text[],
            '"$T_{3}$"'::jsonb,
            FALSE
          ),
          concat('{', tableName, ',columns,2,headers,0}')::text[],
          '"$T_{1}$"'::jsonb,
          FALSE
        ),
        concat('{', tableName, ',columns,1,headers,0}')::text[],
        '"$D_{Apk}$"'::jsonb,
        FALSE
      );
    END IF;
  END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migrate_baseline_years('baseyear_n');
SELECT migrate_baseline_years('baseconditions_n');

DROP FUNCTION migrate_baseline_years(baseTableName text);

-- +goose Down

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION migrate_down_baseline_years(baseTableName text) RETURNS VOID AS $$
DECLARE
  tableName text;
  m text;
BEGIN
  FOR j IN 0..2 LOOP
    SELECT baseTableName INTO tableName;

    IF j > 0 THEN
      SELECT baseTableName || '_' || j::text INTO tableName;
    END IF;

    FOR i IN 0..11 LOOP
      IF i = 0 THEN
        m = '"{\"en\": \"January\", \"pl\": \"Styczeń\", \"ro\": \"Ianuarie\", \"au\": \"Januar\", \"lv\": \"Janvāris\", \"bg\": \"Януари\"}"';
      ELSIF i = 1 THEN
        m = '"{\"en\": \"February\", \"pl\": \"Luty\", \"ro\": \"Februarie\", \"au\": \"Februar\", \"lv\": \"Februāris\", \"bg\": \"Февруару\"}"';
      ELSIF i = 2 THEN
        m = '"{\"en\": \"March\", \"pl\": \"Marzec\", \"ro\": \"Martie\", \"au\": \"März\", \"lv\": \"Marts\", \"bg\": \"Март\"}"';
      ELSIF i = 3 THEN
        m = '"{\"en\": \"April\", \"pl\": \"Kwiecień\", \"ro\": \"Aprilie\", \"au\": \"April\", \"lv\": \"Aprīlis\", \"bg\": \"Април\"}"';
      ELSIF i = 4 THEN
        m = '"{\"en\": \"May\", \"pl\": \"Maj\", \"ro\": \"Mai\", \"au\": \"Mai\", \"lv\": \"Maijs\", \"bg\": \"Май\"}"';
      ELSIF i = 5 THEN
        m = '"{\"en\": \"June\", \"pl\": \"Czerwiec\", \"ro\": \"Iunie\", \"au\": \"Juni\", \"lv\": \"Jūnijs\", \"bg\": \"Юни\"}"';
      ELSIF i = 6 THEN
        m = '"{\"en\": \"July\", \"pl\": \"Lipiec\", \"ro\": \"Iulie\", \"au\": \"Juli\", \"lv\": \"Jūlijs\", \"bg\": \"Юли\"}"';
      ELSIF i = 7 THEN
        m = '"{\"en\": \"August\", \"pl\": \"Sierpień\", \"ro\": \"August\", \"au\": \"August\", \"lv\":\"Augusts\", \"bg\": \"Август\"}"';
      ELSIF i = 8 THEN
        m = '"{\"en\": \"September\", \"pl\": \"Wrzesień\", \"ro\": \"Septembrie\", \"au\": \"September\", \"lv\":\"Septembris\", \"bg\": \"Септември\"}"';
      ELSIF i = 9 THEN
        m = '"{\"en\": \"October\", \"pl\": \"Październik\", \"ro\": \"Octombrie\", \"au\": \"Oktober\", \"lv\":\"Oktobris\", \"bg\": \"Октомври\"}"';
      ELSIF i = 10 THEN
        m = '"{\"en\": \"November\", \"pl\": \"Listopad\", \"ro\": \"Noiembrie\", \"au\": \"November\", \"lv\": \"Novembris\", \"bg\": \"Ноември\"}"';
      ELSIF i = 11 THEN
        m = '"{\"en\": \"December\", \"pl\": \"Grudzień\", \"ro\": \"Decembrie\", \"au\": \"Dezember\", \"lv\": \"Decembris\", \"bg\": \"Декември\"}"';
      END IF;
      UPDATE contracts SET tables = tables || (
        jsonb_set(
          tables,
          concat('{', tableName, ',rows,', i::text, ',0}')::text[],
          m::text::jsonb,
          FALSE
        )
      );
    END LOOP;

    UPDATE contracts SET tables = tables || jsonb_set(
      tables,
      concat('{', tableName, ',columns,0,name')::text[],
      '""'::jsonb,
      FALSE
    );

    IF baseTableName = 'baseyear_n' THEN
      UPDATE contracts SET tables = tables || jsonb_set(
        tables || jsonb_set(
          tables,
          concat('{', tableName, ',columns,2,headers,0}')::text[],
          '"Qt"'::jsonb,
          FALSE
        ),
        concat('{', tableName, ',columns,1,headers,0}')::text[],
        '"DApk"'::jsonb,
        FALSE
      );
    ELSIF baseTableName = 'baseconditions_n' THEN
      UPDATE contracts SET tables = tables || jsonb_set(
        tables || jsonb_set(
          tables || jsonb_set(
            tables,
            concat('{', tableName, ',columns,3,headers,0}')::text[],
            '"T3"'::jsonb,
            FALSE
          ),
          concat('{', tableName, ',columns,2,headers,0}')::text[],
          '"T1"'::jsonb,
          FALSE
        ),
        concat('{', tableName, ',columns,1,headers,0}')::text[],
        '"DApk"'::jsonb,
        FALSE
      );
    END IF;
  END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

SELECT migrate_down_baseline_years('baseyear_n');
SELECT migrate_down_baseline_years('baseconditions_n');

DROP FUNCTION migrate_down_baseline_years(baseTableName text);
