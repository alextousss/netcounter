package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB

func increment(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["key"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}

	key := keys[0]

	log.Println("Url Param 'key' is: " + string(key))

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		rawVar := string(b.Get([]byte(key)))
		var initialVar int
		if rawVar == "" {
			initialVar = 0
		} else {
			var err error
			initialVar, err = strconv.Atoi(rawVar)
			if err != nil {
				return err
			}
		}
		fmt.Println(key, initialVar+1)
		return b.Put([]byte(key), []byte(strconv.Itoa(initialVar+1)))
	})
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte("ok"))
}

func getCounters(w http.ResponseWriter, r *http.Request) {
	var res string
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("MyBucket"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			res += string(k) + " : " + string(v) + "\n"
		}

		return nil
	})

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(res))
}

func main() {
	var err error
	db, err = bolt.Open("counter.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		fmt.Println("bucket created")
		return nil
	})
	defer db.Close()
	http.HandleFunc("/increment/", increment)
	http.HandleFunc("/getcounters/", getCounters)
	log.Fatal(http.ListenAndServe(":3001", nil))

}
