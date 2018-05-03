package database

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	bolt "github.com/coreos/bbolt"
)

// Store is the database store used by the system
type Store struct {
	db *bolt.DB
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
		err = s.GetBucket(ServiceBucket)

		if err != nil {
			return err
		}
	}

	return nil
}

// CloseDatabase will close the current database connection
func (s *Store) CloseDatabase() {
	if s.db != nil {
		defer s.db.Close()
	}
}

// GetBucket will ensure a bucket exists
func (s *Store) GetBucket(bucketName string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))

		if err != nil {
			return err
		}
		return nil
	})
}

// CreateService will create a new or update a service in the database
func (s *Store) CreateService(u *Service) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Retrieve the services bucket.
		// This should be created when the DB is first opened.
		b := tx.Bucket([]byte(ServiceBucket))

		// Marshal user data into bytes.
		buf, err := json.Marshal(*u)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return b.Put(idToKey(u.ID), buf)
	})
}

// GetServices will return all the services in the database
func (s *Store) GetServices() ([]*Service, error) {
	var services []*Service
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ServiceBucket))

		return b.ForEach(func(k, d []byte) error {

			var service = Service{}

			err := json.Unmarshal(d, &service)

			if err != nil {
				return err
			}

			services = append(services, &service)

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return services, nil
}

// GetService will return the service by id
func (s *Store) GetService(id int64) (*Service, error) {
	var service = &Service{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ServiceBucket))

		d := b.Get(idToKey(id))

		if d == nil {
			return errors.New("No service found")
		}

		err := json.Unmarshal(d, *service)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return service, nil
}

// itob returns an 8-byte big endian representation of v.
func idToKey(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
