package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
)

type Timeslot struct {
	StartTime time.Time `firestore:"start_time"`
	EndTime   time.Time `firestore:"end_time"`
	Activity  string    `firestore:"activity"`
}

func (t Timeslot) String() string {
	return fmt.Sprintf("%v - %v: %v", t.StartTime, t.EndTime, t.Activity)
}

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
	fmt.Printf("Number of unused arguments: %v\n", flag.NArg())
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
		retrieveTimes(ctx, client)

	case "clean":
		cleanTimes(ctx, client)

	case "count":
		if len(flag.Args()) < 2 {
			panic("Too few arguments provided for command \"count\"")
		}
		action := flag.Arg(1)
		countTime(ctx, client, action)

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
		return
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

func countTime(ctx context.Context, client *firestore.Client, action string) {
	var timeslot Timeslot
	iter := client.Collection("times").Where("activity", "==", action).Documents(ctx)
	cummTime := time.Duration(0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read activity: %v", err)
		}

		doc.DataTo(&timeslot)
		if timeslot.EndTime.Before(timeslot.StartTime) {
			continue
		}
		cummTime += timeslot.EndTime.Sub(timeslot.StartTime)
	}

	fmt.Printf("Time spent on \"%v\": %v\n", action, cummTime)
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
