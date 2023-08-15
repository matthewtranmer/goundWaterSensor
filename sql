CREATE TABLE readings(
    ID int NOT NULL AUTO_INCREMENT,
    PRIMARY KEY(ID),

    height float(24),
    date DATETIME
)

SELECT t.time, t.height FROM (
    SELECT height, time, ROW_NUMBER() OVER (ORDER BY time) AS rownum
    FROM readings
) AS t
WHERE t.rownum % 1 = 0 
ORDER BY t.time

SELECT t.time, t.height FROM (
	SELECT height, time, ROW_NUMBER() OVER (ORDER BY time) AS rownum
	FROM readings
) AS t
WHERE t.rownum % ? = 0 AND t.time >= ? AND t.time <= ?
ORDER BY t.time

SELECT t.time, t.height FROM (
	SELECT height, time, ROW_NUMBER() OVER (ORDER BY time) AS rownum
	FROM readings
) AS t
WHERE t.rownum % 1 = 0 AND t.time >= "2023-08-14 17:00:00" AND t.time <= "2023-08-15 17:00:00"
ORDER BY t.time
