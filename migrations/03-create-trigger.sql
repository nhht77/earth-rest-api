-- condition: SELECT EXISTS(SELECT tgname FROM pg_trigger WHERE tgname = 'continent_updated') condition --
CREATE TRIGGER continent_updated
    BEFORE UPDATE ON continent FOR EACH ROW
    EXECUTE PROCEDURE trigger_timestamp_updated();

-- condition: SELECT EXISTS(SELECT tgname FROM pg_trigger WHERE tgname = 'country_updated') condition --
CREATE TRIGGER country_updated
    BEFORE UPDATE ON country FOR EACH ROW
    EXECUTE PROCEDURE trigger_timestamp_updated();

-- condition: SELECT EXISTS(SELECT tgname FROM pg_trigger WHERE tgname = 'city_updated') condition --
CREATE TRIGGER city_updated
    BEFORE UPDATE ON city FOR EACH ROW
    EXECUTE PROCEDURE trigger_timestamp_updated();