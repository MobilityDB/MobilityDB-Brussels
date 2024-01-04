DROP TABLE IF EXISTS agency CASCADE;
CREATE TABLE agency (
	agency_id text,
	agency_name text,
	agency_url text,
	agency_timezone text,
	agency_lang text, 
	agency_phone text
);

DROP TABLE IF EXISTS calendar CASCADE;
CREATE TABLE calendar (
	service_id int,
	monday int,
	tuesday int,
	wednesday int,
	thursday int,
	friday int,
	saturday int,
	sunday int,
	start_date int,
	end_date int
);

DROP TABLE IF EXISTS calendar_dates CASCADE;
CREATE TABLE calendar_dates (
	service_id int,
	date int,
	exception_type int
);


DROP TABLE IF EXISTS routes CASCADE;
CREATE TABLE routes (
	route_id text,
	route_short_name text,
	route_long_name text,
	route_desc text,
	route_type int,
	route_url text,
	route_color text,
	route_text_color text
);


DROP TABLE IF EXISTS shapes CASCADE;
CREATE TABLE shapes (
	shape_id text not null,
	shape_pt_lat double precision,
	shape_pt_lon double precision,
	shape_pt_sequence int not null
);

DROP TABLE IF EXISTS stop_times CASCADE;
CREATE TABLE stop_times (
	trip_id text not null,
	arrival_time interval CHECK (arrival_time::interval = arrival_time::interval),
	departure_time interval CHECK (departure_time::interval = departure_time::interval),
	stop_id text,
	stop_sequence int not null,
	pickup_type int,
	drop_off_type int
);


DROP TABLE IF EXISTS stops CASCADE;
CREATE TABLE stops (
	stop_id text not null,
	stop_code text,
	stop_name text,
	stop_desc text,
	stop_lat double precision,
	stop_lon double precision,
	zone_id text,
	stop_url text,
	location_type text,
	parent_station text
);

DROP TABLE IF EXISTS trips CASCADE;
CREATE TABLE trips (
	route_id text not  null,
	service_id text not null,
	trip_id text not null,
	trip_headsign text,
	direction_id int not null,
	block_id int not null,
	shape_id text  not null
);

DROP TABLE IF EXISTS stops_by_lines CASCADE;
CREATE TABLE stops_by_lines (
    lineid TEXT,
    destination TEXT,
	direction TEXT,
    stops TEXT[] -- Array of integers to store the IDs of the stops
);

CREATE TABLE shape_geoms (
	shape_id TEXT,
	trajectory geometry
);


-- Geometries set up
INSERT INTO shape_geoms
SELECT shape_id, ST_MakeLine(array_agg(
ST_SetSRID(ST_MakePoint(shape_pt_lon, shape_pt_lat), 4326) ORDER BY shape_pt_sequence))
FROM shapes
GROUP BY shape_id;


--Create a table to  assign teminus with  the shape
DROP TABLE IF EXISTS terminus_shapes;
CREATE TABLE terminus_shapes as (
select DISTINCT r.route_short_name, sd.stopid,t.trip_headsign, t.shape_id, s.trajectory
from trips as t, shape_geoms as s, stopdetails as sd, routes as r  
where  sd.name LIKE '%' || t.trip_headsign || '%' 
	and r.route_id = t.route_id
	and s.shape_id = t.shape_id
);


DROP TABLE IF EXISTS stib_trips;
CREATE TABLE stib_trips (
	day DATE,
	lineid TEXT,
	tripid TEXT,
	directionid TEXT,
	start_timestamp TIMESTAMPTZ,
	end_timestamp TIMESTAMPTZ,
	is_deviated BOOLEAN,
	current BOOLEAN,
	trip tgeompoint
);