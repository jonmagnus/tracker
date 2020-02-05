package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type Date struct {
	year, month, day int
}

func (date Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", date.year, date.month, date.day)
}

type Time struct {
	hours, minutes, seconds int
}

func (time Time) String() string {
	return fmt.Sprintf("%02d:%02d:%02d", time.hours, time.minutes, time.seconds)
}

type Moment struct {
	date Date
	time Time
}

func (m Moment) String() string {
	return fmt.Sprintf("(%v, %v)", m.date, m.time)
}

func cleanDB(db *sql.DB) {
	db.Exec(`DROP TABLE times`)
	db.Exec(`DROP TYPE moment`)
}

func initDB(db *sql.DB) {
	var err error
	_, err = db.Exec(`CREATE TYPE moment AS (date DATE, time TIME)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE times (start_moment moment, ent_moment moment, activity TEXT)`)
	if err != nil {
		log.Fatal(err)
	}
}

func populateDB(db *sql.DB) {
	_, err := db.Exec(
		`INSERT INTO times VALUES ($1, $2, $3)`,
		Moment{Date{2000, 2, 14}, Time{12, 0, 0}}.String(),
		Moment{Date{2000, 2, 15}, Time{12, 0, 0}}.String(),
		"Birth",
	)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	connStr := "user=tracker password=tracker dbname=mydb"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Ping failed:", err)
	}

	cleanDB(db)
	initDB(db)
	fmt.Println(Moment{Date{2000, 2, 14}, Time{12, 0, 0}})
	populateDB(db)

	var start, end, activity string

	row, err := db.Query(`select * from times`)
	if err != nil {
		log.Fatal(err)
	}
	for row.Next() {
		row.Scan(&start, &end, &activity)
		fmt.Println(start, end, activity)
	}

	t := time.Now()
	fmt.Println(t)
}
