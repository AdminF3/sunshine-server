-- +goose Up

-- work phase scope renovation
UPDATE contracts SET tables = tables || '{
	"workPhase_scope_renovation": {
		"rows": [
      	    [ "Energy audit", "", "", "", "", "" ],
            [ "Civic engineering appraisal", "", "", "", "", "" ],
            [ "Technical design for construction works", "", "", "", "", "" ],
            [ "Technical design for heating, ventilation and domestic hot water systems", "", "", "", "", "" ],
            [ "Project management", "", "", "", "", "" ],
            [ "Preparation of grant application", "", "", "", "", "" ],
            [ "Tendering of renovation works", "", "", "", "", "" ],
            [ "Contracting and commissioning", "", "", "", "", "" ],
            [ "Management and coordination", "", "", "", "", "" ],
	        [ "Energy audit", "", "", "", "", "" ],
            [ "Thermal insulation of exterior walls", "", "", "", "", "" ],
            [ "Thermal insulation of interior walls dividing different thermal zones", "", "", "", "", "" ],
            [ "Windows", "", "", "", "", "" ],
            [ "Windows indoor jambs and sills", "", "", "", "", "" ],
            [ "Entrance doors", "", "", "", "", "" ],
            [ "Doors indoor jambs and sills", "", "", "", "", "" ],
            [ "Plinth thermal and hydro insulation", "", "", "", "", "" ],
            [ "Thermal insulation of the attic", "", "", "", "", "" ],
            [ "Thermal insulation of roofs", "", "", "", "", "" ],
            [ "Thermal insulation of the basement sealing", "", "", "", "", "" ],
            [ "Heating distribution system", "", "", "", "", "" ],
            [ "Heat substation/supply", "", "", "", "", "" ],
            [ "Domestic hot water system", "", "", "", "", "" ],
            [ "Ventilation system", "", "", "", "", "" ],
            [ "Roof structural repairs", "", "", "", "", "" ],
            [ "Roof cover", "", "", "", "", "" ],
            [ "Entrance roofs", "", "", "", "", "" ],
            [ "Staircase roofs", "", "", "", "", "" ],
            [ "Gutters and rainwater canalisation", "", "", "", "", "" ],
            [ "Balcony structural repairs", "", "", "", "", "" ],
            [ "Balcony railing / closing", "", "", "", "", "" ],
            [ "Renovation of staircases", "", "", "", "", "" ],
            [ "Cold water system", "", "", "", "", "" ],
            [ "Electrical system", "", "", "", "", "" ],
            [ "Construction site organisation and maintenance", "", "", "", "", "" ]
        ],
        "columns": [
            {
                "kind": 0,
                "name": "Position",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Planned Date",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Conclusion Date",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Responsible for execution",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Status",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Comments",
                "headers": null
            }
        ]
    }
}';

-- +goose Down

UPDATE contracts SET tables = tables || '{
	"workPhase_scope_renovation": {
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
            ],
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
                "name": "Conclusion Date",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Status",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Comments",
                "headers": null
            },
            {
                "kind": 0,
                "name": "Responsible for execution",
                "headers": null
            }
        ]
    }
}';
