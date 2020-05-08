package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
)

func main() {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "go-time-tracker-882c1"}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	flag.Parse()
	if len(flag.Args()) < 1 {
		panic("Too few arguments provided")
	}

	command := flag.Arg(0)

	switch command {
	case "start":
		if len(flag.Args()) < 2 {
			panic("Too few arguments provided for command \"start\"")
		}
		action := flag.Arg(1)
		if len(action) == 0 {
			log.Fatalf("Invalid action: %v", action)
		}
		startAction(ctx, client, action)

	case "stop":
		stopActiveAction(ctx, client)

	case "list":
		if len(flag.Args()) < 2 {
			retrieveTimes(ctx, client)
		} else {
			now := time.Now()
			switch flag.Arg(1) {
			case "day":
				listAfter(ctx, client, time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()))
			case "month":
				listAfter(ctx, client, time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location()))
			default:
				panic(fmt.Sprintf("Invalid list option %s", flag.Arg(1)))
			}
		}

	case "report":
		now := time.Now()
		stime := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		etime := now.AddDate(100, 0, 0)
		if len(flag.Args()) < 2 {
			reportInterval(ctx, client, stime, etime)
		} else {
			var offset int
			if len(flag.Args()) > 1 {
				if _, err := fmt.Sscanf(flag.Arg(1), "%d", &offset); err != nil {
					log.Println(err)
					panic("Offset must be an integer value")
				}
			} else {
				offset = 0
			}
			stime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			stime = stime.AddDate(0, offset, 0)
			etime = stime.AddDate(0, 1, 0)
			reportInterval(ctx, client, stime, etime)
		}

	/*
		case "clean":
			cleanTimes(ctx, client)
	*/

	case "count":
		if len(flag.Args()) < 2 {
			panic("Too few arguments provided for command \"count\"")
		}
		now := time.Now()
		action := flag.Arg(1)
		if len(flag.Args()) < 3 {
			stime := time.Date(
				now.Year(),
				now.Month(),
				now.Day(), 0, 0, 0, 0,
				now.Location(),
			)
			countAfter(ctx, client, action, stime)
		} else {
			var offset int
			if len(flag.Args()) > 3 {
				if _, err := fmt.Sscanf(flag.Arg(3), "%d", &offset); err != nil {
					log.Println(err)
					panic("Offset must be an integer value")
				}
			} else {
				offset = 0
			}
			switch {
			case flag.Arg(2) == "day":
				stime := time.Date(
					now.Year(),
					now.Month(),
					now.Day(), 0, 0, 0, 0,
					now.Location(),
				).AddDate(0, 0, offset)
				countAfter(ctx, client, action, stime)
			case flag.Arg(2) == "month":
				stime := time.Date(
					now.Year(),
					now.Month(),
					0, 0, 0, 0, 0,
					now.Location(),
				).AddDate(0, offset, 0)
				countAfter(ctx, client, action, stime)
			case flag.Arg(2) == "all":
				countTime(ctx, client, action)
			default:
				panic("Invalid argument")
			}
		}

	case "delete":
		if len(flag.Args()) < 2 {
			panic("Too few arguments provided for command \"count\"")
		}
		action := flag.Arg(1)
		deleteAction(ctx, client, action)

	default:
		flag.Usage()
		panic("Invalid command")
	}
}

func cleanTimes(ctx context.Context, client *firestore.Client) {
	iter := client.Collection("times").Documents(ctx)
	batch := client.Batch()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to get document: %v", err)
		}

		batch.Delete(doc.Ref)
	}

	_, err := batch.Commit(ctx)
	if err != nil {
		log.Fatalf("Failed to commit batch %v", err)
	}
}

func retrieveTimes(ctx context.Context, client *firestore.Client) {
	iter := client.Collection("times").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to retrieve collection: %v", err)
		}

		var timeslot Timeslot
		err = doc.DataTo(&timeslot)
		if err != nil {
			log.Fatalf("Failed to write data to timeslot: %v", err)
		}
		fmt.Println(timeslot)
	}
}

func stopActiveAction(ctx context.Context, client *firestore.Client) {
	var timeslot Timeslot
	doc, err := client.Collection("times").Doc("active_action").Get(ctx)
	if err != nil {
		log.Fatalf("Failed to retrieve active action: %v", err)
	}
	_, err = client.Collection("times").Doc("active_action").Delete(ctx)
	if err != nil {
		log.Fatalf("Failed to delete active_time: %v", err)
	}

	err = doc.DataTo(&timeslot)
	if err != nil {
		log.Fatalf("Failed to write to timeslot: %v", err)
	}
	timeslot.EndTime = time.Now()

	_, _, err = client.Collection("times").Add(ctx, timeslot)
	if err != nil {
		log.Fatalf("Failed to add active action: %v", err)
	}
}

func startAction(ctx context.Context, client *firestore.Client, action string) {
	stopActiveAction(ctx, client)

	timeslot := Timeslot{
		StartTime: time.Now(),
		Activity:  action,
	}
	_, err := client.Collection("times").Doc("active_action").Set(ctx, timeslot)
	if err != nil {
		log.Fatalf("Failed to set active action: %v", err)
	}
}

func deleteAction(ctx context.Context, client *firestore.Client, action string) {
	iter := client.Collection("times").Where("activity", "==", action).Documents((ctx))
	batch := client.Batch()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read timeslot: %v", err)
		}

		batch.Delete(doc.Ref)
	}

	_, err := batch.Commit(ctx)
	if err != nil {
		log.Fatalf("Failed to commit batch deletion: %v", err)
	}
}
