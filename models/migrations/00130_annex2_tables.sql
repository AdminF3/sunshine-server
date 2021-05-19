-- +goose Up
UPDATE contracts SET tables = tables #- '{renovation_overall_budget, rows, 4}';
UPDATE contracts SET tables = tables #- '{renovation_financial_plan_a}';
UPDATE contracts SET tables = tables #- '{renovation_financial_plan_b}';
UPDATE contracts SET tables = tables #- '{renovation_financial_plan_c}';

UPDATE contracts SET tables = tables || '{"renovation_financial_plan": {
"rows": [
 [
             "{\"en\": \"State budget contribution – ALTUM\", \"pl\": \"State budget contribution – ALTUM\", \"ro\": \"State budget contribution – ALTUM\", \"au\": \"State budget contribution – ALTUM\", \"lv\":\"Grants\", \"bg\": \"State budget contribution – ALTUM\"}",
             "0"
         ],
         [
             "{\"en\": \"Municipal budget contribution\", \"pl\": \"Municipal budget contribution\", \"ro\": \"Municipal budget contribution\", \"au\": \"Municipal budget contribution\", \"lv\":\"Pašvaldības budžeta ieguldījums\", \"bg\": \"Municipal budget contribution\"}",
             "0"
         ],
         [
             "{\"en\": \"Client contribution\", \"pl\": \"Client contribution\", \"ro\": \"Client contribution\", \"au\": \"Client contribution\", \"lv\":\"Pasūtītāja ieguldījums\", \"bg\": \"Client contribution\"}",
             "0"
         ],
         [
             "{\"en\": \"Contractor Financial Contribution\", \"pl\": \"Contractor Financial Contribution\", \"ro\": \"Contractor Financial Contribution\", \"au\": \"Contractor Financial Contribution\", \"lv\":\"Izpildītāja finanšu ieguldījums\", \"bg\": \"Contractor Financial Contribution\"}",
             "0"
         ],
         [
             "{\"en\": \"Total costs for Renovation Works (including VAT)\", \"pl\": \"Total costs for Renovation Works (including VAT)\", \"ro\": \"Total costs for Renovation Works (including VAT)\", \"au\": \"Total costs for Renovation Works (including VAT)\", \"lv\":\"Atjaunošanas darbu kopējās izmaksas (ieskaitot PVN)\", \"bg\": \"Total costs for Renovation Works (including VAT)\"}",
             "0"
         ]
     ],
     "title": "",
     "columns": [
         {
             "kind": 1,
             "name": "{\"en\": \"Source of Funding\", \"pl\": \"Source of Funding\", \"ro\": \"Source of Funding\", \"au\": \"Source of Funding\", \"lv\":\"Finansējuma avots\", \"bg\": \"Source of Funding\"}",
             "headers": null
         },
         {
             "kind": 3,
             "name": "{\"en\": \"Costs\", \"pl\": \"Costs\", \"ro\": \"Costs\", \"au\": \"Costs\", \"lv\":\"izmaksas\", \"bg\": \"Costs\"}",
             "headers": null
         }
]
}
}';



-- +goose Down
UPDATE contracts SET tables = tables || '{
 "renovation_overall_budget": {
     "rows": [
         [
             "{\"en\": \"Project development and management costs\", \"pl\": \"Koszty opracowania projektu i zarządzania nim\", \"ro\": \"Costuri de dezvoltare și gestionare a proiectului\", \"au\": \"Projektentwicklungs- und Projektmanagementkosten\", \"lv\":\"Projekta izstrādes un vadības izmaksas\", \"bg\": \"Project development and management costs\"}",
             "0"
         ],
         [
             "{\"en\": \"Construction costs\", \"pl\": \"Koszty budowy i montażu\", \"ro\": \"Costuri de construcţie şi instalare\", \"au\": \"Bau- und Installationskosten\", \"lv\":\"Būvniecības un uzstādīšanas izmaksas\", \"bg\": \"Construction costs\"}",
             "0"
         ],
         [
             "{\"en\": \"Project supervision costs\", \"pl\": \"Koszty nadzoru nad projektem\", \"ro\": \"Costuri de  supraveghere a proiectului\", \"au\": \"Projektsteuerungskosten\", \"lv\":\"Projekta uzraudzības izmaksas\", \"bg\": \"Project supervision costs\"}",
             "0"
         ],
         [
             "{\"en\": \"Financial charges\", \"pl\": \"Koszty finansowe\", \"ro\": \"Cheltuieli financiare\", \"au\": \"Finanzierungskosten\", \"lv\":\"Finanšu maksājumi\", \"bg\": \"Financial charges\"}",
             "0"
         ],
         [
             "CONTRACTOR profit",
             "0"
         ]
     ],
     "title": "",
     "columns": [
         {
             "kind": 1,
             "name": "{\"en\": \"Budget of Renovation works\", \"pl\": \"Budżet Prac renowacyjnych\", \"ro\": \"Bugetul lucrărilor de renovare\", \"au\": \"Budget der Sanierungsarbeiten\", \"lv\":\"Atjaunošanas darbu budžets\", \"bg\": \"Budget of Renovation works\"}",
             "headers": null
         },
         {
             "kind": 3,
             "name": "{\"en\": \"Costs\", \"pl\": \"Costs\", \"ro\": \"Costs\", \"au\": \"Costs\", \"lv\":\"izmaksas\", \"bg\": \"Costs\"}",
             "headers": null
         }
     ]}
 }';

UPDATE contracts SET tables = tables || '{"renovation_financial_plan_a": {"rows": [], "title": "", "columns": []}}';
UPDATE contracts SET tables = tables || '{"renovation_financial_plan_b": {"rows": [], "title": "", "columns": []}}';
UPDATE contracts SET tables = tables || '{"renovation_financial_plan_c": {"rows": [], "title": "", "columns": []}}';
UPDATE contracts SET tables = tables #- '{renovation_financial_plan}';
