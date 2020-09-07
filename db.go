package mini_orm

import "context"

// Config connection
type Config struct {
	Driver       string
	MasterAddr   string
	SlavesAddr   []string
	MaxIdleConns int
	MaxOpenConns int
}

// Engine orm engine define
type Engine struct {
	*DB
	*Config
}

// NewEngine return engine
func NewEngine(driverName, dataSourceName string) (*Engine, error) {
	e := &Engine{nil, nil}
	db, err := Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	e.DB = db
	return e, nil
}

// NewEngineWithMS return engine with master and slaves
func NewEngineWithMS(driverName, masterAddr string, slavesAddr []string) (*Engine, error) {
	e := &Engine{nil, nil}
	db, err := OpenMasterAndSlaves(driverName, masterAddr, slavesAddr)
	if err != nil {
		return nil, err
	}
	e.DB = db
	return e, nil
}

// New return engine instance
func New(cfg *Config) (*Engine, error) {
	if cfg == nil {
		return nil, CFBNotAllowEmpty
	}
	e := &Engine{nil, cfg}

	if cfg.SlavesAddr != nil && len(cfg.SlavesAddr) != 0 {
		Tracef("[New Engine] driver: %s, masterAddr: %s, slaveAddr: %v", cfg.Driver, cfg.MasterAddr, cfg.SlavesAddr)
		db, err := OpenMasterAndSlaves(cfg.Driver, cfg.MasterAddr, cfg.SlavesAddr)
		if err != nil {
			return nil, err
		}
		e.DB = db
	} else {
		Tracef("[New Engine] driver: %s, masterAddr: %s", cfg.Driver, cfg.MasterAddr)
		db, err := Open(cfg.Driver, cfg.MasterAddr)
		if err != nil {
			return nil, err
		}
		e.DB = db
	}
	e.SetMaxIdleConns(cfg.MaxIdleConns)
	e.SetMaxOpenConns(cfg.MaxOpenConns)
	return e, nil
}

// NewSessionCtx return new session instance with ctx
func (e *Engine) NewSessionCtx(ctx context.Context) *Session {
	return &Session{
		db:                     e.DB,
		ctx:                    ctx,
		statement:              &Statement{},
		isAutoCommit:           true,
		hasCommittedOrRollback: false,
		tx:                     nil,
	}
}

// NewSession return new session instance
func (e *Engine) NewSession() *Session {
	return e.NewSessionCtx(nil)
}

// Close Engine close
func (e *Engine) Close() {
	e.DB.Close()
}
