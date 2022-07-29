// Package clause ...
// 用于生成SQL子句
package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[WHERE] = _where
	generators[LIMIT] = _limit
	generators[ORDERBY] = _orderBy
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}

func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}
func _select(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

func _insert(values ...interface{}) (string, []interface{}) {
	// INSERT INTO $tableName $fields
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{}
}

func _values(values ...interface{}) (string, []interface{}) {
	// 拼接SQL语句VALUES
	// VALUES ($v1,$v2...), ($v3,$v4...), ...
	var sql strings.Builder
	var bindStr string
	var vars []interface{}
	sql.WriteString("VALUES")
	// values是二维数组，可包含多组value
	for i, value := range values {
		v := value.([]interface{})
		if bindStr == "" {
			bindStr = genBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%s)", bindStr))
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}

	return sql.String(), vars
}

func _limit(values ...interface{}) (string, []interface{}) {
	// LIMIT ?
	limit := values[0]
	return fmt.Sprintf("LIMIT %d", limit), []interface{}{}
}

func _orderBy(values ...interface{}) (string, []interface{}) {
	// ORDER BY ?
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

func _where(values ...interface{}) (string, []interface{}) {
	// WHERE desc
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars
}

func _update(values ...interface{}) (string, []interface{}) {
	// 接收参数是map类型的键值对
	// UPDATE $tableName SET
	var sql strings.Builder
	sql.WriteString(fmt.Sprintf("UPDATE %s ", values[0]))
	sql.WriteString("SET ")
	items := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for k, v := range items {
		keys = append(keys, k+" = ?")
		vars = append(vars, v)
	}
	sql.WriteString(strings.Join(keys, ", "))

	return sql.String(), vars

}

func _delete(values ...interface{}) (string, []interface{}) {
	// DELETE FROM $tableName
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

func _count(values ...interface{}) (string, []interface{}) {
	// SELECT COUNT(*) FROM $tableName
	// 复用_select
	return _select(values[0], []string{"COUNT(*)"})
}
