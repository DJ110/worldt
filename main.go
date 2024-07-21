package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/bigtable"
	"github.com/labstack/echo/v4"
)

type TempItem struct {
	City        string
	Day         string
	Hour        string
	Temperature int16
}

var project *string
var instance *string

const tableName = "citytemp"
const columnFamily = "temperature"
const columnQualifier = "temperature"

func main() {
	project = flag.String("project", "", "The Google Cloud Platform project ID. Required.")
	instance = flag.String("instance", "", "The Google Cloud Bigtable instance ID. Required.")
	flag.Parse()
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World!")
	})
	e.POST("/update", postTemp)
	e.GET("/get", getLatestTempDay)
	e.GET("/getall", getTempDayAll)
	e.Logger.Fatal(e.Start(":1323"))
}

func getBigTableClient() (*bigtable.Client, context.Context) {
	ctx := context.Background()

	client, err := bigtable.NewClient(ctx, *project, *instance)
	if err != nil {
		log.Fatalf("Could not create admin client: %v", err)
	}
	return client, ctx
}

func getTempDayAll(c echo.Context) error {
	city := c.QueryParam("city")
	day := c.QueryParam("day")
	// Get Temp data from BigTable
	client, ctx := getBigTableClient()
	defer client.Close()
	tbl := client.Open(tableName)
	rowKey := city + "#" + day // tokyo#2024-07-12
	var s []uint64
	// Read the row
	err := tbl.ReadRows(ctx, bigtable.PrefixRange(rowKey), func(row bigtable.Row) bool {
		// Iterate over the column families
		version := 0
		for _, ris := range row[columnFamily] {
			if ris.Column == fmt.Sprintf("%s:%s", columnFamily, columnQualifier) {
				// Print the version timestamp and the cell value
				version++
				data := binary.BigEndian.Uint64(ris.Value)
				s = append(s, data)
				fmt.Printf("Version %d, Timestamp: %v, Value: %d\n", version, ris.Timestamp, data)
			}
		}
		return true
	}, bigtable.RowFilter(bigtable.ColumnFilter(columnQualifier)))
	if err != nil {
		log.Fatalf("Could not read row: %v", err)
		return c.String(http.StatusBadRequest, "Bad request city and day")
	}
	return c.JSON(http.StatusOK, s)
}

func getLatestTempDay(c echo.Context) error {
	city := c.QueryParam("city")
	day := c.QueryParam("day")

	// Get Temp data from BigTable
	client, ctx := getBigTableClient()
	defer client.Close()
	tbl := client.Open(tableName)
	rowKey := city + "#" + day // tokyo#2024-07-12
	row, err := tbl.ReadRow(ctx, rowKey, bigtable.RowFilter(bigtable.ColumnFilter(columnQualifier)))
	if err != nil {
		log.Fatalf("Could not read row with key %s: %v", rowKey, err)
		return c.String(http.StatusBadRequest, "Bad request city and day")
	}
	data := binary.BigEndian.Uint64(row[columnFamily][0].Value)
	log.Printf("\t%s = %s\n", rowKey, strconv.FormatUint(data, 10))

	return c.JSON(http.StatusOK, strconv.FormatUint(data, 10))
}

func postTemp(c echo.Context) error {
	temp := new(TempItem)
	if err := c.Bind(temp); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}
	// Save data into bigtable
	client, ctx := getBigTableClient()
	defer client.Close()

	// Create a mutation
	rowKey := temp.City + "#" + temp.Day // tokyo#2024-07-12

	mut := bigtable.NewMutation()
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint64(temp.Temperature))
	mut.Set(columnFamily, columnQualifier, bigtable.Now(), buf.Bytes())

	// Apply the mutation
	tbl := client.Open(tableName)
	err := tbl.Apply(ctx, rowKey, mut)
	if err != nil {
		log.Fatalf("Could not apply mutation: %v", err)
		return c.String(http.StatusBadRequest, "failed to write into bigtable")
	}
	println(rowKey)
	println(uint64(temp.Temperature))
	return c.JSON(http.StatusOK, temp)
}
