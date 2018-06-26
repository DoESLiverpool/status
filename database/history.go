package database

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	bolt "github.com/coreos/bbolt"
)

// History is an event that happens when a service changes state
type History struct {
	ID int64 `json:"id" binding:"required"`

	OldTimestamp time.Time `json:"old_timestamp" binding:"required"`
	NewTimestamp time.Time `json:"new_timestamp" binding:"required"`

	ChangedFrom ServiceState `json:"changed_from" binding:"required"`
	ChangedTo   ServiceState `json:"changed_to" binding:"required"`

	ServiceID int64 `json:"service_id" binding:"required"`
}

// HistoryBucket is the name of the bucket that contains all the history events
const HistoryBucket = "history"

// HistoryHelper is a helper for all history methods
type HistoryHelper struct {
	store *Store
}

// CreateHistoryEvent will create a new event in the
func (h *HistoryHelper) CreateHistoryEvent(history *History) error {
	return h.store.db.Update(func(tx *bolt.Tx) error {
		b, err := h.getHistoryBucketForService(tx, history.ServiceID)

		if err != nil {
			return err
		}

		id, _ := b.NextSequence()
		history.ID = int64(id)

		// Marshal history data into bytes.
		buf, err := json.Marshal(*history)
		if err != nil {
			return err
		}

		idString := strconv.FormatInt(history.ID, 10)

		// Persist bytes into service id bucket.
		return b.Put([]byte(idString), buf)
	})
}

// GetHistoryForService will return all the history events for this service id
func (h *HistoryHelper) GetHistoryForService(serviceID int64) ([]*History, error) {
	var events []*History
	err := h.store.db.View(func(tx *bolt.Tx) error {
		b, err := h.getHistoryBucketForService(tx, serviceID)

		if err != nil {
			return err
		}

		return b.ForEach(func(k, d []byte) error {

			var event = History{}

			err := json.Unmarshal(d, &event)

			if err != nil {
				return err
			}

			events = append(events, &event)

			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return events, nil
}

func (h *HistoryHelper) getHistoryBucketForService(tx *bolt.Tx, serviceID int64) (*bolt.Bucket, error) {
	historyBucket := tx.Bucket([]byte(HistoryBucket))
	id := strconv.FormatInt(serviceID, 10)

	if historyBucket != nil && tx.Writable() {
		innerBucket, err := historyBucket.CreateBucketIfNotExists([]byte(id))

		if err != nil {
			return nil, err
		}

		return innerBucket, nil
	} else if historyBucket != nil {
		innerBucket := historyBucket.Bucket([]byte(id))

		if innerBucket != nil {
			return innerBucket, nil
		}
	}

	return nil, errors.New("No bucket found")
}
