package database

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"time"

	bolt "github.com/coreos/bbolt"
)

// Store is the database store used by the system
type Store struct {
	db      *bolt.DB
	Service *ServiceHelper
}

// GetDatabase will connect to the database and ensure buckets exist
func (s *Store) GetDatabase(readOnly bool) error {
	path := os.Getenv("DATABASE_PATH")

	if path == "" {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

		if err != nil {
			return err
		}

		path = dir
	}

	if path[len(path)-1:] != "/" {
		path = path + "/"
	}

	var err error
	// Try to open the database at the set path. If the file is locked the system will timeout connecting to the db
	s.db, err = bolt.Open(path+"status.db", 0600, &bolt.Options{Timeout: 1 * time.Second, ReadOnly: readOnly})

	if err != nil {
		return err
	}

	if !readOnly {
		err = s.CreateBucket(ServiceBucket)

		if err != nil {
			return err
		}
	}

	// Create the service helper
	s.Service = &ServiceHelper{}
	s.Service.store = s

	return nil
}

// CloseDatabase will close the current database connection
func (s *Store) CloseDatabase() {
	if s.db != nil {
		defer s.db.Close()
	}
}

// CreateBucket will ensure a bucket exists
func (s *Store) CreateBucket(bucketName string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))

		if err != nil {
			return err
		}
		return nil
	})
}

// itob returns an 8-byte big endian representation of v.
func idToKey(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
