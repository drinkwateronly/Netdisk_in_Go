package utils

import "github.com/go-sql-driver/mysql"

func IsDuplicateEntryErr(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == 1062 {
			return true
		}
	}
	return false
}
