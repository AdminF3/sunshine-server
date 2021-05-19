-- +goose Up
ALTER TABLE projects
ADD column flat_air_temperature real;

ALTER TABLE forfaiting_applications
DROP COLUMN swift;
ALTER TABLE bank_accounts
ADD COLUMN swift TEXT;

UPDATE contracts SET agreement = agreement #- '{contractor_name}' #- '{contractor_id}' #- '{contractor_phone}' #- '{contractor_email}' #- '{client_name}' #- '{client_id}' #- '{client_phone}' #- '{client_email}' #- '{client_address}' #- '{client_representative_name}' #- '{client_representative_id}' #- '{client_representative_phone}' #- '{client_representative_email}' #- '{date-of-meeting}' #- '{address-of-building}' #- '{meeting-opened-by}' #- '{chair-of-meeting}' #- '{meeting-recorded-by}' #- '{tab1-for-n}' #- '{tab1-against-n}' #- '{measurement-implementer}' #- '{tab2-for-n}' #- '{tab2-against-n}' #- '{building-administrator}' #- '{tab3-for-n}' #- '{tab3-against-n}' #- '{contractor_fin_contribution}' #- '{interest_rate_percent}' #- '{interest_rate_offerter}' #- '{floating_part}' #- '{start_date_of_loan}';

-- +goose Down
ALTER TABLE projects
DROP column flat_air_temperature;

ALTER TABLE forfaiting_applications
ADD COLUMN swift TEXT;
ALTER TABLE bank_accounts
DROP COLUMN swift;
