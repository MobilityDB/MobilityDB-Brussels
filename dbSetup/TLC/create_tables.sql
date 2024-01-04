CREATE TABLE traffic_lights (
    FID text,
    gid integer,
    key text,
    intersection_name text,
    b3s text,
    type text,
    geom geometry(Point, 4326)
);


--Import data from  the file that can be found on  mobigis
COPY traffic_lights(FID, gid, key, intersection_name, b3s, type, geom)
FROM '/mnt/c/Users/Wassim/Downloads/traffic_lights.csv'
DELIMITER ','
CSV HEADER;