import json
import psycopg2
import pandas as pd
import os


def dbConnection():
    # Credentials to log into  the databse
    host = "localhost"
    user = "postgres"
    password = "postgres"
    database = "BM"

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


def RtPositionTableCreate():
    # Define the table name and column names
    table_name = 'real_tile_position'
    column1_name = 'terminusID'
    column2_name = 'meter_from_last_stop'
    column3_name = 'last_stops'

    #Query to create the table if do not existe in the db
    query = f"""CREATE TABLE IF NOT EXISTS {table_name} (
            id serial PRIMARY KEY,
            {column1_name} text,
            {column2_name} tint,
            {column3_name} tgeompoint
        )"""

    cursor.execute(query)
    conn.commit()








cursor,conn = dbConnection()
#RtPositionTableCreate()




#Code pour faire le tableau  pour l'analyse des données avec le reverse engineering
"""
folder_path = "../api/rtPositions/"
complete_df = pd.DataFrame( columns=['lineId', 'directionId', 'distanceFromPoint', 'pointId'])

for filename in os.listdir(folder_path):
    file_path = os.path.join(folder_path,filename)
    #print(file_path)
    if os.path.isfile(file_path):
        # Remplacez le chemin du fichier par votre fichier JSON
        with open(file_path, 'r') as json_file:
            data = json.load(json_file)

        # Créez une liste pour stocker les données extraites
        tableau = []

        # Parcourez les résultats
        for result in data['results']:
            line_id = result['lineid']
            vehicle_positions = json.loads(result['vehiclepositions'])

            # Parcourez les positions des véhicules
            for position in vehicle_positions:
                direction_id = position['directionId']
                distance_from_point = position['distanceFromPoint']
                point_id = position['pointId']

                # Ajoutez les données à la liste
                tableau.append([line_id, direction_id, distance_from_point, point_id])

        # Créez un DataFrame Pandas à partir de la liste
        df = pd.DataFrame(tableau, columns=['lineId', 'directionId', 'distanceFromPoint', 'pointId'])
        print("New dataframe shape is",df.shape)
        print("Actual size of concatenated dataframe is",complete_df.shape)
        print(complete_df.empty)
        # Affichez le DataFrame
        complete_df = pd.concat([complete_df, df])
        print("The new size is",complete_df.shape)
        print("==================================================")
complete_df.to_excel("./test.xlsx", index=False)
"""