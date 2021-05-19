-- +goose Up

--project_development_renovations table
update contracts set tables = tables || '{
    "project_development_renovations": {
        "rows": [
            [
                "Energy audit",
                ""
            ],
            [
                "Civic engineering appraisal",
                ""
            ],
            [
                "Technical design for construction works",
                ""
            ],
            [
                "Technical design for heating, ventilation and domestic hot water systems",
                ""
            ],
            [
                "Project management",
                ""
            ],
            [
                "Preparation of grant application",
                ""
            ],
            [
                "Tendering of renovation works",
                ""
            ],
            [
                "Contracting and commissioning",
                ""
            ],
            [
                "Management and coordination",
                ""
            ]
        ],
        "columns": [
            {
                "kind": 0,
                "name": "Position",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Description",
                "headers": null
            }
        ]
    }
}';

-- construction_costs_renovations table
update contracts set tables = tables || '{
    "construction_costs_renovations": {
        "rows": [
            [
                "Energy audit",
                ""
            ],
            [
                "Energy audit",
                ""
            ],
            [
                "Thermal insulation of exterior walls",
                ""
            ],
            [
                "Thermal insulation of interior walls dividing different thermal zones",
                ""
            ],
            [
                "Windows",
                ""
            ],
            [
                "Windows indoor jambs and sills",
                ""
            ],
            [
                "Entrance doors",
                ""
            ],
            [
                "Doors indoor jambs and sills",
                ""
            ],
            [
                "Plinth thermal and hydro insulation",
                ""
            ],
            [
                "Thermal insulation of the attic",
                ""
            ],
            [
                "Thermal insulation of roofs",
                ""
            ],
            [
                "Thermal insulation of the basement sealing",
                ""
            ],
            [
                "Heating distribution system",
                ""
            ],
            [
                "Heat substation/supply",
                ""
            ],
            [
                "Domestic hot water system",
                ""
            ],
            [
                "Ventilation system",
                ""
            ],
            [
                "Roof structural repairs",
                ""
            ],
            [
                "Roof cover",
                ""
            ],
            [
                "Entrance roofs",
                ""
            ],
            [
                "Staircase roofs",
                ""
            ],
            [
                "Gutters and rainwater canalisation",
                ""
            ],
            [
                "Balcony structural repairs",
                ""
            ],
            [
                "Balcony railing / closing",
                ""
            ],
            [
                "Renovation of staircases",
                ""
            ],
            [
                "Cold water system",
                ""
            ],
            [
                "Electrical system",
                ""
            ],
            [
                "Construction site organisation and maintenance",
                ""
            ]
        ],
        "columns": [
            {
                "kind": 0,
                "name": "Position",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Description",
                "headers": null
            }
        ]
    }
}';


--project_supervision table
update contracts set tables = tables || '{
    "project_supervision": {
        "rows": [
            [
                "Building supervision",
                ""
            ],
            [
                "Author supervision",
                ""
            ]
        ],
        "columns": [
            {
                "kind": 0,
                "name": "Position",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Description",
                "headers": null
            }
        ]
    }
}';

--financial_charges table
update contracts set tables = tables || '{
    "financial_charges": {
        "rows": [
            [
                "Bank Fees",
                ""
            ],
            [
                "Forfaiting Fees",
                ""
            ]
        ],
        "columns": [
            {
                "kind": 0,
                "name": "Position",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Description",
                "headers": null
            }
        ]
    }
}';



-- +goose Down

update contracts set tables = tables - 'project_development_renovations';
update contracts set tables = tables - 'construction_costs_renovations';

