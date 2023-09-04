FROM python:3.12.0rc1-bullseye

WORKDIR /app

COPY sensor.py /app/sensor.py

RUN pip3 install RPi.GPIO
RUN pip3 install mysql.connector.python
CMD [ "python3", "sensor.py" ]