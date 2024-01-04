import configparser
from confluent_kafka import Consumer
import json
import  os
import time

def process_count(data,file_path) :
    print("Hello this is counting data")
    print("Writing  in file ", filename)
    with open(file_path,'w+',encoding='utf-8') as file :
        json.dump(data,file)

    print("Writing in the aggregated file")
    with open("data/comptage.json", 'a+') as f:
        json.dump(data, f)
        f.write('\n')


def process_stats(data,file_path):
    print("Hello this is some stats about speed of vehicles")
    print("Writing  in file ", filename)
    with open(file_path, 'w+', encoding='utf-8') as file:
        json.dump(data, file)

    print("Writing in the aggregated file")
    with open("data/statistiques.json", 'a+') as f:
        json.dump(data, f)
        f.write('\n')

if __name__ == "__main__" :

    config = configparser.ConfigParser()
    config.read('client.properties')

    config = dict(config["KAFKA_CONSUME"])

    print(config)
    consumer = Consumer(config)
    consumer.subscribe(["vehicle_count"])

    count_comptage = 1
    count_stat = 1
    timestamp = ""

    try:
        while True:
            msg = consumer.poll(5)
            if msg is not None and msg.error() is None:
                if timestamp == "":
                    timestamp = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())
                    folder_name = timestamp
                    folder_path = "data/" + folder_name
                    os.makedirs(folder_path, exist_ok=True)
                json_data = json.loads(msg.value().decode('utf-8'))
                print("=======================")
                print("I received a new batch of data")
                keys = list(json_data[0].keys())
                if "median_speed" not in keys :
                    filename = "count_batch_" + str(count_comptage)
                    file_path = folder_path + "/" + filename
                    process_count(json_data,file_path)
                    count_comptage += 1
                else:
                    filename = "speed_stats_batch_" + str(count_stat)
                    file_path = folder_path + "/" + filename
                    process_stats(json_data,file_path)
                    count_stat += 1
            else:
                count_comptage = 0
                count_stat = 0
                timestamp = ""
    except KeyboardInterrupt:
            pass
    finally:
            consumer.close()