-- +goose Up
UPDATE contracts SET tables = tables || '{
       "summary":
{
           "rows": [
               [
                   "{\"en\": \"Energy\", \"pl\": \"Energy\", \"ro\": \"Energy\", \"au\": \"Energy\", \"lv\":\"Energy\", \"bg\": \"Energy\"}",
                   "0",
                   "0",
                   "0"
               ],
               [
                   "{\"en\": \"Renovation\", \"pl\": \"Renovation\", \"ro\": \"Renovation\", \"au\": \"Renovation\", \"lv\":\"Renovation\", \"bg\": \"Renovation\"}",
                   "0",
                   "0",
                   "0"
               ],
               [
                   "{\"en\": \"Operation and maintenance\", \"pl\": \"Operation and maintenance\", \"ro\": \"Operation and maintenance\", \"au\": \"Operation and maintenance\", \"lv\":\"Operation and maintenance\", \"bg\": \"Operation and maintenance\"}",
                   "0",
                   "0",
                   "0"
               ]
           ],
           "columns": [
               {
                   "kind": 1,
                   "name": "{\"en\": \"Fee\", \"pl\": \"Fee\", \"ro\": \"Fee\", \"au\": \"Fee\", \"lv\":\"Fee\", \"bg\": \"Fee\"}",
                   "headers": null
               },
               {
                   "kind": 3,
                   "name": "{\"en\": \"EUR/month\", \"pl\": \"EUR\"/miesiąc, \"ro\": \"EUR/lună\", \"au\": \"EUR/monat\", \"lv\":\"EUR/month\", \"bg\": \"EUR/month\"}",
                   "headers": null
               },
               {
                   "kind": 3,
                   "name": "{\"en\": \"VAT\", \"pl\": \"VAT\", \"ro\": \"VAT\", \"au\": \"VAT\", \"lv\":\"VAT\", \"bg\": \"VAT\"}",
                   "headers": null
               },
               {
                   "kind": 3,
                   "name": "{\"en\": \"Total\", \"pl\": \"Total\", \"ro\": \"Total\", \"au\": \"Total\", \"lv\":\"Total\", \"bg\": \"Total\"}",
                   "headers": null
               }
           ]
        }
}';
-- +goose Down
SELECT 1;
