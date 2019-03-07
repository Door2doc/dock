package main

import (
	"os"

	"github.com/denisenkom/go-mssqldb"
)

var _ mssql.Driver

func main() {
	PrintVersion(os.Stdout)
}
