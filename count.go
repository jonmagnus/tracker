package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

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

func countAfter(ctx context.Context, client *firestore.Client, action string, stime time.Time) {
	var timeslot Timeslot
	iter := client.Collection("times").Where("activity", "==", action).Where("end_time", ">", stime).Documents(ctx)
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

	fmt.Printf("Time spent on \"%v\": %v after %s\n", action, cummTime, stime.Format("2. Jan 2006"))

}
