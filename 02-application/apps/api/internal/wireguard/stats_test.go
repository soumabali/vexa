package wireguard

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatsCollector(t *testing.T) {
	repo := NewRepository(nil)
	sc := NewStatsCollector(repo)
	assert.NotNil(t, sc)
	assert.False(t, sc.IsRunning())
}

func TestStatsCollector_StartStop(t *testing.T) {
	repo := NewRepository(nil)
	sc := NewStatsCollector(repo)

	sc.Start()
	assert.True(t, sc.IsRunning())

	sc.Start()
	assert.True(t, sc.IsRunning())

	sc.Stop()
	assert.False(t, sc.IsRunning())

	sc.Stop()
	assert.False(t, sc.IsRunning())
}

func TestStatsCollector_SetInterval(t *testing.T) {
	repo := NewRepository(nil)
	sc := NewStatsCollector(repo)

	sc.SetInterval(5 * time.Second)
	sc.Start()
	sc.Stop()
}

func TestStatsCollector_Collect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	sc := NewStatsCollector(repo)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(1000, 0).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()).
			AddRow(uuid.New(), uuid.New(), uuid.New(), "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
				"connected", true, nil, int64(100), int64(200),
				nil, now, now))

	sc.collect()

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStatsCollector_Collect_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	sc := NewStatsCollector(repo)

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(1000, 0).
		WillReturnError(assert.AnError)

	sc.collect()

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStatsCollector_Collect_AllDisabled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	sc := NewStatsCollector(repo)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(1000, 0).
		WillReturnRows(sqlmock.NewRows(tunnelColumns()).
			AddRow(uuid.New(), uuid.New(), uuid.New(), "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
				"disabled", false, nil, int64(0), int64(0),
				nil, now, now))

	sc.collect()

	assert.NoError(t, mock.ExpectationsWereMet())
}
