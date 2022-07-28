package schema

import (
	"go/ast"
	"reflect"
	"sorm/dialect"
)

// Field 表的列结构
// type User struct{
//		Name string `json:"name"`
// 		Age int
// }
type Field struct {
	// Name name of column
	Name string
	// Type type of column
	Type string
	// Tag tag of column
	Tag string
}

// Schema 表结构
type Schema struct {
	Model interface{}
	// Name name of table
	Name       string
	Fields     []*Field
	FieldNames []string
	FieldMap   map[string]*Field
}

// GetField return field
func (s *Schema) GetField(name string) *Field {
	return s.FieldMap[name]
}

// Parse 将任意对象解析为Schema实例
func Parse(dest interface{}, dialector dialect.Dialector) *Schema {
	// 入参是一个对象的指针，使用reflect.Indirect来获取指针指向的实例
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		FieldMap: make(map[string]*Field),
	}

	// 获取实例字段的个数
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: dialector.ConvertTypeTo(reflect.Indirect(reflect.New(p.Type))),
			}
			// 获取tag
			if v, ok := p.Tag.Lookup("sorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.FieldMap[p.Name] = field
		}
	}
	return schema
}

// RecordValues 将orm映射中session.Insert(&User{"Tom",12},&User{"Bob",15})的值
// 转换为("Tom","Bob"),(12,15) 形式
func (s *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range s.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
