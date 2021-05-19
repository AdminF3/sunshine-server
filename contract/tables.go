package contract

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Tables map[string]Table

func NewTables() Tables {
	var errs [26]error
	v := make(Tables)
	v["renovation_overall_budget"], errs[0] = NewTable(
		[]Column{
			Column{Name: `{"en": "Budget of Renovation works", "pl": "Budżet Prac renowacyjnych", "ro": "Bugetul lucrărilor de renovare", "au": "Budget der Sanierungsarbeiten", "lv":"Atjaunošanas darbu budžets", "bg": "Budget of Renovation works"}`, Kind: Name},
			Column{Name: `{"en": "Costs", "pl": "Costs", "ro": "Costs", "au": "Costs", "lv":"izmaksas", "bg": "Costs"}`, Kind: Money},
		},
		Row{`{"en": "Project development and management costs", "pl": "Koszty opracowania projektu i zarządzania nim", "ro": "Costuri de dezvoltare și gestionare a proiectului", "au": "Projektentwicklungs- und Projektmanagementkosten", "lv":"Projekta izstrādes un vadības izmaksas", "bg": "Project development and management costs"}`, "0"},
		Row{`{"en": "Construction costs", "pl": "Koszty budowy i montażu", "ro": "Costuri de construcţie şi instalare", "au": "Bau- und Installationskosten", "lv":"Būvniecības un uzstādīšanas izmaksas", "bg": "Construction costs"}`, "0"},
		Row{`{"en": "Project supervision costs", "pl": "Koszty nadzoru nad projektem", "ro": "Costuri de  supraveghere a proiectului", "au": "Projektsteuerungskosten", "lv":"Projekta uzraudzības izmaksas", "bg": "Project supervision costs"}`, "0"},
		Row{`{"en": "Financial charges", "pl": "Koszty finansowe", "ro": "Cheltuieli financiare", "au": "Finanzierungskosten", "lv":"Finanšu maksājumi", "bg": "Financial charges"}`, "0"},
	)

	v["renovation_financial_plan"], errs[1] = NewTable(
		[]Column{
			Column{Name: `{"en": "Source of Funding", "pl": "Source of Funding", "ro": "Source of Funding", "au": "Source of Funding", "lv":"Finansējuma avots", "bg": "Source of Funding"}`, Kind: Name},
			Column{Name: `{"en": "Costs", "pl": "Costs", "ro": "Costs", "au": "Costs", "lv":"izmaksas", "bg": "Costs"}`, Kind: Money},
		},
		Row{`{"en": "State budget contribution – ALTUM", "pl": "State budget contribution – ALTUM", "ro": "State budget contribution – ALTUM", "au": "State budget contribution – ALTUM", "lv":"Grants", "bg": "State budget contribution – ALTUM"}`, "0"},
		Row{`{"en": "Municipal budget contribution", "pl": "Municipal budget contribution", "ro": "Municipal budget contribution", "au": "Municipal budget contribution", "lv":"Pašvaldības budžeta ieguldījums", "bg": "Municipal budget contribution"}`, "0"},
		Row{`{"en": "Client contribution", "pl": "Client contribution", "ro": "Client contribution", "au": "Client contribution", "lv":"Pasūtītāja ieguldījums", "bg": "Client contribution"}`, "0"},
		Row{`{"en": "Contractor Financial Contribution", "pl": "Contractor Financial Contribution", "ro": "Contractor Financial Contribution", "au": "Contractor Financial Contribution", "lv":"Izpildītāja finanšu ieguldījums", "bg": "Contractor Financial Contribution"}`, "0"},
		Row{`{"en": "Total costs for Renovation Works (including VAT)", "pl": "Total costs for Renovation Works (including VAT)", "ro": "Total costs for Renovation Works (including VAT)", "au": "Total costs for Renovation Works (including VAT)", "lv":"Atjaunošanas darbu kopējās izmaksas (ieskaitot PVN)", "bg": "Total costs for Renovation Works (including VAT)"}`, "0"},
	)

	financialChargesColumns := []Column{
		Column{Name: `{"en": "Position", "pl": "Position", "ro": "Position", "au": "Position", "lv":"Position", "bg": "Position"}`, Kind: String},
		Column{Name: `{"en": "Description", "pl": "Description", "ro": "Description", "au": "Description", "lv":"Description", "bg": "Description"}`, Kind: String},
		Column{Name: `{"en": "Costs", "pl": "Costs", "ro": "Costs", "au": "Costs", "lv":"izmaksas", "bg": "Costs"}`, Kind: Money},
	}
	financialChargesRows := []Row{
		Row{`{"en": "Bank Fees", "pl": "Bank Fees", "ro": "Bank Fees", "au": "Bank Fees", "lv":"Bank Fees", "bg": "Bank Fees"}`, "", "0"},
		Row{`{"en": "Forfaiting Fees", "pl": "Forfaiting Fees", "ro": "Forfaiting Fees", "au": "Forfaiting Fees", "lv":"Forfaiting Fees", "bg": "Forfaiting Fees"}`, "", "0"},
	}

	v["financial_charges"], errs[2] = NewTable(financialChargesColumns, financialChargesRows...)

	summaryColumns := []Column{
		{Name: `{"en": "Fee", "pl": "Fee", "ro": "Fee", "au": "Fee", "lv":"Fee", "bg": "Fee"}`, Kind: Name},
		{Name: `{"en": "EUR/month", "pl": "EUR/miesiąc", "ro": "EUR/lună", "au": "EUR/monat", "lv":"EUR/month", "bg": "EUR/month"}`, Kind: Money},
		{Name: `{"en": "VAT", "pl": "VAT", "ro": "VAT", "au": "VAT", "lv":"VAT", "bg": "VAT"}`, Kind: Money},
		{Name: `{"en": "Total", "pl": "Total", "ro": "Total", "au": "Total", "lv":"Kopā", "bg": "Total"}`, Kind: Money},
	}
	summaryRows := []Row{
		Row{`{"en": "Energy", "pl": "Energy", "ro": "Energy", "au": "Energy", "lv":"Energy", "bg": "Energy"}`, "0", "0", "0"},
		Row{`{"en": "Renovation", "pl": "Renovation", "ro": "Renovation", "au": "Renovation", "lv":"Atjaunošana", "bg": "Renovation"}`, "0", "0", "0"},
		Row{`{"en": "Operation and maintenance", "pl": "Operation and maintenance", "ro": "Operation and maintenance", "au": "Operation and maintenance", "lv":"Ekspluatācija un apkope", "bg": "Operation and maintenance"}`, "0", "0", "0"},
	}

	v["summary"], errs[3] = NewTable(summaryColumns, summaryRows...)

	//Annex 4 table baseyears
	baseyearColumns := []Column{
		Column{Name: "", Kind: Name, Headers: []string{`{"en": "Symbol", "pl": "Oznaczenie", "ro": "Simbol", "au": "Symbol", "lv":"Simbol", "bg": "Символ"}`, `{"en": "Unit", "pl": "Jednostka", "ro": "Unitate", "au": "Einheit", "lv":"Vienība", "bg": "Единици"}`}},
		Column{Name: `{"en": "Heating days", "pl": "Liczba dni ogrzewania", "ro": "Numărul zilelor de încălzire efectivă", "au": "Anzahl Heiztage", "lv":"Apkures dienu skaits", "bg": "Брой на дни"}`, Kind: Count, Headers: []string{"$D_{Apk}$", `{"en": "Days", "pl": "Dni", "ro": "Zile", "au": "Tage", "lv":"Dienas", "bg": "Дни"}`}},
		Column{Name: `{"en": "Total heat energy consumption", "pl": "Całkowite zużycie energii cieplnej", "ro": "Consumul total e energie termică", "au": "Gesamter Wärmeenergieverbrauch", "lv":"Kopējais siltumenerģijas patēriņš", "bg": "Total heat energy consumption"}`, Kind: Energy, Headers: []string{"$Q_{t}$", "MWh"}},
		Column{Name: `{"en": "Domestic hot water consumption", "pl": "Zużycie ciepłej wody użytkowej", "ro": "Consumul de apă caldă menajeră", "au": "Warmwasserverbrauch", "lv":"Mājsaimniecību karstā ūdens patēriņš", "bg": "Domestic hot water consumption"}`, Kind: Volume, Headers: []string{"V", "m³"}},
		Column{Name: `{"en": "Domestic hot water temperature", "pl": "Temperatura ciepłej wody użytkowej", "ro": "Temperatura apei calde menajere", "au": "Warmwasse rtemperatur", "lv":"Mājsaimniecību karstā ūdens  temperatūra", "bg": "Domestic hot water temperature"}`, Kind: Temperature, Headers: []string{"0ku", "°C"}},
	}
	baseyearRows := monthsRows(12)

	v["baseyear_n_2"], errs[4] = NewTable(baseyearColumns, baseyearRows...)
	v["baseyear_n_1"], errs[5] = NewTable(baseyearColumns, baseyearRows...)
	v["baseyear_n"], errs[6] = NewTable(baseyearColumns, baseyearRows...)

	baselineColumns := []Column{
		Column{Name: "", Kind: Name},
		Column{Name: `{"en": "Symbol", "pl": "Oznaczenie", "ro": "Simbol", "au": "Symbol", "lv":"Simbol", "bg": "Символ"}`, Kind: Name},
		Column{Name: `{"en": "Unit", "pl": "Jednostka", "ro": "Unitate", "au": "Einheit", "lv":"Vienība", "bg": "Единици"}`, Kind: Name},
		Column{Name: "$baseyear^{n-2}$", Kind: Decimal},
		Column{Name: "$baseyear^{n-1}$", Kind: Decimal},
		Column{Name: "$baseyear^{n}$", Kind: Decimal},
		Column{Name: "Reference", Kind: Decimal},
	}

	baselineRows := []Row{
		Row{`{"en": "Total heat energy consumption", "pl": "Całkowite zużycie energii cieplnej", "ro": "Consumul total e energie termică", "au": "Gesamter Wärmeenergieverbrauch", "lv":"Kopējais siltumenerģijas patēriņš", "bg": "Total heat energy consumption"}`, "$Q_{T,ref}$", "MWh/year", "", "", "", ""},
		Row{`{"en": "Space heating consumption", "pl": "Zużycie na ogrzewanie pomieszczeń", "ro": "Consumul de încălzire în încăpere", "au": "Energieverbrauch für Raumwärme", "lv":"Telpu apkures patēriņš", "bg": "Space heating consumption"}`, "$Q_{Apk,ref}$", "MWh/year", "", "", "", ""},
		Row{`{"en": "Circulation losses", "pl": "Straty obiegowe", "ro": "Pierderi de circulaţie", "au": "Zirkulationsverluste", "lv":"Cirkulācijas zudumi", "bg": "Circulation losses"}`, "$Q_{cz,ref}$", "MWh/year", "", "", "", ""},
		Row{`{"en": "Domestic hot water consumption", "pl": "Zużycie ciepłej wody użytkowej", "ro": "Consumul de apă caldă menajeră", "au": "Warmwasserverbrauch", "lv":"Mājsaimniecību karstā ūdens patēriņš", "bg": "Domestic hot water consumption"}`, "$Q_{ku,ref}$", "MWh/year", "", "", "", ""},
		Row{`{"en": "Energy consumption for space heating and circulation losses", "pl": "Zużycie energii na ogrzewanie pomieszczeń i na straty obiegowe", "ro": "Consumul de energie pentru încălzirea spaţiului şi pierderile de circulaţie", "au": "Energieverbrauch für Raumwärme und Zirkulationsverluste", "lv":"Enerģijas patēriņš telpu apkurei un cirkulācijas zudumi", "bg": "Energy consumption forspace heating and circulation losses"}`, "$Q_{Apk,cz,ref}$", "MWh/year", "", "", "", ""},
		Row{`{"en": "Average indoor temperature", "pl": "Średnia temperatura wewnątrz pomieszczeń", "ro": "Temperatura medie interioară", "au": "Durchschnittliche Raumtemperatur", "lv":"Vidējā gaisa temperatūra telpās", "bg": "Average indoor temperature"}`, "$T_{1,ref}$", "${℃}$", "", "", "", ""},
		Row{`{"en": "Degree days", "pl": "Stopniodni", "ro": "Degree Days", "au": "Heizgradtage", "lv":"Grādu dienas", "bg": "Degree days"}`, "$GDD_{ref}$", "{-}", "", "", "", ""},
	}
	v["baseline"], errs[7] = NewTable(baselineColumns, baselineRows...)

	maintananceHeaderColumns := []Column{
		Column{Name: `{"en": "Maintenance activity", "pl": "Maintenance activity", "ro": "Maintenance activity", "au": "Maintenance activity", "lv":"Apkopes darbība", "bg": "Maintenance activity"}`, Kind: Name},
		Column{Name: `{"en": "Minimum Frequency", "pl": "Minimum Frequency", "ro": "Minimum Frequency", "au": "Minimum Frequency", "lv":"Minimālais periodiskums", "bg": "Minimum Frequency"}`, Kind: String},

		Column{Name: `{"en": "Responsible party", "pl": "Responsible party", "ro": "Responsible party", "au": "Responsible party", "lv":"Responsible party", "bg": "Responsible party"}`, Kind: String},
		Column{Name: `{"en": "Material and spare parts", "pl": "Material and spare parts", "ro": "Material and spare parts", "au": "Material and spare parts", "lv":"Materiāli un rezerves daļas", "bg": "Material and spare parts"}`, Kind: String},
		Column{Name: `{"en": "Tools and equipment", "pl": "Tools and equipment", "ro": "Tools and equipment", "au": "Tools and equipment", "lv":"Rīki un aprīkojums", "bg": "Tools and equipment"}`, Kind: String},
	}

	periodicMaintananceRows := []Row{
		Row{`{
"en": "Periodic visual inspection of the buildings and installed systems/equipment",
"pl": "Periodic visual inspection of the buildings and installed systems/equipment",
"ro": "Periodic visual inspection of the buildings and installed systems/equipment",
"au": "Periodic visual inspection of the buildings and installed systems/equipment",
"lv": "Periodiska Ēkas un uzstādīto sistēmu/aprīkojuma vizuāla apsekošana",
"bg": "Periodic visual inspection of the buildings and installed systems/equipment"}`, "Monthly", "Contractor", "", ""},
		Row{`{
"en": "Operation of water softening system / filters for domestic hot water preparation",
"pl": "Operation of water softening system / filters for domestic hot water preparation",
"ro": "Operation of water softening system / filters for domestic hot water preparation",
"au": "Operation of water softening system / filters for domestic hot water preparation",
"lv": "Ūdens mīkstināšanas sistēmas / filtru ekspluatācija mājsaimniecību karstā ūdens sagatavošana",
"bg": "Operation of water softening system / filters for domestic hot water preparation"}`, "Monthly", "Contractor", "", ""},
		Row{`{
"en": "Replacement of Air handling units filters, control and cleaning of ventilation equipment",
"pl": "Replacement of Air handling units filters, control and cleaning of ventilation equipment",
"ro": "Replacement of Air handling units filters, control and cleaning of ventilation equipment",
"au": "Replacement of Air handling units filters, control and cleaning of ventilation equipment",
"lv": "Gaisa apmaiņas ierīču filtru nomaiņa, ventilācijas aprīkojuma kontrole un tīrīšana",
"bg": "Replacement of Air handling units filters, control and cleaning of ventilation equipment"}`, "3 months", "Contractor", "", ""},
		Row{`{
"en": "Start the heating system",
"pl": "Start the heating system",
"ro": "Start the heating system",
"au": "Start the heating system",
"lv": "Apkures sistēmas palaišana",
"bg": "Start the heating system"}`, "Annual", "Contractor", "", ""},
		Row{`{
"en": "Stop the heating system",
"pl": "Stop the heating system",
"ro": "Stop the heating system",
"au": "Stop the heating system",
"lv": "Apkures sistēmas atslēgšana",
"bg": "Stop the heating system"}`, "Annual", "Contractor", "", ""},
		Row{`{
"en": "Periodic cleaning of heat exchangers",
"pl": "Periodic cleaning of heat exchangers",
"ro": "Periodic cleaning of heat exchangers",
"au": "Periodic cleaning of heat exchangers",
"lv": "Siltummaiņu periodiska tīrīšana",
"bg": "Periodic cleaning of heat exchangers"}`, "Annual in summer", "Contractor", "", ""},
		Row{`{
"en": "Periodic cleaning of pump filters (replace if needed)",
"pl": "Periodic cleaning of pump filters (replace if needed)",
"ro": "Periodic cleaning of pump filters (replace if needed)",
"au": "Periodic cleaning of pump filters (replace if needed)",
"lv": "Pumpju filtru periodiska tīrīšana (nomaiņa nepieciešamības gadījumā)",
"bg": "Periodic cleaning of pump filters (replace if needed)"}`, "Annual", "Contractor", "", ""},
		Row{`{
"en": "Periodic inspection of water softening system",
"pl": "Periodic inspection of water softening system",
"ro": "Periodic inspection of water softening system",
"au": "Periodic inspection of water softening system",
"lv": "Ūdens mīkstināšanas sistēmas periodiska apsekošana",
"bg": "Periodic inspection of water softening system"}`, "Annual", "Contractor", "", ""},
		Row{`{
"en": "Control/cleaning air vents",
"pl": "Control/cleaning air vents",
"ro": "Control/cleaning air vents",
"au": "Control/cleaning air vents",
"lv": "Gaisa ventilācijas eju kontrole/tīrīšana",
"bg": "Control/cleaning air vents"}`, "Annual", "Client under Contractor instruction", "", ""},
	}
	v["periodic_maint_activities_covered_by_contractor"], errs[8] = NewTable(
		maintananceHeaderColumns,
		periodicMaintananceRows...,
	)

	midTermMaintananceRows := []Row{
		Row{`{
"en": "Check radiator thermostatic valves (replace if needed)",
"pl": "Check radiator thermostatic valves (replace if needed)",
"ro": "Check radiator thermostatic valves (replace if needed)",
"au": "Check radiator thermostatic valves (replace if needed)",
"lv": "Pārbaudīt radiatoru termostatu vārstus (nepieciešamības gadījumā nomainīt)",
"bg": "Check radiator thermostatic valves (replace if needed)"}`, "5 years", "Contractor", "", ""},
		Row{`{
"en": "Check control vents (replace if needed)",
"pl": "Check control vents (replace if needed)",
"ro": "Check control vents (replace if needed)",
"au": "Check control vents (replace if needed)",
"lv": "Pārbaudīt kontrolvārstus (nepieciešamības gadījumā nomainīt)",
"bg": "Check control vents (replace if needed)"}`, "5 years", "Contractor", "", ""},
		Row{`{
"en": "Check space heating balancing valves (replace if needed)",
"pl": "Check space heating balancing valves (replace if needed)",
"ro": "Check space heating balancing valves (replace if needed)",
"au": "Check space heating balancing valves (replace if needed)",
"lv": "Pārbaudīt telpu apkures balansēšanas vārstus (nepieciešamības gadījumā nomainīt)",
"bg": "Check space heating balancing valves (replace if needed)"}`, "5 years", "Contractor", "", ""},
		Row{`{
"en": "Cleaning ventilation shafts",
"pl": "Cleaning ventilation shafts",
"ro": "Cleaning ventilation shafts",
"au": "Cleaning ventilation shafts",
"lv": "Ventilācijas šahtu tīrīšana",
"bg": "Cleaning ventilation shafts"}`, "5 years", "Contractor", "", ""},
		Row{`{
"en": "Replacing circulation pumps for domestic hot water system",
"pl": "Replacing circulation pumps for domestic hot water system",
"ro": "Replacing circulation pumps for domestic hot water system",
"au": "Replacing circulation pumps for domestic hot water system",
"lv": "Cirkulācijas pumpju nomaiņa mājsaimniecību karstā ūdens sistēmā",
"bg": "Replacing circulation pumps for domestic hot water system"}`, "7 years", "Contractor", "", ""},
		Row{`{
"en": "Replacing circulation pumps for space heating system",
"pl": "Replacing circulation pumps for space heating system",
"ro": "Replacing circulation pumps for space heating system",
"au": "Replacing circulation pumps for space heating system",
"lv": "Cirkulācijas pumpju nomaiņa telpu apkures sistēmā",
"bg": "Replacing circulation pumps for space heating system"}`, "7 years", "Contractor", "", ""},
	}
	v["mid_term_preventative_activity"], errs[9] = NewTable(
		maintananceHeaderColumns,
		midTermMaintananceRows...,
	)

	longTermMaintananceRows := []Row{
		Row{`{
"en": "Overhauls/replacement of air handling units",
"pl": "Overhauls/replacement of air handling units",
"ro": "Overhauls/replacement of air handling units",
"au": "Overhauls/replacement of air handling units",
"lv": "Gaisa apmaiņas ierīču kapitālais remonts/nomaiņa",
"bg": "Overhauls/replacement of air handling units"}`, "10 years", "Contractor", "", ""},
		Row{`{
"en": "Replace radiator thermostatic valves",
"pl": "Replace radiator thermostatic valves",
"ro": "Replace radiator thermostatic valves",
"au": "Replace radiator thermostatic valves",
"lv": "Radiatoru termostatu vārstu nomaiņa",
"bg": "Replace radiator thermostatic valves"}`, "15 years", "Contractor", "", ""},
	}

	v["long_term_provisioned_activities"], errs[10] = NewTable(
		maintananceHeaderColumns,
		longTermMaintananceRows...,
	)

	recommendedMaintananceRows := []Row{
		Row{`{
"en": "Repainting/cleaning/repairs of the staircases",
"pl": "Repainting/cleaning/repairs of the staircases",
"ro": "Repainting/cleaning/repairs of the staircases",
"au": "Repainting/cleaning/repairs of the staircases",
"lv": "Kāpņutelpu krāsas atjaunošana / uzkopšana / remonts",
"bg": "Repainting/cleaning/repairs of the staircases"}`, "10 years", "Client", "", ""},
		Row{`{
"en": "Repainting/cleaning/repairs of the entrances",
"pl": "Repainting/cleaning/repairs of the entrances",
"ro": "Repainting/cleaning/repairs of the entrances",
"au": "Repainting/cleaning/repairs of the entrances",
"lv": "Ieeju krāsas atjaunošana/ uzkopšana / remonts",
"bg": "Repainting/cleaning/repairs of the entrances"}`, "10 years", "Client", "", ""},
		Row{`{
"en": "Repainting/cleaning/repairs of the façade",
"pl": "Repainting/cleaning/repairs of the façade",
"ro": "Repainting/cleaning/repairs of the façade",
"au": "Repainting/cleaning/repairs of the façade",
"lv": "Fasādes krāsas atjaunošana / uzkopšana / remonts",
"bg": "Repainting/cleaning/repairs of the façade"}`, "10 years", "Client", "", ""},
		Row{`{
"en": "Repainting/cleaning/repairs of the plinth",
"pl": "Repainting/cleaning/repairs of the plinth",
"ro": "Repainting/cleaning/repairs of the plinth",
"au": "Repainting/cleaning/repairs of the plinth",
"lv": "Cokola krāsas atjaunošana / tīrīšana / remonts",
"bg": "Repainting/cleaning/repairs of the plinth"}`, "15 years", "Client", "", ""},
	}
	v["reccomended_maintanance_activity"], errs[11] = NewTable(
		maintananceHeaderColumns,
		recommendedMaintananceRows...,
	)

	calcEnergyColumns := []Column{
		Column{Name: "Months of Contract", Kind: Name, Headers: []string{`{"en": "Unit", "pl": "Jednostka", "ro": "Unitate", "au": "Einheit", "lv":"Vienība", "bg": "Единици"}`}},
		Column{Name: "$Q^{m}_{Apk,cz,G}$", Kind: Energy, Headers: []string{"MWh", "A"}},
		Column{Name: "$HT^m$", Kind: Money, Headers: []string{"EUR/MWh", "B"}},
		Column{Name: "$E^{m}_{F,G}$", Kind: Money, Headers: []string{"EUR", "C=AxB"}},
		Column{Name: "$A_{Apk}$", Kind: Area, Headers: []string{"$m^2$", "D"}},
		Column{Name: "$Ap^m$", Kind: Money, Headers: []string{"EUR/$m^2$ month", "E=C/D"}},
	}
	calcEnergyRows := []Row{
		Row{`{"en": "Month 1", "pl": "Miesiąc 1", "ro": "Luna 1", "au": "Monat 1", "lv":"Mēnesis 1", "bg": "Месец 1"}`, "", "", "", "", ""},
		Row{`{"en": "Month 2", "pl": "Miesiąc 2", "ro": "Luna 2", "au": "Monat 2", "lv":"Mēnesis 2", "bg": "Месец 2"}`, "", "", "", "", ""},
		Row{`{"en": "Month 2", "pl": "Miesiąc 2", "ro": "Luna 2", "au": "Monat 2", "lv":"Mēnesis 2", "bg": "Месец 2"}`, "", "", "", "", ""},
		Row{`{"en": "Month 3", "pl": "Miesiąc 3", "ro": "Luna 3", "au": "Monat 3", "lv":"Mēnesis 3", "bg": "Месец 3"}`, "", "", "", "", ""},
		Row{`{"en": "Month 4", "pl": "Miesiąc 4", "ro": "Luna 4", "au": "Monat 4", "lv":"Mēnesis 4", "bg": "Месец 4"}`, "", "", "", "", ""},
		Row{`{"en": "Month 5", "pl": "Miesiąc 5", "ro": "Luna 5", "au": "Monat 5", "lv":"Mēnesis 5", "bg": "Месец 5"}`, "", "", "", "", ""},
		Row{`{"en": "Month 6", "pl": "Miesiąc 6", "ro": "Luna 6", "au": "Monat 6", "lv":"Mēnesis 6", "bg": "Месец 6"}`, "", "", "", "", ""},
		Row{`{"en": "Month 7", "pl": "Miesiąc 7", "ro": "Luna 7", "au": "Monat 7", "lv":"Mēnesis 7", "bg": "Месец 7"}`, "", "", "", "", ""},
		Row{`{"en": "Month 8", "pl": "Miesiąc 8", "ro": "Luna 8", "au": "Monat 8", "lv":"Mēnesis 8", "bg": "Месец 8"}`, "", "", "", "", ""},
		Row{`{"en": "Month 9", "pl": "Miesiąc 9", "ro": "Luna 9", "au": "Monat 9", "lv":"Mēnesis 9", "bg": "Месец 9"}`, "", "", "", "", ""},
		Row{`{"en": "Month 10", "pl": "Miesiąc 10", "ro": "Luna 10", "au": "Monat 10", "lv":"Mēnesis 10", "bg": "Месец 10"}`, "", "", "", "", ""},
		Row{`{"en": "Month 11", "pl": "Miesiąc 11", "ro": "Luna 11", "au": "Monat 11", "lv":"Mēnesis 11", "bg": "Месец 11"}`, "", "", "", "", ""},
		Row{`{"en": "Month 12", "pl": "Miesiąc 12", "ro": "Luna 12", "au": "Monat 12", "lv":"Mēnesis 12", "bg": "Месец 12"}`, "", "", "", "", ""},
	}
	v["calc_energy_fee"], errs[12] = NewTable(
		calcEnergyColumns,
		calcEnergyRows...,
	)

	settlementColumns := []Column{
		Column{Name: `{"en": "Settlement period", "pl": "Settlement period", "ro": "Settlement period", "au": "Settlement period", "lv":"Norēķinu periods", "bg": "Settlement period"}`, Kind: Name, Headers: []string{`{"en": "Unit", "pl": "Jednostka", "ro": "Unitate", "au": "Einheit", "lv":"Vienība", "bg": "Единици"}`}},
		Column{Name: "$Q^{m}_{Apk,cz,G}$", Kind: Energy, Headers: []string{"MWh", "A"}},
		Column{Name: "$ET^m$", Kind: Money, Headers: []string{"EUR/MWh", "B"}},
		Column{Name: "$E^{m}_{F,G}$", Kind: Money, Headers: []string{"EUR", "AxB"}},
		Column{Name: "$A_{Apk}$", Kind: Area, Headers: []string{"㎡", "D"}},
		Column{Name: "$Q^{m}_{Apk,cz,S}$", Kind: Energy, Headers: []string{"MWh", "F"}},
		Column{Name: "$E^{m}_{F,S}$", Kind: Money, Headers: []string{"EUR/month", "FxB"}},
	}
	settlementRows := []Row{
		Row{`{"en": "Month 1", "pl": "Miesiąc 1", "ro": "Luna 1", "au": "Monat 1", "lv":"Mēnesis 1", "bg": "Месец 1"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 2", "pl": "Miesiąc 2", "ro": "Luna 2", "au": "Monat 2", "lv":"Mēnesis 2", "bg": "Месец 2"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 3", "pl": "Miesiąc 3", "ro": "Luna 3", "au": "Monat 3", "lv":"Mēnesis 3", "bg": "Месец 3"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 4", "pl": "Miesiąc 4", "ro": "Luna 4", "au": "Monat 4", "lv":"Mēnesis 4", "bg": "Месец 4"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 5", "pl": "Miesiąc 5", "ro": "Luna 5", "au": "Monat 5", "lv":"Mēnesis 5", "bg": "Месец 5"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 6", "pl": "Miesiąc 6", "ro": "Luna 6", "au": "Monat 6", "lv":"Mēnesis 6", "bg": "Месец 6"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 7", "pl": "Miesiąc 7", "ro": "Luna 7", "au": "Monat 7", "lv":"Mēnesis 7", "bg": "Месец 7"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 8", "pl": "Miesiąc 8", "ro": "Luna 8", "au": "Monat 8", "lv":"Mēnesis 8", "bg": "Месец 8"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 9", "pl": "Miesiąc 9", "ro": "Luna 9", "au": "Monat 9", "lv":"Mēnesis 9", "bg": "Месец 9"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 10", "pl": "Miesiąc 10", "ro": "Luna 10", "au": "Monat 10", "lv":"Mēnesis 10", "bg": "Месец 10"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 11", "pl": "Miesiąc 11", "ro": "Luna 11", "au": "Monat 11", "lv":"Mēnesis 11", "bg": "Месец 11"}`, "", "", "", "", "", ""},
		Row{`{"en": "Month 12", "pl": "Miesiąc 12", "ro": "Luna 12", "au": "Monat 12", "lv":"Mēnesis 12", "bg": "Месец 12"}`, "", "", "", "", "", ""},
	}

	v["balancing_period_fee"], errs[13] = NewTable(
		settlementColumns,
		settlementRows...,
	)

	v["operation_maintenance_budget"], errs[14] = NewTable(
		[]Column{
			Column{Name: `{"en": "Operation and maintenance", "pl": "Operation and maintenance", "ro": "Operation and maintenance", "au": "Operation and maintenance", "lv":"Ekspluatācija un apkope", "bg": "Operation and maintenance"}`, Kind: Name},
			Column{Name: `{"en": "Costs", "pl": "Costs", "ro": "Costs", "au": "Costs", "lv":"izmaksas", "bg": "Costs"}`, Kind: Money},
		},
		Row{`{
"en": "Periodic maintenance activities",
"pl": "Periodic maintenance activities",
"ro": "Periodic maintenance activities",
"au": "Periodic maintenance activities",
"lv": "Periodiska apkope",
"bg": "Periodic maintenance activities"}`, "0"},
		Row{`{
"en": "Mid-term preventative maintenance",
"pl": "Mid-term preventative maintenance",
"ro": "Mid-term preventative maintenance",
"au": "Mid-term preventative maintenance",
"lv": "Starpposma profilaktiskā apkope",
"bg": "Mid-term preventative maintenance"}`, "0"},
		Row{`{
"en": "Long term provisioned maintenance",
"pl": "Long term provisioned maintenance",
"ro": "Long term provisioned maintenance",
"au": "Long term provisioned maintenance",
"lv": "Ilgtermiņā nodrošinātā apkope",
"bg": "Long term provisioned maintenance"}`, "0"},
		Row{`{
"en": "Operating costs",
"pl": "Operating costs",
"ro": "Operating costs",
"au": "Operating costs",
"lv": "Ekspluatācijas izmaksas",
"bg": "Operating costs"}`, "0"},
	)

	operationMaintenanceFeeColumns := []Column{
		Column{Name: `{
"en": "Year of Contract",
"pl": "Year of Contract",
"ro": "Year of Contract",
"au": "Year of Contract",
"lv": "Pakalpojumu sniegšanas perioda gad",
"bg": "Year of Contract"}`, Kind: Name},
		Column{Name: `{
"en": "Inflation index",
"pl": "Inflation index",
"ro": "Inflation index",
"au": "Inflation index",
"lv": "CPI",
"bg": "Inflation index"}`, Kind: Percent, Headers: []string{`\%`, `\%`}},
		Column{Name: `{
"en": "Annual Operational and Maintenance Fee",
"pl": "Annual Operational and Maintenance Fee",
"ro": "Annual Operational and Maintenance Fee",
"au": "Annual Operational and Maintenance Fee",
"lv": "Gada ekspluatācijas un apkopes maksa",
"bg": "Annual Operational and Maintenance Fee"}`, Kind: Money, Headers: []string{"EUR/year", "$OM_y$"}},
		Column{Name: `{
"en": "Monthly Operational and Maintenance Fee",
"pl": "Monthly Operational and Maintenance Fee",
"ro": "Monthly Operational and Maintenance Fee",
"au": "Monthly Operational and Maintenance Fee",
"lv": "Ikmēneša ekspluatācijas un apkopes maksa",
"bg": "Monthly Operational and Maintenance Fee"}`, Kind: Money, Headers: []string{"EUR/month", "$OM_y/12$"}},
		Column{Name: "$A_{Apk}$", Kind: Area, Headers: []string{"$m^2$", "D"}},
		Column{Name: `{
"en": "Monthly Operational and Maintenance Fee",
"pl": "Monthly Operational and Maintenance Fee",
"ro": "Monthly Operational and Maintenance Fee",
"au": "Monthly Operational and Maintenance Fee",
"lv": "Ikmēneša ekspluatācijas un apkopes maksa",
"bg": "Monthly Operational and Maintenance Fee"}`, Kind: Money, Headers: []string{"EUR/$m^2$ month", "$OM_m/D$"}},
	}
	v["operations_maintenance_fee"], errs[15] = NewTable(
		operationMaintenanceFeeColumns,
		Row{"y=1", "", "", "", "", ""},
		Row{"y=2", "", "", "", "", ""},
		Row{"y=3", "", "", "", "", ""},
		Row{"y=4", "", "", "", "", ""},
		Row{"y=5", "", "", "", "", ""},
		Row{"y=6", "", "", "", "", ""},
		Row{"y=7", "", "", "", "", ""},
		Row{"y=8", "", "", "", "", ""},
		Row{"y=9", "", "", "", "", ""},
		Row{"y=10", "", "", "", "", ""},
		Row{"y=11", "", "", "", "", ""},
		Row{"y=12", "", "", "", "", ""},
		Row{"y=13", "", "", "", "", ""},
		Row{"y=14", "", "", "", "", ""},
		Row{"y=15", "", "", "", "", ""},
		Row{"y=16", "", "", "", "", ""},
		Row{"y=17", "", "", "", "", ""},
		Row{"y=18", "", "", "", "", ""},
		Row{"y=19", "", "", "", "", ""},
		Row{"y=20", "", "", "", "", ""},
	)

	maintenanceLogColumns := []Column{
		Column{Name: "Activity", Kind: String},
		Column{Name: "Company responsible for execution", Kind: String},
		Column{Name: "Planned date", Kind: String},
		Column{Name: "Done date", Kind: String},
		Column{Name: "Status", Kind: String},
		Column{Name: "Comments", Kind: String},
	}
	v["maitenance_log"], errs[16] = NewTable(
		maintenanceLogColumns,
		Row{"a=1", "", "", "", "", ""},
	)
	//Create monitoring phase table
	v["monitoring_phase_table"], errs[17] = newMPTable()

	//Annex 4 table baseline-Conditions
	baseConditionsColumns := []Column{
		Column{Name: "", Kind: Name, Headers: []string{`{"en": "Symbol", "pl": "Oznaczenie", "ro": "Simbol", "au": "Symbol", "lv":"Simbol", "bg": "Символ"}`, `{"en": "Unit", "pl": "Jednostka", "ro": "Unitate", "au": "Einheit", "lv":"Vienība", "bg": "Единици"}`}},
		Column{Name: `{"en": "Heating days", "pl": "Liczba dni ogrzewania", "ro": "Numărul zilelor de încălzire efectivă", "au": "Anzahl Heiztage", "lv":"Apkures dienu skaits", "bg": "Брой на дни"}`, Kind: Count, Headers: []string{"$D_{Apk}$", `{"en": "Days", "pl": "Dni", "ro": "Zile", "au": "Tage", "lv":"Dienas", "bg": "Дни"}`}},
		Column{Name: `{"en": "Outdoor temperature", "pl": "Temperatura zewnętrzna", "ro": "Temperatură exterioară", "au": "Außentemperatur", "lv":"Ārējā gaisa temperatūra", "bg": "Средна температура на външния въздух"}`, Kind: Temperature, Headers: []string{"$T_{1}$", "°C"}},
		Column{Name: `{"en": "Average indoor temperature", "pl": "Średnia temperatura zewnętrzna", "ro": "Temperatura medie interioară", "au": "Durchschnittliche Raumtemperatur", "lv":"Vidējā gaisa temperatūra telpās", "bg": "Средна обемна температура на помещенията"}`, Kind: Temperature, Headers: []string{"$T_{3}$", "°C"}},
		Column{Name: `{"en": "Degree Days", "pl": "Stopniodni", "ro": "Numărul zilelor ce necestită încălzire", "au": "Heizgradtage", "lv":"Grādu dienas", "bg": "Денградуси"}`, Kind: Count, Headers: []string{"GDD", "-"}},
	}
	baseConditionsRows := monthsRows(12)

	v["baseconditions_n_2"], errs[18] = NewTable(baseConditionsColumns, baseConditionsRows...)
	v["baseconditions_n_1"], errs[19] = NewTable(baseConditionsColumns, baseConditionsRows...)
	v["baseconditions_n"], errs[20] = NewTable(baseConditionsColumns, baseConditionsRows...)

	//Createproject  measurements  table
	v["project_measurements_table"], errs[21] = newMeasurementsTable()

	projectDevelopmentColumns := []Column{
		Column{Name: `{"en": "Position", "pl": "Position", "ro": "Position", "au": "Position", "lv":"Position", "bg": "Position"}`, Kind: String},
		Column{Name: `{"en": "Description", "pl": "Description", "ro": "Description", "au": "Description", "lv":"Description", "bg": "Description"}`, Kind: String},
		Column{Name: `{"en": "Costs", "pl": "Costs", "ro": "Costs", "au": "Costs", "lv":"izmaksas", "bg": "Costs"}`, Kind: Money},
	}
	projectDevelopmentRows := []Row{
		Row{`{"en": "Energy audit", "pl": "Energy audit", "ro": "Energy audit", "au": "Energy audit", "lv":"Energy audit", "bg": "Energy audit"}`, "", "0"},
		Row{`{"en": "Civic engineering appraisal", "pl": "Civic engineering appraisal", "ro": "Civic engineering appraisal", "au": "Civic engineering appraisal", "lv":"Civic engineering appraisal", "bg": "Civic engineering appraisal"}`, "", "0"},
		Row{`{"en": "Technical design for construction works", "pl": "Technical design for construction works", "ro": "Technical design for construction works", "au": "Technical design for construction works", "lv":"Technical design for construction works", "bg": "Technical design for construction works"}`, "", "0"},
		Row{`{"en": "Technical design for heating, ventilation and domestic hot water systems", "pl": "Technical design for heating, ventilation and domestic hot water systems", "ro": "Technical design for heating, ventilation and domestic hot water systems", "au": "Technical design for heating, ventilation and domestic hot water systems", "lv":"Technical design for heating, ventilation and domestic hot water systems", "bg": "Technical design for heating, ventilation and domestic hot water systems"}`, "", "0"},
		Row{`{"en": "Project management", "pl": "Project management", "ro": "Project management", "au": "Project management", "lv":"Project management", "bg": "Project management"}`, "", "0"},
		Row{`{"en": "Preparation of grant application", "pl": "Preparation of grant application", "ro": "Preparation of grant application", "au": "Preparation of grant application", "lv":"Preparation of grant application", "bg": "Preparation of grant application"}`, "", "0"},
		Row{`{"en": "Tendering of renovation works", "pl": "Tendering of renovation works", "ro": "Tendering of renovation works", "au": "Tendering of renovation works", "lv":"Tendering of renovation works", "bg": "Tendering of renovation works"}`, "", "0"},
		Row{`{"en": "Contracting and commissioning", "pl": "Contracting and commissioning", "ro": "Contracting and commissioning", "au": "Contracting and commissioning", "lv":"Contracting and commissioning", "bg": "Contracting and commissioning"}`, "", "0"},
		Row{`{"en": "Management and coordination", "pl": "Management and coordination", "ro": "Management and coordination", "au": "Management and coordination", "lv":"Management and coordination", "bg": "Management and coordination"}`, "", "0"},
	}

	v["project_development_renovations"], errs[22] = NewTable(projectDevelopmentColumns, projectDevelopmentRows...)

	constructionCostsColumns := []Column{
		Column{Name: `{"en": "Position", "pl": "Position", "ro": "Position", "au": "Position", "lv":"Position", "bg": "Position"}`, Kind: String},
		Column{Name: `{"en": "Description", "pl": "Description", "ro": "Description", "au": "Description", "lv":"Description", "bg": "Description"}`, Kind: String},
		Column{Name: `{"en": "Costs", "pl": "Costs", "ro": "Costs", "au": "Costs", "lv":"izmaksas", "bg": "Costs"}`, Kind: Money},
	}
	constructionCostsRows := []Row{
		Row{`{"en": "Thermal insulation of exterior walls", "pl": "Thermal insulation of exterior walls", "ro": "Thermal insulation of exterior walls", "au": "Thermal insulation of exterior walls", "lv":"Thermal insulation of exterior walls", "bg": "Thermal insulation of exterior walls"}`, "", "0"},
		Row{`{"en": "Thermal insulation of interior walls dividing different thermal zones", "pl": "Thermal insulation of interior walls dividing different thermal zones", "ro": "Thermal insulation of interior walls dividing different thermal zones", "au": "Thermal insulation of interior walls dividing different thermal zones", "lv":"Thermal insulation of interior walls dividing different thermal zones", "bg": "Thermal insulation of interior walls dividing different thermal zones"}`, "", "0"},
		Row{`{"en": "Windows", "pl": "Windows", "ro": "Windows", "au": "Windows", "lv":"Windows", "bg": "Windows"}`, "", "0"},
		Row{`{"en": "Windows indoor jambs and sills", "pl": "Windows indoor jambs and sills", "ro": "Windows indoor jambs and sills", "au": "Windows indoor jambs and sills", "lv":"Windows indoor jambs and sills", "bg": "Windows indoor jambs and sills"}`, "", "0"},
		Row{`{"en": "Entrance doors", "pl": "Entrance doors", "ro": "Entrance doors", "au": "Entrance doors", "lv":"Entrance doors", "bg": "Entrance doors"}`, "", "0"},
		Row{`{"en": "Doors indoor jambs and sills", "pl": "Doors indoor jambs and sills", "ro": "Doors indoor jambs and sills", "au": "Doors indoor jambs and sills", "lv":"Doors indoor jambs and sills", "bg": "Doors indoor jambs and sills"}`, "", "0"},
		Row{`{"en": "Plinth thermal and hydro insulation", "pl": "Plinth thermal and hydro insulation", "ro": "Plinth thermal and hydro insulation", "au": "Plinth thermal and hydro insulation", "lv":"Plinth thermal and hydro insulation", "bg": "Plinth thermal and hydro insulation"}`, "", "0"},
		Row{`{"en": "Thermal insulation of the attic", "pl": "Thermal insulation of the attic", "ro": "Thermal insulation of the attic", "au": "Thermal insulation of the attic", "lv":"Thermal insulation of the attic", "bg": "Thermal insulation of the attic"}`, "", "0"},
		Row{`{"en": "Thermal insulation of roofs", "pl": "Thermal insulation of roofs", "ro": "Thermal insulation of roofs", "au": "Thermal insulation of roofs", "lv":"Thermal insulation of roofs", "bg": "Thermal insulation of roofs"}`, "", "0"},
		Row{`{"en": "Thermal insulation of the basement sealing", "pl": "Thermal insulation of the basement sealing", "ro": "Thermal insulation of the basement sealing", "au": "Thermal insulation of the basement sealing", "lv":"Thermal insulation of the basement sealing", "bg": "Thermal insulation of the basement sealing"}`, "", "0"},
		Row{`{"en": "Heating distribution system", "pl": "Heating distribution system", "ro": "Heating distribution system", "au": "Heating distribution system", "lv":"Heating distribution system", "bg": "Heating distribution system"}`, "", "0"},
		Row{`{"en": "Heat substation/supply", "pl": "Heat substation/supply", "ro": "Heat substation/supply", "au": "Heat substation/supply", "lv":"Heat substation/supply", "bg": "Heat substation/supply"}`, "", "0"},
		Row{`{"en": "Domestic hot water system", "pl": "Domestic hot water system", "ro": "Domestic hot water system", "au": "Domestic hot water system", "lv":"Domestic hot water system", "bg": "Domestic hot water system"}`, "", "0"},
		Row{`{"en": "Ventilation system", "pl": "Ventilation system", "ro": "Ventilation system", "au": "Ventilation system", "lv":"Ventilation system", "bg": "Ventilation system"}`, "", "0"},
		Row{`{"en": "Roof structural repairs", "pl": "Roof structural repairs", "ro": "Roof structural repairs", "au": "Roof structural repairs", "lv":"Roof structural repairs", "bg": "Roof structural repairs"}`, "", "0"},
		Row{`{"en": "Roof cover", "pl": "Roof cover", "ro": "Roof cover", "au": "Roof cover", "lv":"Roof cover", "bg": "Roof cover"}`, "", "0"},
		Row{`{"en": "Entrance roofs", "pl": "Entrance roofs", "ro": "Entrance roofs", "au": "Entrance roofs", "lv":"Entrance roofs", "bg": "Entrance roofs"}`, "", "0"},
		Row{`{"en": "Staircase roofs", "pl": "Staircase roofs", "ro": "Staircase roofs", "au": "Staircase roofs", "lv":"Staircase roofs", "bg": "Staircase roofs"}`, "", "0"},
		Row{`{"en": "Gutters and rainwater canalisation", "pl": "Gutters and rainwater canalisation", "ro": "Gutters and rainwater canalisation", "au": "Gutters and rainwater canalisation", "lv":"Gutters and rainwater canalisation", "bg": "Gutters and rainwater canalisation"}`, "", "0"},
		Row{`{"en": "Balcony structural repairs", "pl": "Balcony structural repairs", "ro": "Balcony structural repairs", "au": "Balcony structural repairs", "lv":"Balcony structural repairs", "bg": "Balcony structural repairs"}`, "", "0"},
		Row{`{"en": "Balcony railing / closing", "pl": "Balcony railing / closing", "ro": "Balcony railing / closing", "au": "Balcony railing / closing", "lv":"Balcony railing / closing", "bg": "Balcony railing / closing"}`, "", "0"},
		Row{`{"en": "Renovation of staircases", "pl": "Renovation of staircases", "ro": "Renovation of staircases", "au": "Renovation of staircases", "lv":"Renovation of staircases", "bg": "Renovation of staircases"}`, "", "0"},
		Row{`{"en": "Cold water system", "pl": "Cold water system", "ro": "Cold water system", "au": "Cold water system", "lv":"Cold water system", "bg": "Cold water system"}`, "", "0"},
		Row{`{"en": "Electrical system", "pl": "Electrical system", "ro": "Electrical system", "au": "Electrical system", "lv":"Electrical system", "bg": "Electrical system"}`, "", "0"},
		Row{`{"en": "Construction site organisation and maintenance", "pl": "Construction site organisation and maintenance", "ro": "Construction site organisation and maintenance", "au": "Construction site organisation and maintenance", "lv":"Construction site organisation and maintenance", "bg": "Construction site organisation and maintenance"}`, "", "0"},
	}

	v["construction_costs_renovations"], errs[23] = NewTable(constructionCostsColumns, constructionCostsRows...)

	projectSupervisionColumns := []Column{
		Column{Name: `{"en": "Position", "pl": "Position", "ro": "Position", "au": "Position", "lv":"Position", "bg": "Position"}`, Kind: String},
		Column{Name: `{"en": "Description", "pl": "Description", "ro": "Description", "au": "Description", "lv":"Description", "bg": "Description"}`, Kind: String},
		Column{Name: `{"en": "Costs", "pl": "Costs", "ro": "Costs", "au": "Costs", "lv":"izmaksas", "bg": "Costs"}`, Kind: Money},
	}
	projectSupervisionRows := []Row{
		Row{`{"en": "Building supervision", "pl": "Building supervision", "ro": "Building supervision", "au": "Building supervision", "lv":"Building supervision", "bg": "Building supervision"}`, "", ""},
		Row{`{"en": "Author supervision", "pl": "Author supervision", "ro": "Author supervision", "au": "Author supervision", "lv":"Author supervision", "bg": "Author supervision"}`, "", ""},
	}

	v["project_supervision"], errs[24] = NewTable(projectSupervisionColumns, projectSupervisionRows...)

	workPhaseScopeOfRenovationsColumns := []Column{
		Column{Name: `{"en": "Position", "pl": "Position", "ro": "Position", "au": "Position", "lv":"Position", "bg": "Position"}`, Kind: String},
		Column{Name: `{"en": "Planned Date", "pl": "Planned Date", "ro": "Planned Date", "au": "Planned Date", "lv": "Planned Date", "bg": "Planned Date"}`, Kind: String},
		Column{Name: `{"en": "Conclusion Date", "pl": "Conclusion Date", "ro": "Conclusion Date", "au": "Conclusion Date", "lv": "Conclusion Date", "bg": "Conclusion Date"}`, Kind: String},
		Column{Name: `{"en": "Responsible for execution", "pl": "Responsible for execution", "ro": "Responsible for execution", "au": "Responsible for execution", "lv": "Responsible for execution", "bg": "Responsible for execution"}`, Kind: String},
		Column{Name: `{"en": "Status", "pl": "Status", "ro": "Status", "au": "Status", "lv": "Status", "bg": "Status"}`, Kind: String},
		Column{Name: `{"en": "Comments", "pl": "Comments", "ro": "Comments", "au": "Comments", "lv": "Comments", "bg": "Comments"}`, Kind: String},
	}
	workPhaseScopeOfRenovationsRows := []Row{
		//annex2 2.1
		Row{`{"en": "Energy audit", "pl": "Energy audit", "ro": "Energy audit", "au": "Energy audit", "lv":"Energy audit", "bg": "Energy audit"}`, "", "", "", "", ""},
		Row{`{"en": "Civic engineering appraisal", "pl": "Civic engineering appraisal", "ro": "Civic engineering appraisal", "au": "Civic engineering appraisal", "lv":"Civic engineering appraisal", "bg": "Civic engineering appraisal"}`, "", "", "", "", ""},
		Row{`{"en": "Technical design for construction works", "pl": "Technical design for construction works", "ro": "Technical design for construction works", "au": "Technical design for construction works", "lv":"Technical design for construction works", "bg": "Technical design for construction works"}`, "", "", "", "", ""},
		Row{`{"en": "Technical design for heating, ventilation and domestic hot water systems", "pl": "Technical design for heating, ventilation and domestic hot water systems", "ro": "Technical design for heating, ventilation and domestic hot water systems", "au": "Technical design for heating, ventilation and domestic hot water systems", "lv":"Technical design for heating, ventilation and domestic hot water systems", "bg": "Technical design for heating, ventilation and domestic hot water systems"}`, "", "", "", "", ""},
		Row{`{"en": "Project management", "pl": "Project management", "ro": "Project management", "au": "Project management", "lv":"Project management", "bg": "Project management"}`, "", "", "", "", ""},
		Row{`{"en": "Preparation of grant application", "pl": "Preparation of grant application", "ro": "Preparation of grant application", "au": "Preparation of grant application", "lv":"Preparation of grant application", "bg": "Preparation of grant application"}`, "", "", "", "", ""},
		Row{`{"en": "Tendering of renovation works", "pl": "Tendering of renovation works", "ro": "Tendering of renovation works", "au": "Tendering of renovation works", "lv":"Tendering of renovation works", "bg": "Tendering of renovation works"}`, "", "", "", "", ""},
		Row{`{"en": "Contracting and commissioning", "pl": "Contracting and commissioning", "ro": "Contracting and commissioning", "au": "Contracting and commissioning", "lv":"Contracting and commissioning", "bg": "Contracting and commissioning"}`, "", "", "", "", ""},
		Row{`{"en": "Management and coordination", "pl": "Management and coordination", "ro": "Management and coordination", "au": "Management and coordination", "lv":"Management and coordination", "bg": "Management and coordination"}`, "", "", "", "", ""},
		//annex2 2.2
		Row{`{"en": "Energy audit", "pl": "Energy audit", "ro": "Energy audit", "au": "Energy audit", "lv":"Energy audit", "bg": "Energy audit"}`, "", "", "", "", ""},
		Row{`{"en": "Thermal insulation of exterior walls", "pl": "Thermal insulation of exterior walls", "ro": "Thermal insulation of exterior walls", "au": "Thermal insulation of exterior walls", "lv":"Thermal insulation of exterior walls", "bg": "Thermal insulation of exterior walls"}`, "", "", "", "", ""},
		Row{`{"en": "Thermal insulation of interior walls dividing different thermal zones", "pl": "Thermal insulation of interior walls dividing different thermal zones", "ro": "Thermal insulation of interior walls dividing different thermal zones", "au": "Thermal insulation of interior walls dividing different thermal zones", "lv":"Thermal insulation of interior walls dividing different thermal zones", "bg": "Thermal insulation of interior walls dividing different thermal zones"}`, "", "", "", "", ""},
		Row{`{"en": "Windows", "pl": "Windows", "ro": "Windows", "au": "Windows", "lv":"Windows", "bg": "Windows"}`, "", "", "", "", ""},
		Row{`{"en": "Windows indoor jambs and sills", "pl": "Windows indoor jambs and sills", "ro": "Windows indoor jambs and sills", "au": "Windows indoor jambs and sills", "lv":"Windows indoor jambs and sills", "bg": "Windows indoor jambs and sills"}`, "", "", "", "", ""},
		Row{`{"en": "Entrance doors", "pl": "Entrance doors", "ro": "Entrance doors", "au": "Entrance doors", "lv":"Entrance doors", "bg": "Entrance doors"}`, "", "", "", "", ""},
		Row{`{"en": "Doors indoor jambs and sills", "pl": "Doors indoor jambs and sills", "ro": "Doors indoor jambs and sills", "au": "Doors indoor jambs and sills", "lv":"Doors indoor jambs and sills", "bg": "Doors indoor jambs and sills"}`, "", "", "", "", ""},
		Row{`{"en": "Plinth thermal and hydro insulation", "pl": "Plinth thermal and hydro insulation", "ro": "Plinth thermal and hydro insulation", "au": "Plinth thermal and hydro insulation", "lv":"Plinth thermal and hydro insulation", "bg": "Plinth thermal and hydro insulation"}`, "", "", "", "", ""},
		Row{`{"en": "Thermal insulation of the attic", "pl": "Thermal insulation of the attic", "ro": "Thermal insulation of the attic", "au": "Thermal insulation of the attic", "lv":"Thermal insulation of the attic", "bg": "Thermal insulation of the attic"}`, "", "", "", "", ""},
		Row{`{"en": "Thermal insulation of roofs", "pl": "Thermal insulation of roofs", "ro": "Thermal insulation of roofs", "au": "Thermal insulation of roofs", "lv":"Thermal insulation of roofs", "bg": "Thermal insulation of roofs"}`, "", "", "", "", ""},
		Row{`{"en": "Thermal insulation of the basement sealing", "pl": "Thermal insulation of the basement sealing", "ro": "Thermal insulation of the basement sealing", "au": "Thermal insulation of the basement sealing", "lv":"Thermal insulation of the basement sealing", "bg": "Thermal insulation of the basement sealing"}`, "", "", "", "", ""},
		Row{`{"en": "Heating distribution system", "pl": "Heating distribution system", "ro": "Heating distribution system", "au": "Heating distribution system", "lv":"Heating distribution system", "bg": "Heating distribution system"}`, "", "", "", "", ""},
		Row{`{"en": "Heat substation/supply", "pl": "Heat substation/supply", "ro": "Heat substation/supply", "au": "Heat substation/supply", "lv":"Heat substation/supply", "bg": "Heat substation/supply"}`, "", "", "", "", ""},
		Row{`{"en": "Domestic hot water system", "pl": "Domestic hot water system", "ro": "Domestic hot water system", "au": "Domestic hot water system", "lv":"Domestic hot water system", "bg": "Domestic hot water system"}`, "", "", "", "", ""},
		Row{`{"en": "Ventilation system", "pl": "Ventilation system", "ro": "Ventilation system", "au": "Ventilation system", "lv":"Ventilation system", "bg": "Ventilation system"}`, "", "", "", "", ""},
		Row{`{"en": "Roof structural repairs", "pl": "Roof structural repairs", "ro": "Roof structural repairs", "au": "Roof structural repairs", "lv":"Roof structural repairs", "bg": "Roof structural repairs"}`, "", "", "", "", ""},
		Row{`{"en": "Roof cover", "pl": "Roof cover", "ro": "Roof cover", "au": "Roof cover", "lv":"Roof cover", "bg": "Roof cover"}`, "", "", "", "", ""},
		Row{`{"en": "Entrance roofs", "pl": "Entrance roofs", "ro": "Entrance roofs", "au": "Entrance roofs", "lv":"Entrance roofs", "bg": "Entrance roofs"}`, "", "", "", "", ""},
		Row{`{"en": "Staircase roofs", "pl": "Staircase roofs", "ro": "Staircase roofs", "au": "Staircase roofs", "lv":"Staircase roofs", "bg": "Staircase roofs"}`, "", "", "", "", ""},
		Row{`{"en": "Gutters and rainwater canalisation", "pl": "Gutters and rainwater canalisation", "ro": "Gutters and rainwater canalisation", "au": "Gutters and rainwater canalisation", "lv":"Gutters and rainwater canalisation", "bg": "Gutters and rainwater canalisation"}`, "", "", "", "", ""},
		Row{`{"en": "Balcony structural repairs", "pl": "Balcony structural repairs", "ro": "Balcony structural repairs", "au": "Balcony structural repairs", "lv":"Balcony structural repairs", "bg": "Balcony structural repairs"}`, "", "", "", "", ""},
		Row{`{"en": "Balcony railing / closing", "pl": "Balcony railing / closing", "ro": "Balcony railing / closing", "au": "Balcony railing / closing", "lv":"Balcony railing / closing", "bg": "Balcony railing / closing"}`, "", "", "", "", ""},
		Row{`{"en": "Renovation of staircases", "pl": "Renovation of staircases", "ro": "Renovation of staircases", "au": "Renovation of staircases", "lv":"Renovation of staircases", "bg": "Renovation of staircases"}`, "", "", "", "", ""},
		Row{`{"en": "Cold water system", "pl": "Cold water system", "ro": "Cold water system", "au": "Cold water system", "lv":"Cold water system", "bg": "Cold water system"}`, "", "", "", "", ""},
		Row{`{"en": "Electrical system", "pl": "Electrical system", "ro": "Electrical system", "au": "Electrical system", "lv":"Electrical system", "bg": "Electrical system"}`, "", "", "", "", ""},
		Row{`{"en": "Construction site organisation and maintenance", "pl": "Construction site organisation and maintenance", "ro": "Construction site organisation and maintenance", "au": "Construction site organisation and maintenance", "lv":"Construction site organisation and maintenance", "bg": "Construction site organisation and maintenance"}`, "", "", "", "", ""},
	}

	v["workPhase_scope_renovation"], errs[25] = NewTable(workPhaseScopeOfRenovationsColumns, workPhaseScopeOfRenovationsRows...)

	for i, err := range errs {
		if err != nil {
			panic(fmt.Sprintf("fail on table %d: %v", i, err))
		}
	}
	return v
}

