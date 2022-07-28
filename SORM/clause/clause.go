package clause

import "strings"

// Type of table operation
type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
)

type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

// Set 根据Type调用对应generator，生成该子句对应的SQL
func (c *Clause) Set(typ Type, values ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[typ](values...)
	c.sql[typ] = sql
	c.sqlVars[typ] = vars
}

// Build 根据传入顺序的Type构造出完整SQL子句
func (c *Clause) Build(typs ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, typ := range typs {
		if sql, ok := c.sql[typ]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[typ]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
