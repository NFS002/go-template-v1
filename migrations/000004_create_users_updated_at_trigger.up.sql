CREATE OR REPLACE TRIGGER users_updated_at_trigger
    BEFORE UPDATE
    ON
        users
    FOR EACH ROW
EXECUTE PROCEDURE auto_set_update_at();