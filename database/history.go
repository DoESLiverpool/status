package database

import (
	"encoding/json"
	"errors"
	"time"

	bolt "github.com/coreos/bbolt"

	"github.com/google/uuid"
)

// History is an event that happens when a service changes state
type History struct {
	ID uuid.UUID `json:"id" binding:"required"`

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

		history.ID = uuid.New()

		// Marshal user data into bytes.
		buf, err := json.Marshal(*history)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return b.Put([]byte(history.ID.String()), buf)
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
	b := tx.Bucket([]byte(HistoryBucket))

	if b != nil && tx.Writable() {
		b, err := b.CreateBucketIfNotExists(idToKey(serviceID))

		if err != nil {
			return nil, err
		}

		return b, nil
	}

	return nil, errors.New("No bucket found")
}
