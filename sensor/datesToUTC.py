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

#243
id = 128
last_id = 146
#last_id = 11016
while id <= last_id:
    print(id)
    c.execute(f"SELECT time FROM readings WHERE ID={id}")
    result = c.fetchone()[0]

    result = result - datetime.timedelta(hours=1)

    if result != None:
        c.execute(f"UPDATE readings SET time=%s WHERE ID=%s", (result, id))
        db.commit()

    id += 1

    