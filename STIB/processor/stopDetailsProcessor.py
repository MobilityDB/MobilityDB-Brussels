import json
import psycopg2


def dbConnection():
    # Credentials to log into  the databse
    host = "localhost"
    user = "postgres"
    password = "postgres"
    database = "brussels"

    # Establish a connection
    connection = psycopg2.connect(
        host=host,
        user=user,
        password=password,
        database=database
    )
    #cursor creation
    cursor = connection.cursor()

    return cursor,connection


def stopDetailsTableCreate():
    # Define the table name and column names
    table_name = 'stopDetails'
    column1_name = 'stopID'
    column2_name = 'coordinates'
    column3_name = 'name'

    #Query to create the table if do not existe in the db
    query = f"""CREATE TABLE IF NOT EXISTS {table_name} (
            id serial PRIMARY KEY,
            {column1_name} text,
            {column2_name} GEOMETRY,
            {column3_name} text
        )"""

    cursor.execute(query)
    conn.commit()

def stopDetailsTableFilling(cursor,filepath) :
    with open(filepath,'r') as file :
        data = json.load(file)
        file.close()

    query = f"""
                INSERT INTO stopDetails (stopID, coordinates, name) 
                VALUES (%s, ST_GeomFromText(%s), %s)
                """
    for i in range(len(data['results'])) :
        # Select coords
        coords = data['results'][i]['gpscoordinates'].split(",")
        coords[0]=coords[0][13:]
        coords[1]=coords[1][14:-1]
        coords = f"POINT({coords[0]} {coords[1]})"
        # select ID
        id = data["results"][i]['id']

        # Select name
        name = data['results'][i]['name'][1:-1]


        cursor.execute(query, (id, coords, name))
        conn.commit()


cursor,conn = dbConnection()
stopDetailsTableCreate()
for i in range(0,2600,100):
    filename = "../api/stopDetails/stopsDetails" + str(i) +"-"+str(i+100)+".json"
    print("Uplaoding file",filename)
    stopDetailsTableFilling(cursor,filename)




