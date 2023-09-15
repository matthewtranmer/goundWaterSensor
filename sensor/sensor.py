import mysql.connector
import time
import logging
import random
import os
import sys
import json
import traceback

try:
    import RPi.GPIO as GPIO
except Exception as e:
    print(e, file=sys.stderr)

class HCSR04:
    trigger = -1
    echo = -1
    trigger_pulse_length = 0.0001
    debug = False

    def __init__(self, trigger, echo, debug=False):
        self.trigger = trigger
        self.echo = echo
        self.debug = debug

        if not self.debug:
            self.init()

    def init(self):
        GPIO.setmode(GPIO.BCM)

        GPIO.setup(self.trigger, GPIO.OUT)
        GPIO.setup(self.echo, GPIO.IN)

    def __del__(self):
        if not self.debug:
            self.clean()

    def clean(self):
        GPIO.cleanup()

    def pulseTrigger(self):
        GPIO.output(self.trigger, True)
        time.sleep(self.trigger_pulse_length)
        GPIO.output(self.trigger, False)

    def getDistance(self):
        iterate = True
        while iterate:
            self.pulseTrigger()

            time_limit = 0.5
            error_time = time.time()

            count = 0
            while True:
                if GPIO.input(self.echo) == 1:
                    break

                if time.time() - error_time > time_limit:
                    logging.info("Reached time limit 1")
                    time.sleep(0.1)
                    error_time = time.time()
                    self.pulseTrigger()
                
                count += 1
            
            start_time = time.time()

            error_time = time.time()
            count = 0
            while True:
                if GPIO.input(self.echo) == 0:
                    iterate = False
                    break

                if time.time() - error_time > time_limit:
                    logging.info("Reached time limit 2")
                    break
                count += 1

            stop_time = time.time()
        
        time_elapsed = stop_time - start_time
        return time_elapsed / 2  * 343
    
    def debug_getDistance(self):
        return random.uniform(0.30, 0.40)
    
    def getAverageReading(self, readings, delay):
        vals = []
        average = 0
        count = 0
        while count < readings:
            if not self.debug:
                distance = self.getDistance()
            else:
                distance = self.debug_getDistance()

            vals.append(distance)
            
            if (distance > 0.85):
                logging.info("Greater Than 0.85m. Ignoring Reading")
                continue
            
            average += distance
            count += 1

            time.sleep(delay)

        return average / readings
    
def connect_db(db_password):
    db = mysql.connector.connect(
        host="mysql",
        user="WorkerRW",
        password=db_password,
        database="sensor"
    )

    return db

def getCalibrationInfo():
    with open("calibration.json", "r") as file:
        return json.loads(file.read())
        
def mainLoop(db, cursor, sensor):
    calibration = getCalibrationInfo()

    while True:
        average = sensor.getAverageReading(500, 0.1)
        logging.debug("Avr: " + str(average))
        distance_from_max = round(average - calibration["Min_Distance"], 2)
        logging.debug("DFM: " + str(distance_from_max))

        if distance_from_max - calibration["Min_Max_Uncertanty"] < 0:
            distance_from_max = 0

        if distance_from_max + calibration["Min_Max_Uncertanty"] > calibration["Max_Distance"]:
            distance_from_max = calibration["Max_Distance"]

        water_level = round(calibration["Max_Distance"] - distance_from_max, 2)

        logging.debug("WL: " + str(water_level))

        percentage = round(water_level / calibration["Max_Distance"] * 100, 0)

        query = "INSERT INTO readings (height, percentage, time, max_distance, min_distance, min_max_uncertanty) VALUES (%s, %s, UTC_TIMESTAMP(), %s, %s, %s);"
        cursor.execute(query, (water_level, percentage, calibration["Max_Distance"], calibration["Min_Distance"], calibration["Min_Max_Uncertanty"]))
        db.commit()  

def main():
    if "DATABASE_SECRET_FILE" not in os.environ:
        raise Exception("DATABASE_SECRET_FILE environment variable not given")

    secret_file = os.environ["DATABASE_SECRET_FILE"]

    db_password = ""
    with open(secret_file, 'r') as file:
        db_password = file.read().splitlines()[0]

    db = connect_db(db_password)

    cursor = db.cursor()

    if "TRIGGER" not in os.environ:
        raise Exception("TRIGGER environment variable not given")

    if "ECHO" not in os.environ:
        raise Exception("ECHO environment variable not given")

    trigger = int(os.environ["TRIGGER"])
    echo = int(os.environ["ECHO"])

    debug = False
    if "DEBUG" in os.environ:
        logging.info("Starting in debug mode")
        debug = True

    sensor = HCSR04(trigger, echo, debug)
    #sensor = HCSR04(14, 15)

    logging.info("Sensor Started.")
    mainLoop(db, cursor, sensor)
    
    

    

if __name__ == "__main__":
    level = logging.INFO
    if "DEBUG" in os.environ:
        level = logging.DEBUG

    logging.basicConfig(level=level, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
    logging.info("Starting Up")
    logging.debug("test")

    current_time = time.time()
    allowed_errors_per_minute = 10
    errors = 0
    
    for i in range(allowed_errors_per_minute):
        try:
            main()
        except Exception:
            logging.error(traceback.format_exc())

            errors += 1

            if (current_time > time.time() - 60):
                current_time = time.time()
                errors = 0

            if errors > allowed_errors_per_minute:
                break

            time.sleep(0.2)

    logging.critical("The error limit has been surpassed")

