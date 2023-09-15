REVOKE ALL PRIVILEGES ON *.* FROM 'WorkerRW'@'%';
GRANT INSERT, SELECT ON *.* TO 'WorkerRW'@'%';

CREATE TABLE IF NOT EXISTS readings(
    ID INT NOT NULL AUTO_INCREMENT,
    PRIMARY KEY(ID),
    height FLOAT(24),
    percentage INT,
    time DATETIME,
    max_distance FLOAT(24),
    min_distance FLOAT(24),
    min_max_uncertanty FLOAT(24)
)