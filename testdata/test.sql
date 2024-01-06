-- TEST1
SELECT * FROM assistant_services as ass
    JOIN assistants as a ON a.id=ass.assistant_id
    JOIN vehicles as v ON v.id=a.vehicle_id
    WHERE ass.id=$1;

-- TEST2
UPDATE assistants SET provider_id = $1 WHERE id = $1;

-- CITIES
SELECT c.id, c.name, c.country_id FROM cities;