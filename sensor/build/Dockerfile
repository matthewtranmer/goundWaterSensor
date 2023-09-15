FROM python

WORKDIR /app

COPY sensor.py /app/sensor.py
COPY calibration.json /app/calibration.json

RUN pip3 install RPi.GPIO
RUN pip3 install mysql.connector.python
CMD [ "python3", "sensor.py" ]