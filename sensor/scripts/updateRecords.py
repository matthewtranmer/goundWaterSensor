import mysql.connector, datetime

def connect_db(db_password):
    db = mysql.connector.connect(
        host="localhost",
        user="WorkerRW",
        password=db_password,
        database="sensor"
    )

    return db

db_password = "debugdevuserpassword"

db = connect_db(db_password)
c = db.cursor()

calibration = {
    "Max_Distance":
    ""
}

#43343
for i in range(43343):
    c.execute(f"SELECT height FROM old_readings LIMIT 1 OFFSET {i}")
    height = c.fetchone()[0]

    percentage = round(height / calibration["Max_Distance"] * 100, 0)

    if result != None:
        c.execute(f"INSERT INTO readings ({height}, {percentage})")
        db.commit()

while id <= last_id:
    print(id)
    

    id += 1

    