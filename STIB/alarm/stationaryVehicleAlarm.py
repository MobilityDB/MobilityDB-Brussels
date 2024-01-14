import re
import json
from pymeos import *
from pymeos.db.psycopg2 import MobilityDB


def parse_astext_field(astext):
    # Split the string into points
    points = astext.strip('[]').split(', ')

    # Extract the timestamp from the first point
    first_point = points[0]
    match_first = re.search(r'@(.+)$', first_point)
    start_time = match_first.group(1) if match_first else None

    # Extract the coordinates from the last point
    last_point = points[-1]
    match_last = re.search(r'POINT\((.+)\)', last_point)
    coords = match_last.group(1).split() if match_last else (None, None)
    x, y = map(float, coords)

    return start_time, x, y


# Initialize the MEOS library
pymeos_initialize()

# Database connection parameters
host = 'localhost'
port = 5432
db = 'brussels'
user = 'postgres'
password = 'postgres'

# Connect to the MobilityDB
connection = MobilityDB.connect(host=host, port=port, database=db, user=user, password=password)
cursor = connection.cursor()

# SQL query to execute
#Thedistance writtent in the report is  not the same here because  that unit mesure in SRID 4326 is not meter
query = """
select lineid,tripid,stop_name,transport_type,asText(unnest(sequences(stops(trip,0.0,'10 minutes')))) 
from stib_trips as st, stops, transport_type as tt
where tempSubType(trip) != 'Instant' and
		st.directionid = stop_id and
		st.current = true and is_deviated = false and
		st.lineid = tt.route_short_name;
"""

# Execute the query
cursor.execute(query)

# List to hold the entries for the JSON file
json_entries = []

# Fetch one result at a time
res = cursor.fetchone()
while res is not None:
    # Parse the 'astext' field to extract the start time and last position
    start_time, x, y = parse_astext_field(res[4])

    # Create the JSON entry
    json_entry = {
        "lineid": res[0],
        "tripid": res[1],
        "direction": res[2],
        "transportType": res[3],
        "position": {"x": x, "y": y},
        "startTime": start_time,
        "type": "NOT_MOVING",
        "description": "The vehicle is not  moving since at least 5 minutes"
    }

    # Add the entry to the list
    json_entries.append(json_entry)

    # Fetch the next result
    res = cursor.fetchone()

# Finalize the MEOS library
pymeos_finalize()

# Close the cursor and connection
cursor.close()
connection.close()

# Convert the list of entries to a JSON string
json_data = json.dumps(json_entries, indent=4)

# Write the JSON string to a file
json_file_path = './data/vehicle_stops.json'
with open(json_file_path, 'w') as json_file:
    json_file.write(json_data)

# Output the path to the JSON file for reference
print(f"JSON file created at {json_file_path}")
