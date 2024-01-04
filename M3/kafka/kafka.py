import configparser
from confluent_kafka import Consumer
import json
from psycopg2.extras import Json
from datetime import datetime, timedelta
import psycopg2


# Function to establish a connection to the PostgreSQL database
def connect():
    conn = psycopg2.connect(
        database="brussels",
        user="postgres",
        password="postgres",
        host="localhost"
    )
    return conn

# Function to set current to false if the time difference is equal or greater than 1 hour
def set_current_false(conn):
    cursor = conn.cursor()
    query = "UPDATE M3_COUNTING SET current = FALSE WHERE (end_timestamp - start_timestamp) >= interval '1 hour';"
    cursor.execute(query)
    conn.commit()
    cursor.close()

# Function to check if the ID exists in the table and current is true
def check_id_exists_with_true_current(conn, camera_id):
    cursor = conn.cursor()
    query = "SELECT * FROM M3_COUNTING WHERE camera_id = %s AND current = TRUE;"
    cursor.execute(query, (camera_id,))
    result = cursor.fetchone()
    cursor.close()
    return result is not None

# Function to update the end timestamp and count
def update_count(conn, camera_id, end_timestamp, count):
    cursor = conn.cursor()
    query = "UPDATE M3_COUNTING SET end_timestamp = %s, count = update_count(%s , count) WHERE camera_id = %s AND current = TRUE;"
    cursor.execute(query, (end_timestamp, count, camera_id))
    conn.commit()
    cursor.close()

# Function to insert a new row
def insert_row(conn, data):
    cursor = conn.cursor()
    query = "INSERT INTO M3_COUNTING (camera_id, location, start_timestamp, end_timestamp, count) VALUES (%s, %s, %s, %s, %s);"
    cursor.execute(query, (
        data["_id"],
        Json(data["location"]),
        datetime.utcfromtimestamp(data["_start_timestamp"]["$date"] / 1000),
        datetime.utcfromtimestamp(data["_end_timestamp"]["$date"] / 1000),
        data["count"]
    ))
    conn.commit()
    cursor.close()
def process_count(data):
    conn = connect()

    # Step 1: Set current to false if the time difference is equal or greater than 1 hour
    set_current_false(conn)

    # Step 2: For each entry in the data, update or insert as needed
    for entry in json_data:
        camera_id = entry["_id"]

        if check_id_exists_with_true_current(conn, camera_id):
            # If true, update the end timestamp and count
            update_count(
                conn,
                camera_id,
                datetime.utcfromtimestamp(entry["_end_timestamp"]["$date"] / 1000),
                entry["count"]
            )
        else:
            # If false, insert a new row
            insert_row(conn, entry)

    conn.close()

def process_stats(data):

    return None


if __name__ == "__main__":

    config = configparser.ConfigParser()
    config.read('client.properties')

    config = dict(config["KAFKA_CONSUME"])

    print(config)
    consumer = Consumer(config)
    consumer.subscribe(["vehicle_count"])

    try:
        while True:
            msg = consumer.poll(5)
            if msg is not None and msg.error() is None:
                json_data = json.loads(msg.value().decode('utf-8'))
                print("=======================")
                print("I received a new batch of data")
                keys = list(json_data[0].keys())
                if "median_speed" not in keys :
                    process_count(json_data)
                else:
                    process_stats(json_data)
    except KeyboardInterrupt:
            pass
    finally:
            consumer.close()