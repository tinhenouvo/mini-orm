package mini_orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	dbColumnName = "columnName"
	dbPk         = "pk"
	dbReadonly   = "readOnly"
)

// Scanner convert rows to entity
// Don't scan into interface{} but the type you would expect, the database/sql package converts the returned type for you then.
type Scanner struct {
	rows          *sql.Rows
	fields        []string
	entity        interface{}
	entityValue   reflect.Value
	entityPointer reflect.Value
	Model         *Model
}

// Model describe table struct
type Model struct {
	TableName string
	Value     reflect.Value
	Fields    map[string]*Field
	PkName    string
	PkIdx     int
}

// Field describe table field
type Field struct {
	Name         string
	idx          int
	Column       reflect.StructField
	Tags         map[string]string
	IsPrimaryKey bool
	IsReadOnly   bool
}

// NewModel return new model instanc
func NewModel(value reflect.Value) *Model {
	m := &Model{Value: value}
	_, ok := value.Type().MethodByName("TableName")
	if ok {
		vals := value.MethodByName("TableName").Call([]reflect.Value{})
		if len(vals) > 0 {
			switch vals[0].Kind() {
			case reflect.String:
				m.TableName = vals[0].String()
			}
		}
	}
	m.Fields = make(map[string]*Field)
	elem := value.Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := &Field{}
		df := elem.Type().Field(i)
		fieldName := ToSnakeCase(df.Name)
		tags := make(map[string]string)
		tag := strings.Split(df.Tag.Get("sql"), ",")
		for _, t := range tag {
			ts := strings.Split(t, "=")
			if len(ts) == 1 {
				if ts[0] == dbColumnName {
					field.IsPrimaryKey = true
				}
				if ts[0] == dbReadonly {
					field.IsReadOnly = true
				}
			} else if len(ts) == 2 {
				tags[ts[0]] = ts[1]
				if ts[0] == dbColumnName {
					fieldName = ts[1]
				}
			}
		}
		field.Name = fieldName
		field.idx = i
		field.Column = df
		field.Tags = tags
		if field.IsPrimaryKey == true {
			m.PkName = fieldName
			m.PkIdx = i
		}
		m.Fields[fieldName] = field
	}
	return m
}

// NewScanner return new scanner instance
func NewScanner(dest interface{}) (*Scanner, error) {
	entityValue := reflect.ValueOf(dest)
	s := &Scanner{
		entity:        dest,
		entityValue:   entityValue,
		entityPointer: reflect.Indirect(entityValue),
	}

	switch s.entityPointer.Kind() {
	case reflect.Slice:
		if s.entityPointer.Type().Elem().Kind() == reflect.Struct {
			t := reflect.New(s.entityPointer.Type().Elem())
			s.Model = NewModel(t)
		} else if s.entityPointer.Type().Elem().Kind() == reflect.Ptr {
			t := reflect.New(s.entityPointer.Type().Elem().Elem())
			s.Model = NewModel(t)
		} else {
			return nil, ModelNotSupportType
		}
	case reflect.Struct:
		s.Model = NewModel(s.entityValue)
	default:
		return nil, ScannerEntiryTypeNotSupport
	}
	return s, nil
}

// Close close
func (sc *Scanner) Close() {
	if sc.rows != nil {
		sc.rows.Close()
	}
}

// SetRows set row
func (sc *Scanner) SetRows(rows *sql.Rows) {
	sc.rows = rows
}

// GetTableName try get table from dest
func (sc *Scanner) GetTableName() string {
	if sc.Model != nil {
		return sc.Model.TableName
	}
	return ""
}

// Convert scan rows to dest
func (sc *Scanner) Convert() error {
	if sc.rows == nil {
		return ScannerRowsPointerNil
	}
	if !sc.entityPointer.CanSet() {
		return ScannerEntityNeedCanSet
	}
	fields, err := sc.rows.Columns()
	if err != nil {
		return err
	}
	sc.fields = fields
	switch sc.entityPointer.Kind() {
	case reflect.Slice:
		return sc.convertAll()
	case reflect.Struct:
		return sc.convertOne()
	default:
		return ScannerEntiryTypeNotSupport
	}
}

