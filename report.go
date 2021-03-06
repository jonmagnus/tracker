package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func reportInterval(ctx context.Context, client *firestore.Client, stime time.Time, etime time.Time) {
	iter := client.Collection("times").Where("end_time", ">=", stime).Where("end_time", "<=", etime).Documents(ctx)
	days := make(map[string]time.Duration)
	activities := make(map[string]time.Duration)

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

		key := timeslot.EndTime.Format("2006/01/02") + " " + timeslot.Activity
		if duration, ok := days[key]; ok {
			days[key] = duration + timeslot.Duration()
		} else {
			days[key] = timeslot.Duration()
		}

		if duration, ok := activities[timeslot.Activity]; ok {
			activities[timeslot.Activity] = duration + timeslot.Duration()
		} else {
			activities[timeslot.Activity] = timeslot.Duration()
		}
	}

	// Check if we are currently tracking
	doc, err := client.Collection("times").Doc("active_action").Get(ctx)
	if status.Code(err) != codes.NotFound || err == nil {
		var timeslot Timeslot
		err = doc.DataTo(&timeslot)
		if err != nil {
			log.Fatalf("Failed to parse active action")
		}
		key := time.Now().Format("2006/01/02") + " " + timeslot.Activity
		timeslot.EndTime = time.Now()
		fmt.Printf("Current activity: %s - %v\n", timeslot.Activity, timeslot.Duration())

		if duration, ok := days[key]; ok {
			days[key] = duration + timeslot.Duration()
		} else {
			days[key] = timeslot.Duration()
		}

		if duration, ok := activities[timeslot.Activity]; ok {
			activities[timeslot.Activity] = duration + timeslot.Duration()
		} else {
			activities[timeslot.Activity] = timeslot.Duration()
		}
	}

	keys := make([]string, 0)
	for k := range days {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %v\n", k, days[k])
	}
	fmt.Println("----------")
	for k, duration := range activities {
		fmt.Printf("%s: %v\n", k, duration)
	}
}
