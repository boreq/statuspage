package db

import (
	"crypto/rand"
	"github.com/boreq/errors"
	"github.com/boreq/statuspage-backend/monitor"
	"github.com/boreq/statuspage-backend/query"
	"github.com/dgraph-io/badger"
	"github.com/oklog/ulid/v2"
	"strings"
	"time"
)

const delimiter = "-"

type MeasurementsStorage struct {
	db *badger.DB
}

func NewMeasurementsStorage(db *badger.DB) *MeasurementsStorage {
	return &MeasurementsStorage{db: db}
}

func (s *MeasurementsStorage) Add(m monitor.Measurement) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return s.add(txn, m)
	})
}

func (s *MeasurementsStorage) add(txn *badger.Txn, m monitor.Measurement) error {
	persistedStatus, err := encodeStatus(m.Status())
	if err != nil {
		return errors.Wrap(err, "error encoding status")
	}

	persistedMeasurement := &PersistedMeasurement{
		Id:        m.Id(),
		Timestamp: m.Timestamp().Unix(),
		Duration:  m.Duration().Seconds(),
		Status:    persistedStatus,
		Output:    m.Output(),
	}

	b, err := persistedMeasurement.MarshalMsg(nil)
	if err != nil {
		return errors.Wrap(err, "error marshaling persisted measurement")
	}

	measurementULID, err := ulid.New(uint64(m.Timestamp().UnixMilli()), rand.Reader)
	if err != nil {
		return errors.Wrap(err, "error creating an ulid")
	}

	if err := txn.Set(makeKey(m.Id(), measurementULID.Bytes()), b); err != nil {
		return errors.Wrap(err, "error calling set")
	}

	return nil
}

func (s *MeasurementsStorage) Last(id string) (monitor.Measurement, error) {
	var result monitor.Measurement
	if err := s.db.View(func(txn *badger.Txn) error {
		tmp, err := s.last(txn, id)
		if err != nil {
			return errors.Wrap(err, "error calling last")
		}
		result = tmp
		return nil
	}); err != nil {
		return monitor.Measurement{}, errors.Wrap(err, "transaction failed")
	}
	return result, nil
}

func (s *MeasurementsStorage) Get(id string, start, end query.Date) ([]monitor.Measurement, error) {
	var result []monitor.Measurement
	if err := s.db.View(func(txn *badger.Txn) error {
		tmp, err := s.get(txn, id, start, end)
		if err != nil {
			return errors.Wrap(err, "error calling get")
		}
		result = tmp
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}
	return result, nil
}

func (s *MeasurementsStorage) get(txn *badger.Txn, id string, start, end query.Date) ([]monitor.Measurement, error) {
	startTimestamp := time.Date(start.Year, start.Month, start.Day, 0, 0, 0, 0, time.Local)
	endTimestamp := time.Date(end.Year, end.Month, end.Day, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)

	startULID, err := ulid.New(uint64(startTimestamp.UnixMilli()), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "error making start ULID")
	}

	startKey := makeKey(id, startULID.Bytes()[:6])

	iterator := txn.NewIterator(badger.DefaultIteratorOptions)
	defer iterator.Close()

	var result []monitor.Measurement

	for iterator.Seek(startKey); iterator.Valid(); iterator.Next() {
		measurement, err := s.load(iterator.Item())
		if err != nil {
			return nil, errors.Wrap(err, "error loading measurement")
		}

		if !measurement.Timestamp().Before(endTimestamp) {
			break
		}

		result = append(result, measurement)
	}

	return result, nil
}

func (s *MeasurementsStorage) last(txn *badger.Txn, id string) (monitor.Measurement, error) {
	prefix := makeKey(id, nil)
	seek := makeKey(id, []byte{0xff})

	options := badger.DefaultIteratorOptions
	options.Reverse = true
	options.Prefix = prefix

	iterator := txn.NewIterator(options)
	defer iterator.Close()

	iterator.Seek(seek)
	if !iterator.Valid() {
		return monitor.Measurement{}, query.ErrMeasurementNotFound
	}

	return s.load(iterator.Item())
}

func (s *MeasurementsStorage) load(item *badger.Item) (monitor.Measurement, error) {
	persistedMeasurement := &PersistedMeasurement{}

	if err := item.Value(func(val []byte) error {
		_, err := persistedMeasurement.UnmarshalMsg(val)
		return err
	}); err != nil {
		return monitor.Measurement{}, errors.Wrap(err, "error getting item value")
	}

	t := time.Unix(persistedMeasurement.Timestamp, 0)

	d := time.Duration(persistedMeasurement.Duration) * time.Second

	status, err := decodeStatus(persistedMeasurement.Status)
	if err != nil {
		return monitor.Measurement{}, errors.Wrap(err, "error decoding status")
	}

	measurement, err := monitor.NewMeasurement(
		persistedMeasurement.Id,
		t,
		d,
		status,
		persistedMeasurement.Output,
	)
	if err != nil {
		return monitor.Measurement{}, errors.Wrap(err, "error creating a measurement")
	}

	return measurement, nil
}

func makeKey(id string, b []byte) []byte {
	key := []byte(escapeId(id))
	key = append(key, []byte(delimiter)...)
	key = append(key, b...)
	return key
}

func escapeId(id string) string {
	return strings.ReplaceAll(id, delimiter, "\\"+delimiter)
}
