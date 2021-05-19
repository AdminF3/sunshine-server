-- +goose Up
DELETE FROM "organization_roles" WHERE "deleted_at" IS NOT NULL;
ALTER TABLE "organization_roles" DROP COLUMN "deleted_at";

DELETE FROM "project_roles" WHERE "deleted_at" IS NOT NULL;
ALTER TABLE "project_roles" DROP COLUMN "deleted_at";

-- +goose Down
ALTER TABLE "organization_roles" ADD COLUMN "deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL;
ALTER TABLE "project_roles" ADD COLUMN "deleted_at" TIMESTAMP WITH TIME ZONE DEFAULT NULL;
