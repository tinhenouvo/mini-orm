package mini_orm

import (
	"errors"
)

var (
	CFBNotAllowEmpty            = errors.New("config not allow empty")
	StatementTableNotSet        = errors.New("statement table not set")
	StatementTypeNotSet         = errors.New("statement type not set")
	ScannerRowsPointerNil       = errors.New("Scanner rows could not be nil pointer")
	ScannerEntityNeedCanSet     = errors.New("Entity need can set")
	ScannerEntiryTypeNotSupport = errors.New("Scanner Entity not support. it should be struct or slice")
	FindAllExpectSlice          = errors.New("FindAll method expect slice like []*model")
	FindOneExpectStruct         = errors.New("FindOne method expect struct like &model")
	DeleteExpectSliceOrStruct   = errors.New("Delete Method expect struct or slice")
	InsertExpectSliceOrStruct   = errors.New("Insert Method expect struct or slice")
	UpdateExpectSliceOrStruct   = errors.New("Update Method expect struct or slice")
	ModelMissingPrimaryKey      = errors.New("model missing primary key")
	ModelNotSupportType         = errors.New("model onl support model{} or &model{}")
	RecordNotFound              = errors.New("record not found")
)
