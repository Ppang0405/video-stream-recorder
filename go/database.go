package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.etcd.io/bbolt"
)

type DatabaseItem struct {
	ID   uint64
	Name string
	Len  float64
	T    int64
}

var (
	database *bbolt.DB
	dbbucket = []byte("main")
)

func database_init() {
	var err error
	database, err = bbolt.Open("./db.db", 0755, nil)
	if err != nil {
		log.Fatalf("Open database error: %v", err)
	}
	database.Update(func(tx *bbolt.Tx) error {
		tx.CreateBucketIfNotExists(dbbucket)
		return nil
	})
}

func database_store(item *DatabaseItem) {

	database.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbbucket)
		ID, _ := bucket.NextSequence()
		item.ID = ID
		encoded, _ := json.Marshal(item)
		bucket.Put(itob(int(item.T)), encoded)
		return nil

	})
}

func database_last_5() []*DatabaseItem {

	var rt []*DatabaseItem
	var tmp []*DatabaseItem

	database.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbbucket)
		cursor := bucket.Cursor()

		n := 0
		for k, v := cursor.Last(); k != nil; k, v = cursor.Prev() {
			if n >= 7 {
				break
			}
			n++
			var item *DatabaseItem
			err := json.Unmarshal(v, &item)
			if err != nil {
				log.Printf("error %v", err)
				continue
			}
			tmp = append(tmp, item)

		}
		return nil
	})

	for i := len(tmp) - 1; i > 0; i-- {
		rt = append(rt, tmp[i])
	}

	return rt
}

func database_get(s, l string) []*DatabaseItem {

	si, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("database_get error: %v", err)
		return nil
	}
	li, err := strconv.Atoi(l)
	if err != nil {
		log.Printf("database_get error: %v", err)
		return nil
	}

	var rt []*DatabaseItem

	database.View(func(tx *bbolt.Tx) error {
		start := itob(si * 1000000000)
		end := itob(si + li*60*1000000000)

		bucket := tx.Bucket(dbbucket)
		cursor := bucket.Cursor()

		for k, v := cursor.Seek(start); k != nil; k, v = cursor.Next() {
			if bytes.Compare(k, end) < 0 {
				return nil
			}
			var item *DatabaseItem
			err := json.Unmarshal(v, &item)
			if err != nil {
				log.Printf("error %v", err)
				continue
			}
			rt = append(rt, item)

		}
		return nil
	})

	return rt
}

func database_worker() {
	for {
		current := time.Now().UnixNano() - int64(*flagTail*60*60*int(time.Nanosecond))
		err := database.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket(dbbucket)
			cursor := bucket.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				_ = v
				if bytes.Compare(k, itob(int(current))) < 0 {
					return nil
				}
				var item DatabaseItem
				err := json.Unmarshal(v, &item)
				if err != nil {
					return err
				}
				log.Printf("removing ./files/%v", item.Name)
				os.RemoveAll(fmt.Sprintf("./files/%v", item.Name))
				bucket.Delete(k)

			}
			return nil
		})
		if err != nil {
			log.Printf("error %v", err)
		}
		time.Sleep(time.Second)
	}
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// database_count_segments returns the total number of segments in the database
func database_count_segments() int {
	var count int
	
	database.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("segments"))
		if b == nil {
			return nil
		}
		
		stats := b.Stats()
		count = stats.KeyN
		return nil
	})
	
	return count
}
