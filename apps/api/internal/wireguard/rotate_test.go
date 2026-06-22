package wireguard

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRotationScheduler(t *testing.T) {
	repo := NewRepository(nil)
	ifaceCtrl := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	rs := NewRotationScheduler(repo, ifaceCtrl, 90)

	assert.NotNil(t, rs)
	assert.False(t, rs.IsRunning())
}

func TestRotationScheduler_StartStop(t *testing.T) {
	repo := NewRepository(nil)
	ifaceCtrl := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	rs := NewRotationScheduler(repo, ifaceCtrl, 90)

	rs.Start()
	assert.True(t, rs.IsRunning())

	rs.Start()
	assert.True(t, rs.IsRunning())

	rs.Stop()
	assert.False(t, rs.IsRunning())

	rs.Stop()
	assert.False(t, rs.IsRunning())
}

func TestRotationScheduler_SetInterval(t *testing.T) {
	repo := NewRepository(nil)
	ifaceCtrl := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	rs := NewRotationScheduler(repo, ifaceCtrl, 90)

	rs.SetInterval(10 * time.Second)
	rs.Start()
	rs.Stop()
}

func TestRotateTunnelKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	ifaceCtrl := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	rs := NewRotationScheduler(repo, ifaceCtrl, 90)

	tunnelID := uuid.New()

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = rs.rotateTunnelKeys(context.Background(), tunnelID.String())
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateTunnelKeys_InvalidID(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	ifaceCtrl := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	rs := NewRotationScheduler(repo, ifaceCtrl, 90)

	err = rs.rotateTunnelKeys(context.Background(), "not-a-uuid")
	assert.Error(t, err)
}

func TestRotateDue_WithTunnels(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	ifaceCtrl := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	rs := NewRotationScheduler(repo, ifaceCtrl, 90)

	tunnelID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE is_enabled = true`).
		WithArgs("90").
		WillReturnRows(sqlmock.NewRows(tunnelColumns()).
			AddRow(tunnelID, userID, uuid.New(), "wg0",
				"priv", "pub", "cpub", nil, "10.200.200.1/24", "10.200.200.2/24",
				51820, []byte("{0.0.0.0/0}"), []byte("{}"), 1420, 25,
				"connected", true, nil, int64(100), int64(200),
				nil, now, now))

	mock.ExpectExec(`UPDATE wireguard_tunnels SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rs.rotateDue()

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateDue_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	ifaceCtrl := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	rs := NewRotationScheduler(repo, ifaceCtrl, 90)

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE is_enabled = true`).
		WithArgs("90").
		WillReturnRows(sqlmock.NewRows(tunnelColumns()))

	rs.rotateDue()

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRotateDue_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	ifaceCtrl := NewInterfaceController("/nonexistent/wg", "10.200.200.0/24")
	rs := NewRotationScheduler(repo, ifaceCtrl, 90)

	mock.ExpectQuery(`SELECT .+ FROM wireguard_tunnels WHERE is_enabled = true`).
		WithArgs("90").
		WillReturnError(assert.AnError)

	rs.rotateDue()

	assert.NoError(t, mock.ExpectationsWereMet())
}