func (sc *Scanner) convertAll() error {
	dest := reflect.MakeSlice(sc.entityPointer.Type(), 0, 0)
	for sc.rows.Next() {
		srcValue := make([]interface{}, len(sc.fields))
		for i := 0; i < len(sc.fields); i++ {
			var v interface{}
			srcValue[i] = &v
		}
		err := sc.rows.Scan(srcValue...)
		if err != nil {
			return err
		}
		t := reflect.New(sc.entityPointer.Type().Elem().Elem())
		sc.SetEntity(srcValue, t.Elem())
		dest = reflect.Append(dest, t)
	}
	sc.entityPointer.Set(dest)
	return nil
}

func (sc *Scanner) convertOne() error {
	srcValue := make([]interface{}, len(sc.fields))
	for i := 0; i < len(sc.fields); i++ {
		var v interface{}
		srcValue[i] = &v
	}
	if sc.rows.Next() {
		err := sc.rows.Scan(srcValue...)
		if err != nil {
			return err
		}
		return sc.SetEntity(srcValue, sc.entityPointer)
	}
	return RecordNotFound
}

// SetEntity set entity
func (sc *Scanner) SetEntity(srcValue []interface{}, dest reflect.Value) error {
	tmpMap := make(map[string]interface{})
	for i := 0; i < len(sc.fields); i++ {
		f := sc.fields[i]
		v := srcValue[i]
		tmpMap[f] = v
	}
	for name, field := range sc.Model.Fields {
		val, ok := tmpMap[name]
		if !ok {
			continue
		}
		ff := dest.Field(field.idx)
		rawVal := reflect.Indirect(reflect.ValueOf(val))
		rawValInterface := rawVal.Interface()
		if rawValInterface == nil {
			continue
		}
		switch ff.Kind() {
		case reflect.String:
			switch d := rawValInterface.(type) {
			case string:
				ff.SetString(d)
			case bool:
				ff.SetString(strconv.FormatBool(d))
			case float64:
				ff.SetString(strconv.FormatFloat(d, 'f', -1, 64))
			case float32:
				ff.SetString(strconv.FormatFloat(float64(d), 'f', -1, 32))
			case int:
				ff.SetString(strconv.FormatInt(int64(d), 10))
			case int8:
				ff.SetString(strconv.FormatInt(int64(d), 10))
			case int16:
				ff.SetString(strconv.FormatInt(int64(d), 10))
			case int32:
				ff.SetString(strconv.FormatInt(int64(d), 10))
			case int64:
				ff.SetString(strconv.FormatInt(d, 10))
			case uint:
				ff.SetString(strconv.FormatUint(uint64(d), 10))
			case uint8:
				ff.SetString(strconv.FormatUint(uint64(d), 10))
			case uint16:
				ff.SetString(strconv.FormatUint(uint64(d), 10))
			case uint32:
				ff.SetString(strconv.FormatUint(uint64(d), 10))
			case uint64:
				ff.SetString(strconv.FormatUint(uint64(d), 10))
			case []byte:
				ff.SetString(string(d))
			default:
				// https://github.com/spf13/cast/blob/master/caste.go#L778
				// 尝试转化成 fmt.Stringer
				var errorType = reflect.TypeOf((*error)(nil)).Elem()
				var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
				v := reflect.ValueOf(d)
				for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
					v = v.Elem()
				}
				switch vv := v.Interface().(type) {
				case fmt.Stringer:
					ff.SetString(vv.String())
				}
			}
		case reflect.Bool:
			switch d := rawValInterface.(type) {
			case bool:
				ff.SetBool(d)
			case *bool:
				ff.SetBool(*d)
			case int, int8, int16, int32, int64:
				vv := reflect.ValueOf(rawValInterface)
				if vv.Int() > 0 {
					ff.SetBool(true)
				}
			case uint, uint8, uint16, uint32, uint64:
				vv := reflect.ValueOf(rawValInterface)
				if vv.Uint() > 0 {
					ff.SetBool(true)
				}
			case float32, float64:
				vv := reflect.ValueOf(rawValInterface)
				if vv.Float() > 0 {
					ff.SetBool(true)
				}
			case string:
				if d != "" {
					ff.SetBool(true)
				}
			case []byte:
				if string(d) != "" {
					ff.SetBool(true)
				}
			default:
				sc.defaultConvert(rawValInterface, &ff, field)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			switch d := rawValInterface.(type) {
			case int:
				ff.SetInt(int64(d))
			case int8:
				ff.SetInt(int64(d))
			case int16:
				ff.SetInt(int64(d))
			case int32:
				ff.SetInt(int64(d))
			case int64:
				ff.SetInt(d)
			case uint:
				ff.SetInt(int64(d))
			case uint8:
				ff.SetInt(int64(d))
			case uint16:
				ff.SetInt(int64(d))
			case uint32:
				ff.SetInt(int64(d))
			case uint64:
				ff.SetInt(int64(d))
			case float32:
				ff.SetInt(int64(d))
			case float64:
				ff.SetInt(int64(d))
			case []uint8:
				v, err := strconv.ParseInt(string(d), 0, 0)
				if err != nil {
					return fmt.Errorf("can not convert field:%s %s to int64 err:%v", name, string(d), err)
				}
				ff.SetInt(v)
			case string:
				v, err := strconv.ParseInt(d, 0, 0)
				if err != nil {
					return fmt.Errorf("can not convert field:%s %s to int64 err:%v", name, string(d), err)
				}
				ff.SetInt(v)
			case bool:
				if d {
					ff.SetInt(1)
				}
			default:
				sc.defaultConvert(rawValInterface, &ff, field)
			}
		case reflect.Float32, reflect.Float64:
			switch d := rawValInterface.(type) {
			case int:
				ff.SetFloat(float64(d))
			case int8:
				ff.SetFloat(float64(d))
			case int16:
				ff.SetFloat(float64(d))
			case int32:
				ff.SetFloat(float64(d))
			case int64:
				ff.SetFloat(float64(d))
			case uint:
				ff.SetFloat(float64(d))
			case uint8:
				ff.SetFloat(float64(d))
			case uint16:
				ff.SetFloat(float64(d))
			case uint32:
				ff.SetFloat(float64(d))
			case uint64:
				ff.SetFloat(float64(d))
			case float32:
				ff.SetFloat(float64(d))
			case float64:
				ff.SetFloat(d)
			case []uint8:
				v, err := strconv.ParseFloat(string(d), 64)
				if err != nil {
					return fmt.Errorf("can not convert field:%s %s to float64 err:%v", name, string(d), err)
				}
				ff.SetFloat(float64(v))
			case string:
				v, err := strconv.ParseFloat(d, 64)
				if err != nil {
					return fmt.Errorf("can not convert field:%s %s to float64 err:%v", name, string(d), err)
				}
				ff.SetFloat(float64(v))
			case bool:
				if d {
					ff.SetFloat(1)
				}
			default:
				sc.defaultConvert(rawValInterface, &ff, field)
			}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			switch d := rawValInterface.(type) {
			case int:
				ff.SetUint(uint64(d))
			case int8:
				ff.SetUint(uint64(d))
			case int16:
				ff.SetUint(uint64(d))
			case int32:
				ff.SetUint(uint64(d))
			case int64:
				ff.SetUint(uint64(d))
			case uint:
				ff.SetUint(uint64(d))
			case uint8:
				ff.SetUint(uint64(d))
			case uint16:
				ff.SetUint(uint64(d))
			case uint32:
				ff.SetUint(uint64(d))
			case uint64:
				ff.SetUint(d)
			case float32:
				ff.SetUint(uint64(d))
			case float64:
				ff.SetUint(uint64(d))
			case []uint8:
				v, err := strconv.ParseInt(string(d), 0, 0)
				if err != nil {
					return fmt.Errorf("can not convert field:%s %s to int64 err:%v", name, string(d), err)
				}
				ff.SetUint(uint64(v))
			case string:
				v, err := strconv.ParseInt(d, 0, 0)
				if err != nil {
					return fmt.Errorf("can not convert field:%s %s to int64 err:%v", name, string(d), err)
				}
				ff.SetUint(uint64(v))
			case bool:
				if d {
					ff.SetUint(1)
				}
			default:
				sc.defaultConvert(rawValInterface, &ff, field)
			}
		default:
			sc.defaultConvert(rawValInterface, &ff, field)
		}
	}
	return nil
}

func (sc *Scanner) defaultConvert(rawValInterface interface{}, ff *reflect.Value, field *Field) {
	vv := reflect.ValueOf(rawValInterface)
	if vv.IsValid() {
		if vv.Type().ConvertibleTo(ff.Type()) {
			ff.Set(vv.Convert(ff.Type()))
		} else {
			if ff.Kind() == reflect.Ptr {
				if ff.IsNil() {
					ff.Set(reflect.New(field.Column.Type.Elem()))
				}
				ffElem := ff.Elem()
				if vv.Type().ConvertibleTo(ffElem.Type()) {
					ffElem.Set(vv.Convert(ffElem.Type()))
				}
			}
		}
	}
}
