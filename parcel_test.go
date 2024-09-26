package main

import (
	"database/sql"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

var db *sql.DB

func TestMain(m *testing.M) {
	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	);`)
	if err != nil {
		log.Fatal(err)
	}

	m.Run()
}

func TestAddGetDelete(t *testing.T) {
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)

	retrievedParcel.Number = id

	require.Equal(t, parcel, retrievedParcel)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, retrievedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, retrievedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	store := NewParcelStore(db)

	parcels := []Parcel{
		{
			Client:    randRange.Intn(10_000_000),
			Status:    ParcelStatusRegistered,
			Address:   "test address 1",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		{
			Client:    randRange.Intn(10_000_000),
			Status:    ParcelStatusRegistered,
			Address:   "test address 2",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		{
			Client:    randRange.Intn(10_000_000),
			Status:    ParcelStatusRegistered,
			Address:   "test address 3",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	parcelMap := map[int]Parcel{}

	for i := range parcels {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	client := parcels[0].Client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	require.Equal(t, 1, len(storedParcels))

	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists)

		expectedParcel.Number = parcel.Number

		require.Equal(t, expectedParcel, parcel)
	}
}
