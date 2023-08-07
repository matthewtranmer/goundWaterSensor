import RPi.GPIO as GPIO
import mysql.connector
import time
import logging
import random

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
        return random.uniform(0.15, 0.25)
    
    def getAverageReading(self, readings, delay):
        average = 0
        count = 0
        while count < readings:
            distance = self.getDistance()
            
            if (distance > 0.85):
                logging.info("Greater Than 0.85m. Ignoring Reading")
                continue
            
            average += distance
            count += 1

            time.sleep(delay)

        return average / readings
    
def connect_db(db_password):
    db = mysql.connector.connect(
        host="localhost",
        user="WorkerRW",
        password=db_password,
        database="sensor"
    )

    return db

def main():
    db_password = ""
    with open('/home/sensoruser/goundWaterSensor/sensor/dbpassword', 'r') as file:
        db_password = file.read().splitlines()[0]

    db = connect_db(db_password)

    cursor = db.cursor()
    sensor = HCSR04(14, 15)

    logging.info("Sensor Started")
    
    #Distance between the sensor and the max water level
    distance_difference = 0.235

    while True:
        average = sensor.getAverageReading(1000, 0.05)
        distance_from_max = round(average - distance_difference, 2)
        
        if distance_from_max < 0:
            distance_from_max = 0

        query = "INSERT INTO readings (height, time) VALUES (%s, NOW());"
        cursor.execute(query, (distance_from_max,))
        db.commit()  

if __name__ == "__main__":
    logging.basicConfig(filename="sensor.log", level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")

    current_time = time.time()
    allowed_errors_per_minute = 10
    errors = 0
    
    for i in range(allowed_errors_per_minute):
        try:
            main()
        except Exception as e:
            logging.error(e)

            errors += 1

            if (current_time > time.time() - 60):
                current_time = time.time()
                errors = 0

            if errors > allowed_errors_per_minute:
                break

            time.sleep(0.2)

    logging.critical("The error limit has been surpassed")

