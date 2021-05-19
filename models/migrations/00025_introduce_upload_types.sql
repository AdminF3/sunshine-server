-- +goose Up
CREATE TYPE upload_type AS ENUM (
	'general leaflet', -- General leaflet
 	'process leaflet', -- Process leaflet for owners
	'template agenda', -- Template agenda for meeting
	'template protocol of meeting', -- Template protocol of meeting
	'template of survey', -- Template of survey
	'renovation vote', -- Vote for to go with renovation
	'signed protocol to start project', -- Signed protocol to start project development
	'agreement on cost, audit and compensation', -- Agreements on costs, audit and compensation
	'template of tender documentation', -- Template of tender documentation
	'energy audit report', -- Energy audit report
	'technical report', -- Technical inspection report
	'template of agreements', -- Template of agreements
	'agenda', -- Agenda
	'protocol', -- Protocol
	'survey', -- Survey
	'vote', -- Vote for/against
	'signed protocol to continue project', -- Signed protocol to continue project development
	'agreement', -- Agreement to prepare ALTUM documentation
	'decision on financial option', -- Decision on financial option
	'technical design', -- Technical design
	'cost estimates', -- Certified cost estimates
	'procurement', -- Procurement of construction and installation works(templates)
	'tender documentation', -- Tender documentation
	'financing offer', -- Financing offer and ALTUM application
	'project design documentation', -- Project design documentation
	'draft epc contract', -- Draft EPC contract
	'epc vote', -- Vote for EPC and/or agreement with ALTUM
	'presentation', -- Presentations
	'signed epc', -- Signed EPC(and uploaded for approval back to the platform)
	'signed agreement' -- Signed agreement with ALTUM and/or bank loan
);

ALTER TABLE attachments ADD COLUMN upload_type upload_type;

-- +goose Down
ALTER TABLE attachments DROP COLUMN upload_type;
DROP TYPE upload_type;
