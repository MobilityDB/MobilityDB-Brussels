import pandas as pd
import psycopg2
import json

# Define the PostgreSQL connection parameters
db_params = {
    "host": "localhost",
    "database": "brussels",
    "user": "postgres",
    "password": "postgres",
}

# Define the CSV file path
csv_file = 'stops_by_line.csv'

# Establish a connection to the PostgreSQL database
conn = psycopg2.connect(dbname=db_params["database"], user=db_params["user"], password=db_params["password"], host=db_params["host"])

# Create a cursor
cur = conn.cursor()

# Read the CSV file with the correct delimiter
df = pd.read_csv(csv_file, sep=';')

# Define the table name
table_name = 'stops_by_lines'

# Insert data from the DataFrame into the database table
for index, row in df.iterrows():
    destination = json.loads(row['destination'])
    direction = row['direction']
    lineid = row['lineId']
    points = json.loads(row['points'])

    # Sort the points data based on the 'order' field
    sorted_points = sorted(points, key=lambda x: x['order'])

    # Extract the 'id' values from the sorted points
    points_data = [point['id'] for point in sorted_points]

    # Convert the ordered "id" values to a PostgreSQL integer array
    points_array = "{" + ",".join(map(str, points_data)) + "}"
    insert_query = '''
        INSERT INTO stops_by_lines (lineid, destination, direction, stops)
        VALUES (%s, %s, %s, %s);
    '''
    cur.execute(insert_query, (lineid, json.dumps(destination), direction, points_array))
    conn.commit()

# Close the database connection
conn.close()
