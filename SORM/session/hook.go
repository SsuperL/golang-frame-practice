// Package session ...
// 预设钩子
package session

import (
	"fmt"
	"reflect"
)

// 具体的钩子函数由结构体各自实现
const (
	BeforeQuery  = "BeforeQuery"
	BeforeInsert = "BeforeInsert"
	BeforeUpdate = "BeforeUpdate"
	BeforeDelete = "BeforeDelete"
	AfterQuery   = "AfterQuery"
	AfterInsert  = "AfterInsert"
	AfterUpdate  = "AfterUpdate"
	AfterDelete  = "AfterDelete"
)

// IBeforeQuery hook
type IBeforeQuery interface {
	BeforeQuery(s *Session) error
}

// IBeforeInsert hook
type IBeforeInsert interface {
	BeforeInsert(s *Session) error
}

// IBeforeUpdate hook
type IBeforeUpdate interface {
	BeforeUpdate(s *Session) error
}

// IBeforeDelete hook
type IBeforeDelete interface {
	BeforeDelete(s *Session) error
}

// IAfterQuery hook
type IAfterQuery interface {
	AfterQuery(s *Session) error
}

// IAfterInsert hook
type IAfterInsert interface {
	AfterInsert(s *Session) error
}

// IAfterUpdate hook
type IAfterUpdate interface {
	AfterUpdate(s *Session) error
}

// IAfterDelete hook
type IAfterDelete interface {
	AfterDelete(s *Session) error
}

// CallMethod hook function
func (s *Session) CallMethod(method string, value interface{}) error {
	if value == nil {
		return nil
	}

	param := reflect.ValueOf(value)
	switch method {
	case BeforeQuery:
		if i, ok := param.Interface().(IBeforeQuery); ok {
			i.BeforeQuery(s)
		}
	case BeforeInsert:
		if i, ok := param.Interface().(IBeforeInsert); ok {
			i.BeforeInsert(s)
		}
	case BeforeUpdate:
		if i, ok := param.Interface().(IBeforeUpdate); ok {
			i.BeforeUpdate(s)
		}
	case BeforeDelete:
		if i, ok := param.Interface().(IBeforeDelete); ok {
			i.BeforeDelete(s)
		}
	case AfterQuery:
		if i, ok := param.Interface().(IAfterQuery); ok {
			i.AfterQuery(s)
		}
	case AfterInsert:
		if i, ok := param.Interface().(IAfterInsert); ok {
			i.AfterInsert(s)
		}
	case AfterUpdate:
		if i, ok := param.Interface().(IAfterUpdate); ok {
			i.AfterUpdate(s)
		}
	case AfterDelete:
		if i, ok := param.Interface().(IAfterDelete); ok {
			i.AfterDelete(s)
		}
	default:
		return fmt.Errorf("Unsupported hook method: %v", method)
	}
	return nil
}
