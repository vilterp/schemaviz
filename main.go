package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/emicklei/dot"
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

	g := makeGraph(schema)
	fmt.Println(g)
}

type tableWithFKs struct {
	table *models.Table
	fks   []*models.ForeignKey
}

type schema []*tableWithFKs

func loadSchema(conn *sql.DB) (schema, error) {
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

func makeGraph(schema schema) *dot.Graph {
	g := dot.NewGraph(dot.Directed)
	g.Attr("splines", "ortho")
	g.Attr("nodesep", "0.4")
	g.Attr("ranksep", "0.8")

	tableNodes := map[string]dot.Node{}

	for _, tfk := range schema {
		tableNode := g.Node(tfk.table.TableName).
			Box().
			Attr("color", "paleturquoise").
			Attr("id", tfk.table.TableName)
		tableNodes[tfk.table.TableName] = tableNode
	}
	for _, tfk := range schema {
		for _, fk := range tfk.fks {
			fromName := tfk.table.TableName
			toName := fk.RefTableName
			fromNode := tableNodes[fromName]
			toNode := tableNodes[toName]
			g.Edge(fromNode, toNode).
				Attr("xlabel", fk.ColumnName).
				Attr("id", fmt.Sprintf("%s->%s", fromName, toName))
		}
	}
	return g
}
