package mini_orm

import (
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
)

// DB sql driver that support master and slaves
type DB struct {
	master   *sql.DB
	slaves   []*sql.DB
	nextIdx  uint64
	EnableMS bool
}

func formatDSN(dsn string) string {
	if strings.Index(dsn, "parseTime=true") <= 0 {
		return fmt.Sprintf("%s&%s", dsn, "parseTime=true")
	}
	return dsn
}

// Open return DB instance
func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, formatDSN(dataSourceName))
	if err != nil {
		return nil, err
	}
	return &DB{master: db, slaves: nil, EnableMS: false}, nil
}

// OpenMasterAndSlaves return DB instance
func OpenMasterAndSlaves(driverName, master string, slaves []string) (*DB, error) {
	mdb, err := sql.Open(driverName, formatDSN(master))
	if err != nil {
		return nil, err
	}
	sdbs := make([]*sql.DB, 0)
	for _, s := range slaves {
		sdb, err := sql.Open(driverName, formatDSN(s))
		if err != nil {
			return nil, err
		}
		sdbs = append(sdbs, sdb)
	}
	return &DB{master: mdb, slaves: sdbs, EnableMS: true}, nil
}

// SetMaxIdleConns set max idle conns
func (db *DB) SetMaxIdleConns(n int) {
	db.master.SetMaxIdleConns(n)
	if db.EnableMS {
		for _, s := range db.slaves {
			s.SetMaxIdleConns(n)
		}
	}
}

// SetMaxOpenConns set max idle conns
func (db *DB) SetMaxOpenConns(n int) {
	db.master.SetMaxOpenConns(n)
	if db.EnableMS {
		for _, s := range db.slaves {
			s.SetMaxOpenConns(n)
		}
	}
}

// Master return master
func (db *DB) Master() *sql.DB {
	return db.master
}

// Slave return slave
func (db *DB) Slave() *sql.DB {
	if db.EnableMS {
		slaveNum := uint64(len(db.slaves))
		if slaveNum == 0 {
			return db.master
		}
		return db.slaves[atomic.AddUint64(&db.nextIdx, 1)%slaveNum]
	}
	return db.Master()
}

// Close impl Conn close method
func (db *DB) Close() error {
	err := db.master.Close()
	if err != nil {
		return err
	}
	if db.EnableMS {
		for _, s := range db.slaves {
			err := s.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Begin impl Conn Begin method
func (db *DB) Begin() (*sql.Tx, error) {
	return db.Master().Begin()
}

// Prepare impl Conn Prepare method
func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	if db.EnableMS {
		return db.Slave().Prepare(query)
	}
	return db.Master().Prepare(query)
}
