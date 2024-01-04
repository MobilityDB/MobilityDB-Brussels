import os
import pandas as pd
import psycopg2
from psycopg2 import sql

# Define the PostgreSQL connection parameters
db_params = {
    "host": "localhost",
    "database": "brussels",
    "user": "postgres",
    "password": "postgres",
}

# Directory containing your GTFS CSV files
gtfs_data_dir = "/home/wassim/MemoBM/STIB/GTFS/"

# Create a PostgreSQL connection
try:
    conn = psycopg2.connect(**db_params)
    cursor = conn.cursor()
    print("Connected to the database")

    # Define a function to load CSV data into a PostgreSQL table
    def load_data(table_name, csv_file):
        with open(csv_file, "r") as f:
            # Create a PostgreSQL COPY query
            copy_sql = sql.SQL("COPY {} FROM stdin WITH CSV HEADER").format(sql.Identifier(table_name))
            cursor.copy_expert(sql=copy_sql, file=f)
            conn.commit()
            print(f"Data loaded into {table_name} table")

    # List of GTFS tables and corresponding CSV files
    gtfs_tables = {
        "agency": "agency.txt",
        "calendar_dates": "calendar_dates.txt",
        "calendar": "calendar.txt",
        "routes": "routes.txt",
        "stop_times": "stop_times.txt",
        "stops": "stops.txt",
        "shapes": "shapes.txt",
        "trips": "trips.txt"
    }

    # Load data into PostgreSQL tables
    for table_name, csv_file in gtfs_tables.items():
        csv_file_path = os.path.join(gtfs_data_dir, csv_file)
        load_data(table_name, csv_file_path)

    cursor.close()
    conn.close()
    print("Database connection closed")

except (Exception, psycopg2.Error) as error:
    print("Error connecting to the database:", error)