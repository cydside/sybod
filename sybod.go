package main

//______________________________________________________________________________

import (
	"fmt"
	"io/ioutil"
	"os"

	bolt "go.etcd.io/bbolt"
)

//______________________________________________________________________________

// Entry Model for Keys and Values bucket's data
type Entry struct {
	Key   []byte
	Value []byte
}

//______________________________________________________________________________

// Container Model for BoltDB's buckets structure
type Container struct {
	Name       []byte
	SubBuckets []Container
	Entries    []Entry
}

//______________________________________________________________________________

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("sybod (ver 1.0.3): No db file provided, bye!")
		os.Exit(1)
	}

	fileName := args[1]
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		fmt.Println("Can't open source db", err)
		os.Exit(1)
	}

	destFileName := "newcopy_" + fileName
	if len(args) == 3 {
		if isValid(args[2]) {
			destFileName = args[2]
		} else {
			fmt.Println("Destination file name not valid, changed to default: ", destFileName)
		}
	}

	data := dump(db)
	pour(data, destFileName)

	err = db.Close()
	if err != nil {
		fmt.Println("Can't close source db", err)
		os.Exit(1)
	}

}

//______________________________________________________________________________

func dump(db *bolt.DB) *Container {
	data := new(Container)

	db.View(func(tx *bolt.Tx) error {
		tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			cb := new(Container)
			cb.Name = name
			readBucket(cb, b)
			data.SubBuckets = append(data.SubBuckets, *cb)

			return nil
		})
		return nil
	})

	// fmt.Printf("Database:\n%v\n\n", data)
	// fmt.Printf("Database:\n%#v\n\n", data)

	return data
}

//______________________________________________________________________________

func readBucket(bkt *Container, b *bolt.Bucket) {
	b.ForEach(func(k, v []byte) error {
		if subB := b.Bucket(k); subB != nil {
			sb := new(Container)
			sb.Name = k
			readBucket(sb, subB)

			bkt.SubBuckets = append(bkt.SubBuckets, *sb)

			return nil
		}

		et := new(Entry)
		et.Key = k
		et.Value = v
		bkt.Entries = append(bkt.Entries, *et)

		return nil
	})

	return
}

//______________________________________________________________________________

func pour(bkt *Container, destFileName string) {
	db, err := bolt.Open(destFileName, 0600, nil)
	if err != nil {
		fmt.Println("Can't open destination db", err)
		os.Exit(1)
	}

	var bktPath []string
	copyTree(db, bkt, bktPath)

	err = db.Close()
	if err != nil {
		fmt.Println("Can't close destination db", err)
		os.Exit(1)
	}
}

//______________________________________________________________________________

func copyTree(db *bolt.DB, bkt *Container, bktPath []string) {
	var err error

	for _, cb := range bkt.SubBuckets {
		bktPath = append(bktPath, string(cb.Name))
		fmt.Printf("Bucket path: %v\nEntries: %d\n\n", bktPath, len(cb.Entries))

		if cb.SubBuckets != nil {
			copyTree(db, &cb, bktPath)
		}

		err = makeBucket(db, &bktPath)
		if err != nil {
			fmt.Printf("Can't create bucket %s: %s", string(cb.Name), err)
			os.Exit(1)
		}

		err = insertEntry(db, cb.Entries, &bktPath)
		if err != nil {
			fmt.Printf("Can't insert into bucket %s: %s", string(bkt.Name), err)
			os.Exit(1)
		}

		bktPath = bktPath[:len(bktPath)-1]
	}
}

//______________________________________________________________________________

func makeBucket(db *bolt.DB, bktPath *[]string) error {
	var err error
	var b *bolt.Bucket

	err = db.Update(func(tx *bolt.Tx) (err error) {
		for _, bktName := range *bktPath {
			if b != nil {
				if b, err = b.CreateBucketIfNotExists([]byte(bktName)); err != nil {
					return err
				}

			} else {
				if b, err = tx.CreateBucketIfNotExists([]byte(bktName)); err != nil {
					return err
				}
			}
		}

		return err
	})

	return err
}

//______________________________________________________________________________

func getBucket(tx *bolt.Tx, bktPath *[]string) *bolt.Bucket {
	var b *bolt.Bucket

	for _, bktName := range *bktPath {
		if b != nil {
			if b = b.Bucket([]byte(bktName)); b == nil {
				return nil
			}
		} else {
			if b = tx.Bucket([]byte(bktName)); b == nil {
				return nil
			}
		}
	}

	return b
}

//______________________________________________________________________________

func insertEntry(db *bolt.DB, entry []Entry, bktPath *[]string) error {
	var err error

	for _, et := range entry {
		err = db.Update(func(tx *bolt.Tx) error {
			var b = getBucket(tx, bktPath)
			err = b.Put(et.Key, et.Value)
			return err
		})
		if err != nil {
			return err
		}
	}

	return err
}

//______________________________________________________________________________

func isValid(fp string) bool {
	fmt.Println("isValid: ", fp)
	// Check if file already exists
	if _, err := os.Stat(fp); err == nil {
		return true
	}

	fmt.Println("isValid")
	// Attempt to create it
	var d []byte
	if err := ioutil.WriteFile(fp, d, 0644); err == nil {
		os.Remove(fp) // And delete it
		return true
	}

	fmt.Println("isValid")
	return false
}
