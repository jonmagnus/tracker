package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func listAfter(ctx context.Context, client *firestore.Client, stime time.Time) {
	iter := client.Collection("times").Where("end_time", ">=", stime).Documents(ctx)
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
