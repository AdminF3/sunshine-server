-- +goose Up

/*This migration is duplicate for 00092 
This migration did not go up because te command
up was written with lower case
 and the goose ignored the up command
the issue is when we try do goose down
in migration 92 it tries to drop the colum, but it was dropped here.
*/

SELECT 1;


-- +goose Down
SELECT 1;

