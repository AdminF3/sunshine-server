package contract

// NewFields creates empty map with all supported contract fields.
func NewFields() map[string]string {
	return map[string]string{
		"date":                            "",
		"client_id":                       "",
		"client_name":                     "",
		"invoiced_days":                   "",
		"country_adj":                     "",
		"country_issuer":                  "",
		"contractor_name":                 "",
		"contractor_id":                   "",
		"contractor_phone":                "",
		"contractor_email":                "",
		"client_phone":                    "",
		"client_email":                    "",
		"client_representative_name":      "",
		"client_representative_id":        "",
		"client_representative_phone":     "",
		"client_representative_email":     "",
		"contractor_representative_name":  "",
		"contractor_representative_id":    "",
		"contractor_representative_phone": "",
		"contractor_representative_email": "",
		"date-of-meeting":                 "",
		"address-of-building":             "",
		"meeting-opened-by":               "",
		"chair-of-meeting":                "",
		"meeting-recorded-by":             "",
		"tab1-for-n":                      "",
		"tab1-against-n":                  "",
		"measurement-implementer":         "",
		"tab2-for-n":                      "",
		"tab2-against-n":                  "",
		"building-administrator":          "",
		"tab3-for-n":                      "",
		"tab3-against-n":                  "",
		"contractor_fin_contribution":     "",
		"interest_rate_percent":           "",
		"interest_rate_offerter":          "",
		"floating_part":                   "",
		"start_date_of_loan":              "",
	}
}

// NewAgreement creates empty map with all the supported fields for the
// contract agreement.
func NewAgreement() map[string]string {
	return map[string]string{
		"assignor-name":                     "",
		"assignors-client-name":             "",
		"date-of-energy-contract":           "",
		"date-of-forfaiting-agreement":      "",
		"energy-efficient-performance":      "",
		"energy-saving-amount":              "",
		"forfaiting-assignee":               "",
		"manager-name":                      "",
		"payment_first_amount":              "",
		"payment_first_date":                "",
		"payment_second_amount":             "",
		"payment_second_date":               "",
		"payment_third_amount":              "",
		"payment_third_date":                "",
		"place-of-forfaiting-agreement":     "",
		"front_page_free_text":              "",
		"front_page_date":                   "",
		"forfaiting_assignee_payment":       "",
		"forfaiting_assignee_payment_words": "",
		"performance_fee":                   "",
		"late_payment_fee":                  "",
		"late_payment_fee_words":            "",
		"outstanding_amount":                "",
		"outstanding_amount_words":          "",
		"annex2_tarea":                      "",
		"annex3_tarea":                      "",
		"annex4_tarea":                      "",
	}
}

// NewMaintenance creates empty map needed for all maintenance period
// inputs. It is empty initially, unlike all other contract tables,
// gets filled with each entry.
func NewMaintenance() map[string]string {
	return map[string]string{}
}
