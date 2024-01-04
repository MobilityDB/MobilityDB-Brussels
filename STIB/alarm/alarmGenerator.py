from pymeos import *
from pymeos.db.psycopg2 import MobilityDB

pymeos_initialize()

host = 'localhost'
port = 5432
db = 'brussels'
user = 'postgres'
password = 'postgres'

connection = MobilityDB.connect(host=host, port=port, database=db, user=user, password=password)
cursor = connection.cursor()


cursor.execute("select lineid,tripid, stops(trip,200,'10 minutes') from stib_trips where numInstants(trip) > 1;")

res = cursor.fetchone()
while res != None :
    if res[0] == '95':
        print(res[0], res[1],res[2])
    res = cursor.fetchone()




pymeos_finalize()