func (t Tables) Value() (driver.Value, error) {
	if len(t) == 0 {
		return nil, nil
	}
	return json.Marshal(t)
}

func (t *Tables) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid tables: %[1]T(%[1]v)", value)
	}

	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, t)
}

func newMPTable() (Table, error) {
	monitoringColumns := []Column{
		Column{Name: `{"en": "Year", "pl": "Rok", "ro": "An", "au": "Jahr", "lv":"Gadā", "bg": "Година"}`, Kind: Count, Headers: []string{`{"en": "Symbol", "pl": "Oznaczenie", "ro": "Simbol", "au": "Symbol", "lv":"Simbol", "bg": "Символ"}`, `{"en": "Unit", "pl": "Jednostka", "ro": "Unitate", "au": "Einheit", "lv":"Vienība", "bg": "Единици"}`}},
		Column{Name: `{"en": "Month", "pl": "Miesiąc", "ro": "Lună", "au": "Monat", "lv":"Mēnesī", "bg": "Месец"}`, Kind: Name, Headers: []string{`{"en": "Symbol", "pl": "Oznaczenie", "ro": "Simbol", "au": "Symbol", "lv":"Simbol", "bg": "Символ"}`, `{"en": "Unit", "pl": "Jednostka", "ro": "Unitate", "au": "Einheit", "lv":"Vienība", "bg": "Единици"}`}},
		Column{Name: `{"en": "Heating days", "pl": "Liczba dni ogrzewania", "ro": "Numărul zilelor de încălzire efectivă", "au": "Anzahl Heiztage", "lv":"Apkures dienu skaits", "bg": "Брой на дни"}`, Kind: Count, Headers: []string{"$D_{Apk}$", `{"en": "Days", "pl": "Dni", "ro": "Zile", "au": "Tage", "lv":"Dienas", "bg": "Дни"}`}},
		Column{Name: `{"en": "Total heat energy consumption", "pl": "Całkowite zużycie energii cieplnej", "ro": "Consumul total e energie termică", "au": "Gesamter Wärmeenergieverbrauch", "lv":"Kopējais siltumenerģijas patēriņš", "bg": "Total heat energy consumption"}`, Kind: Energy, Headers: []string{"$Q_{t}$", "MWh"}},
		Column{Name: `{"en": "Domestic hot water consumption", "pl": "Zużycie ciepłej wody użytkowej", "ro": "Consumul de apă caldă menajeră", "au": "Warmwasserverbrauch", "lv":"Mājsaimniecību karstā ūdens patēriņš", "bg": "Domestic hot water consumption"}`, Kind: Volume, Headers: []string{"V", "m³"}},
		Column{Name: `{"en": "Domestic hot water temperature", "pl": "Temperatura ciepłej wody użytkowej", "ro": "Temperatura apei calde menajere", "au": "Warmwasse rtemperatur", "lv":"Mājsaimniecību karstā ūdens  temperatūra", "bg": "Domestic hot water temperature"}`, Kind: Temperature, Headers: []string{"0ku", "°C"}},
		Column{Name: "Measured by", Kind: String, Headers: []string{"", ""}},
		Column{Name: "Measured date", Kind: String, Headers: []string{"date", "date"}},
		Column{Name: "Savings deviations", Kind: Count, Headers: []string{"", ""}},
	}
	monitoringRows := []Row{
		Row{"1", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"1", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 2
		Row{"2", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"2", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 3
		Row{"3", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"3", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 4
		Row{"4", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"4", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 5
		Row{"5", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"5", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 6
		Row{"6", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"6", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 7
		Row{"7", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"7", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 8
		Row{"8", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"8", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 9
		Row{"9", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"9", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 10
		Row{"10", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"10", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 11
		Row{"11", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 12
		Row{"11", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"11", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 13
		Row{"13", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"13", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 14
		Row{"14", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"14", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		//Year 15
		Row{"15", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"15", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"16", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"17", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"18", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"19", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "January", "pl": "Styczeń", "ro": "Ianuarie", "au": "Januar", "lv":"Janvāris", "bg": "Януари"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "February", "pl": "Luty", "ro": "Februarie", "au": "Februar", "lv":"Februāris", "bg": "Февруару"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "March", "pl": "Marzec", "ro": "Martie", "au": "März", "lv":"Marts", "bg": "Март"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "April", "pl": "Kwiecień", "ro": "Aprilie", "au": "April", "lv":"Aprīlis", "bg": "Април"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "May", "pl": "Maj", "ro": "Mai", "au": "Mai", "lv":"Maijs", "bg": "Май"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "June", "pl": "Czerwiec", "ro": "Iunie", "au": "Juni", "lv":"Jūnijs", "bg": "Юни"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "July", "pl": "Lipiec", "ro": "Iulie", "au": "Juli", "lv":"Jūlijs", "bg": "Юли"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "August", "pl": "Sierpień", "ro": "August", "au": "August", "lv":"Augusts", "bg": "Август"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "September", "pl": "Wrzesień", "ro": "Septembrie", "au": "September", "lv":"Septembris", "bg": "Септември"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "October", "pl": "Październik", "ro": "Octombrie", "au": "Oktober", "lv":"Oktobris", "bg": "Октомври"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "November", "pl": "Listopad", "ro": "Noiembrie", "au": "November", "lv":"Novembris", "bg": "Ноември"}`, "", "", "", "", "", "", ""},
		Row{"20", `{"en": "December", "pl": "Grudzień", "ro": "Decembrie", "au": "Dezember", "lv":"Decembris", "bg": "Декември"}`, "", "", "", "", "", "", ""},
	}
	return NewTable(monitoringColumns, monitoringRows...)

}

func newMeasurementsTable() (Table, error) {

	measurementColumns := []Column{
		Column{Name: `{"en": "No.", "pl": "No.", "ro": "No.", "au": "No.", "lv":"No.", "bg": "No."}`, Kind: String},
		Column{Name: `{"en": "Payment Date", "pl": "Payment Date", "ro": "Payment Date", "au": "Payment Date", "lv":"Payment Date", "bg": "Payment Date"}`, Kind: String},
		Column{Name: `{"en": "Beginning Balance", "pl": "Beginning Balance", "ro": "Beginning Balance", "au": "Beginning Balance", "lv":"Beginning Balance", "bg": "Beginning Balance"}`, Kind: Money},
		Column{Name: `{"en": "Payment", "pl": "Payment", "ro": "Payment", "au": "Payment", "lv":"Payment", "bg": "Payment"}`, Kind: Money},
		Column{Name: `{"en": "Ending Balance", "pl": "Ending Balance", "ro": "Ending Balance", "au": "Ending Balance", "lv":"Ending Balance", "bg": "Ending Balance"}`, Kind: Money},
	}

	measurementRows := monthsRows(240)

	return NewTable(measurementColumns, measurementRows...)

}

func monthsRows(l int) []Row {
	rows := []Row{}
	for i := 1; i <= l; i++ {
		rows = append(rows, Row{Cell(fmt.Sprintf("%d", i)), "", "", "", ""})
	}
	return rows
}
