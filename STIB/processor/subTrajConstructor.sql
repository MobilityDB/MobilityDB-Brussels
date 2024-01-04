WITH
	get_real_terminus AS (
		SELECT stopid as stop  from  stopdetails where stopid LIKE '%' || '9600' || '%' 
	),
	--Get coordinate of the stop 
	points AS (
		SELECT ST_SetSRID(ST_MakePoint(p1.stop_lon,p1.stop_lat),4326) as point1
		FROM stops as p1
		WHERE p1.stop_id  LIKE '%' || '1131B' || '%'  --Replace by pointID
	),
	
	-- Calculate the start and end fractions
	fractions AS (
		SELECT
			ST_LineLocatePoint(trajectory.trajectory, points.point1) AS start_fraction
		FROM
			terminus_shapes AS trajectory, points, get_real_terminus as grt
		WHERE
			trajectory.route_short_name = '12' AND trajectory.stopid = grt.stop --Replace by  terminus ID
		LIMIT 1
	),
	
	-- Extract the sub-trajectory
	sub_trajectory AS (
		SELECT
			ST_SetSRID(ST_LineSubstring(trajectory.trajectory, fractions.start_fraction,1),4326) AS sub_trajectory
		FROM
			terminus_shapes AS trajectory, fractions, get_real_terminus as grt
		WHERE 
			trajectory.route_short_name = '12' AND trajectory.stopid = grt.stop --Replace by  terminus ID
		limit 1
)

select * from  sub_trajectory

/*
SELECT ST_astext(ST_LineInterpolatePoint(st.sub_trajectory, 10/ST_Length(st.sub_trajectory::geography))) --Replace 150 by  dustance in  Rt position
FROM sub_trajectory  as st
*/
--select ST_length(sub_trajectory.sub_trajectory)  from sub_trajectory

--select  * from get_real_terminus