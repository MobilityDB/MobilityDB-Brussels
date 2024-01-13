CREATE OR REPLACE FUNCTION insert_or_update_stib_trip(
    p_day DATE,
    p_lineid TEXT,
    p_tripid TEXT,
    p_directionid TEXT,
    p_start_timestamp TIMESTAMPTZ,
    p_end_timestamp TIMESTAMPTZ,
    p_is_deviated BOOLEAN,
    p_current BOOLEAN,
    p_trip tgeompoint
)
RETURNS VOID AS $$
BEGIN
    -- Check if there is an existing record with the same day, lineid, and tripid
    PERFORM 1
    FROM stib_trips
    WHERE day = p_day AND lineid = p_lineid AND tripid = p_tripid AND directionid = p_directionid AND current;

    IF FOUND THEN
        -- Update the existing record
        UPDATE stib_trips
        SET
            end_timestamp = p_end_timestamp,
            is_deviated = p_is_deviated,
            trip = appendInstant(stib_trips.trip,p_trip)
        WHERE
            day = p_day AND lineid = p_lineid AND tripid = p_tripid AND directionid = p_directionid AND current;
    ELSE
        -- Insert a new record
        INSERT INTO stib_trips (day, lineid, tripid, directionid, start_timestamp, end_timestamp, is_deviated, current, trip)
        VALUES (p_day, p_lineid, p_tripid, p_directionid, p_start_timestamp, p_end_timestamp, p_is_deviated, p_current, p_trip);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Define a function to calculate the updated value for trip  column
CREATE OR REPLACE FUNCTION update_trip(existing_tgeompoint tgeompoint, new_tgeompoint tgeompoint)
RETURNS tgeompoint AS $$
DECLARE
    updated_tgeompoint tgeompoint;
BEGIN
    SELECT appendInstant(existing_tgeompoint,new_tgeompoint) INTO updated_tgeompoint;
    RETURN updated_tgeompoint;
END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION public.historical_trip(
			z integer, x integer, y  integer, p_tripid text, p_start text, p_end text)
RETURNS bytea
AS $$
	WITH bounds AS (
		SELECT ST_TileEnvelope(z,x,y) as geom
	),
	vals AS (
		SELECT asMVTGeom(transform(attime(trip,span(p_start::timestamptz, p_end::timestamptz, true, true)),3857), transform((bounds.geom)::stbox,3857))
			as geom_times
		FROM stib_trips, bounds
		WHERE tripid = p_tripid
	),
	mvtgeom AS (
		SELECT (geom_times).geom, (geom_times).times
		FROM vals
	)
	SELECT ST_AsMVT(mvtgeom) 
	FROM mvtgeom
$$
LANGUAGE 'sql'
STABLE
PARALLEL SAFE;


CREATE OR REPLACE FUNCTION public.historical_line(
			z integer, x integer, y  integer, p_lineid text)
RETURNS bytea
AS $$
	WITH bounds AS (
		SELECT ST_TileEnvelope(z,x,y) as geom
	),
	vals AS (
		SELECT tripid,asMVTGeom(transform(trip,3857), transform((bounds.geom)::stbox,3857))
			as geom_times
		FROM stib_trips, bounds
		WHERE lineid = p_lineid
	),
	mvtgeom AS (
		SELECT (geom_times).geom, (geom_times).times,tripid
		FROM vals
	)
	SELECT ST_AsMVT(mvtgeom) 
	FROM mvtgeom
$$
LANGUAGE 'sql'
STABLE
PARALLEL SAFE;




CREATE OR REPLACE FUNCTION public.trip_last_instant(z integer, x integer, y integer, p_tripid text)
RETURNS bytea
AS $$
WITH bounds AS (
  SELECT ST_TileEnvelope(z, x, y) AS geom
),
last_instant AS (
  SELECT tripid,asMVTGeom(transform(endinstant(trip), 3857), (bounds.geom)::stbox) AS geom_times
  FROM stib_trips, bounds
  WHERE tripid = p_tripid and current = true -- Filter by 'current' column
),
mvtgeom AS (
  SELECT (geom_times).geom, (geom_times).times
  FROM last_instant
)
SELECT ST_AsMVT(mvtgeom)
FROM mvtgeom
$$
LANGUAGE 'sql'
STABLE
PARALLEL SAFE;

CREATE OR REPLACE FUNCTION public.line_last_instant(z integer, x integer, y integer, p_lineid text)
RETURNS bytea
AS $$
WITH bounds AS (
  SELECT ST_TileEnvelope(z, x, y) AS geom
),
last_instant AS (
  SELECT asMVTGeom(transform(endinstant(trip), 3857), (bounds.geom)::stbox) AS geom_times
  FROM stib_trips, bounds
  WHERE lineid = p_lineid and current = true -- Filter by 'current' column
),
mvtgeom AS (
  SELECT (geom_times).geom, (geom_times).times
  FROM last_instant
)
SELECT ST_AsMVT(mvtgeom)
FROM mvtgeom
$$
LANGUAGE 'sql'
STABLE
PARALLEL SAFE;


CREATE OR REPLACE
FUNCTION public.default_trajectory(
            z integer, x integer, y  integer,p_lineid text)
RETURNS bytea
AS $$
    WITH bounds AS (
        SELECT ST_TileEnvelope(z,x,y) as geom
    ),
    vals AS (
        SELECT shape_id,st_asMVTGeom(st_transform(trajectory,3857), (bounds.geom))
            as geom_times
        FROM terminus_shapes, bounds where route_short_name = p_lineid
    ),
    mvtgeom AS (
        SELECT (geom_times),shape_id
        FROM vals
    )
    SELECT ST_AsMVT(mvtgeom) FROM mvtgeom
$$
LANGUAGE 'sql'
STABLE
PARALLEL SAFE;
