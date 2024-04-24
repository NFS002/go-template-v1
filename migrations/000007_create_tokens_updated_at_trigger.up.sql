CREATE OR REPLACE TRIGGER tokens_updated_at_trigger
    BEFORE UPDATE
    ON
        tokens
    FOR EACH ROW
EXECUTE PROCEDURE auto_set_update_at();