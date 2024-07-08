package util

import (
	"fmt"
	"github.com/xwb1989/sqlparser"
)

func parseUpdateSQL(sql string) ([]string, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}
	update, ok := stmt.(*sqlparser.Update)
	if !ok {
		return nil, fmt.Errorf("not an UPDATE statement")
	}
	return []string{sqlparser.String(update.TableExprs), sqlparser.String(update.Exprs), sqlparser.String(update.Where)}, nil
}

func MergeUpdateSQL(sqlStr1, sqlStr2 string) (string, error) {
	if sqlStr1 == "" || sqlStr2 == "" {
		return "", fmt.Errorf("ToSQL error: %s, %s", sqlStr1, sqlStr2)
	}

	sql1, err := parseUpdateSQL(sqlStr1)
	if err != nil {
		return "", err
	}
	sql2, err := parseUpdateSQL(sqlStr2)
	if err != nil {
		return "", err
	}

	if sql1[0] != sql2[0] || sql1[2] != sql2[2] {
		return "", fmt.Errorf("parseUpdateSQL error: %s, %s", sqlStr1, sqlStr2)
	}

	return fmt.Sprintf(`UPDATE %s SET %s, %s %s`, sql1[0], sql2[1], sql1[1], sql1[2]), nil
}
