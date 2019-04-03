package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/xo/xo/loaders"
	"github.com/xo/xo/models"
)

func main() {
	connString := os.Args[1]

	conn, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal("connecting to db: ", err)
	}
	schema, err := loadSchema(conn)
	if err != nil {
		log.Fatal("loading schema: ", err)
	}

	printDot(schema)
}

type tableWithFKs struct {
	table *models.Table
	fks   []*models.ForeignKey
}

type schema []*tableWithFKs

func loadSchema(conn *sql.DB) ([]*tableWithFKs, error) {
	var out schema
	tables, err := loaders.PgTables(conn, "public", "r")
	if err != nil {
		return nil, err
	}
	for _, table := range tables {
		foreignKeys, err := models.PgTableForeignKeys(conn, "public", table.TableName)
		if err != nil {
			return nil, err
		}
		out = append(out, &tableWithFKs{
			table: table,
			fks:   foreignKeys,
		})
	}
	return out, nil
}

func printDot(schema []*tableWithFKs) {
	fmt.Println("digraph schema {")

	fmt.Println("splines=ortho")
	fmt.Println("nodesep=0.4")
	fmt.Println("ranksep=0.8")
	fmt.Println(`node [shape="box",style="rounded,filled"]`)
	fmt.Println(`edge [arrowsize="0.5"]`)

	for _, table := range schema {
		fmt.Printf("\t\"%s\" [color=\"paleturquoise\"];\n", table.table.TableName)

		for _, fk := range table.fks {
			fmt.Printf(
				"\t\"%s\" -> \"%s\" [xlabel=\"%s\"];\n",
				table.table.TableName, fk.RefTableName, fk.ColumnName,
			)
		}
	}

	fmt.Println("}")
}
