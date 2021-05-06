package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
)

type Response map[string]interface{}

type ColumnInfo struct {
	Name         string
	IsPrimaryKey bool
	IsAutoInc    bool
	HasDefault   bool
	DbType       string
	IsNullable   bool
}

type Column struct {
	Field   sql.NullString
	Type    sql.NullString
	Null    sql.NullString
	Key     sql.NullString
	Default sql.NullString
	Extra   sql.NullString
}

type BadRequestError struct {
	msg string
}

func (br *BadRequestError) Error() string {
	return br.msg
}

type DbExplorer struct {
	db     *sql.DB
	tables map[string][]ColumnInfo
}

func NewDbExplorer(db *sql.DB) (*DbExplorer, error) {
	tables := readAllTables(db)
	return &DbExplorer{db, tables}, nil
}

func (explorer *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		explorer.ServeHTTPGet(w, r)
	}
	if r.Method == http.MethodPut {
		explorer.ServeHTTPPut(w, r)
	}
	if r.Method == http.MethodPost {
		explorer.ServeHTTPPost(w, r)
	}
	if r.Method == http.MethodDelete {
		explorer.ServeHTTPDelete(w, r)
	}
}

func (explorer *DbExplorer) ServeHTTPGet(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		response, err := explorer.getAllTablesResponse()
		setHeaderAndResponse(w, response, err)
		return
	}

	if len(splitAndSkeepEmpty(r.URL.Path, "/")) > 2 {
		setBadRequest(w)
		return
	}

	if len(splitAndSkeepEmpty(r.URL.Path, "/")) == 1 {
		explorer.ServeHTTPGetAllRecords(w, r)
		return
	}

	if len(splitAndSkeepEmpty(r.URL.Path, "/")) == 2 {
		explorer.ServeHTTPGetSingleRecord(w, r)
		return
	}

	setBadRequest(w)
}

func splitAndSkeepEmpty(s string, sep string) []string {
	result := make([]string, 0)
	for _, sp := range strings.Split(s, sep) {
		if sp != "" {
			result = append(result, sp)
		}
	}
	return result
}

func (explorer *DbExplorer) ServeHTTPPut(w http.ResponseWriter, r *http.Request) {
	if len(splitAndSkeepEmpty(r.URL.Path, "/")) != 1 {
		setBadRequest(w)
		return
	}

	explorer.ServeHTTPCreateNewRecord(w, r)
}

func (explorer *DbExplorer) ServeHTTPPost(w http.ResponseWriter, r *http.Request) {
	if len(splitAndSkeepEmpty(r.URL.Path, "/")) != 2 {
		setBadRequest(w)
		return
	}

	tableName := (splitAndSkeepEmpty(r.URL.Path, "/"))[0]
	id := (splitAndSkeepEmpty(r.URL.Path, "/"))[1]
	tableColumns, hasTable := explorer.tables[tableName]
	if !hasTable {
		setUnknownTable(w)
		return
	}

	decoder := json.NewDecoder(r.Body)
	request := make(map[string]interface{})
	err := decoder.Decode(&request)
	if err != nil {
		setHeaderAndResponse(w, nil, err)
		return
	}

	response, err := explorer.EditRecord(tableName, id, tableColumns, request)
	setHeaderAndResponse(w, response, err)
}

func (explorer *DbExplorer) ServeHTTPDelete(w http.ResponseWriter, r *http.Request) {
	if len(splitAndSkeepEmpty(r.URL.Path, "/")) != 2 {
		setBadRequest(w)
		return
	}
	tableName := (splitAndSkeepEmpty(r.URL.Path, "/"))[0]
	id := (splitAndSkeepEmpty(r.URL.Path, "/"))[1]
	tableColumns, hasTable := explorer.tables[tableName]
	if !hasTable {
		setUnknownTable(w)
		return
	}
	response, err := explorer.DeleteRecord(tableName, id, tableColumns)
	setHeaderAndResponse(w, response, err)
}

func (explorer *DbExplorer) DeleteRecord(table string, id string, columns []ColumnInfo) (Response, error) {
	response := make(map[string]interface{})
	primaryKey := ""
	for _, col := range columns {
		if col.IsPrimaryKey {
			primaryKey = col.Name
		}
	}
	query := "delete from " + table + " where " + primaryKey + " = ?"
	result, err := explorer.db.Exec(query, id)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	affected, _ := result.RowsAffected()
	response["deleted"] = affected
	return response, nil
}

func (explorer *DbExplorer) EditRecord(table string, id string, columns []ColumnInfo, editParams map[string]interface{}) (Response, error) {
	response := make(map[string]interface{})
	if len(editParams) == 0 {
		response["updated"] = 0
		return response, nil
	}
	primaryKey := ""
	for _, col := range columns {
		if col.IsPrimaryKey {
			primaryKey = col.Name
		}
	}
	values := make([]interface{}, 0)
	query := "update " + table + " set "
	first := true
	for k, v := range editParams {
		if !isValid(columns, k, v) {
			return nil, &BadRequestError{msg: "field " + k + " have invalid type"}
		}
		values = append(values, v)
		if !first {
			query = query + ","
		}
		query = query + "`" + k + "` = ?"
		first = false
	}
	query = query + " where `" + primaryKey + "` = ?"
	values = append(values, id)
	result, err := explorer.db.Exec(query, values...)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	affected, _ := result.RowsAffected()
	response["updated"] = affected
	return response, nil
}

func isValid(columns []ColumnInfo, column string, value interface{}) bool {
	for _, col := range columns {
		if col.Name == column {
			if col.IsAutoInc || col.IsPrimaryKey {
				return false
			}
			if value == nil && col.IsNullable {
				return true
			}
			if strings.Contains(col.DbType, "char") || strings.Contains(col.DbType, "text") {
				_, ok := value.(string)
				return ok
			}
			if strings.Contains(col.DbType, "bool") {
				_, ok := value.(bool)
				return ok
			}
			if strings.Contains(col.DbType, "int") {
				_, ok := value.(int)
				return ok
			}
			if strings.Contains(col.DbType, "float") {
				_, okInt := value.(int)
				_, okFloat := value.(float32)
				return okInt || okFloat
			}
		}
	}
	return false
}

func (explorer *DbExplorer) ServeHTTPCreateNewRecord(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	request := make(map[string]interface{})
	err := decoder.Decode(&request)
	if err != nil {
		setHeaderAndResponse(w, nil, err)
		return
	}

	tableName := (splitAndSkeepEmpty(r.URL.Path, "/"))[0]
	columns, hasTable := explorer.tables[tableName]

	if !hasTable {
		setUnknownTable(w)
		return
	}

	values := make([]interface{}, 0)
	columTitles := make([]string, 0)
	questionMarks := make([]string, 0)
	primaryKey := ""
	for _, col := range columns {
		if col.IsPrimaryKey {
			primaryKey = col.Name
		}

		if col.IsAutoInc {
			continue
		}
		fromRequest, hasParam := request[col.Name]

		if hasParam {
			columTitles = append(columTitles, "`"+col.Name+"`")
			values = append(values, fromRequest)
			questionMarks = append(questionMarks, "?")

			continue
		}

		if col.HasDefault || col.IsNullable {
			continue
		}

		columTitles = append(columTitles, "`"+col.Name+"`")
		values = append(values, getDefaultValue(&col))
		questionMarks = append(questionMarks, "?")
	}

	query := fmt.Sprintf("insert into %s (%s) values (%s)", tableName, strings.Join(columTitles, ","), strings.Join(questionMarks, ","))

	result, err := explorer.db.Exec(query, values...)

	if err != nil {
		fmt.Print(err.Error())
		setHeaderAndResponse(w, nil, err)
		return
	}

	lastId, _ := result.LastInsertId()

	response := make(map[string]interface{})
	response[primaryKey] = lastId
	setHeaderAndResponse(w, response, nil)
}

func getDefaultValue(col *ColumnInfo) interface{} {
	if strings.Contains(col.DbType, "char") || strings.Contains(col.DbType, "text") {
		return ""
	}

	if strings.Contains(col.DbType, "bool") {
		return false
	}

	return 0
}

func (explorer *DbExplorer) ServeHTTPGetAllRecords(w http.ResponseWriter, r *http.Request) {
	_, tableName := path.Split(r.URL.Path)

	if _, hasTable := explorer.tables[tableName]; !hasTable {
		setUnknownTable(w)
		return
	}

	limit := getParamInt(r, "limit", 5)
	offset := getParamInt(r, "offset", 0)

	response, err := explorer.getTableRecords(tableName, limit, offset)
	setHeaderAndResponse(w, response, err)
}

func (explorer *DbExplorer) ServeHTTPGetSingleRecord(w http.ResponseWriter, r *http.Request) {
	dir, id := path.Split(r.URL.Path)
	tableName := path.Base(dir)
	tableColumns, hasTable := explorer.tables[tableName]
	if !hasTable {
		setUnknownTable(w)
		return
	}

	primaryKey := ""
	for _, col := range tableColumns {
		if col.IsPrimaryKey {
			primaryKey = col.Name
		}
	}

	if primaryKey == "" {
		setBadRequest(w)
		return
	}

	response, err := explorer.getTableRecordById(tableName, primaryKey, id)
	if response == nil && err == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"record not found\"}"))
		return
	}
	setHeaderAndResponse(w, response, err)
}

func getParamInt(r *http.Request, paramName string, defaultValue int) int {
	result := defaultValue
	resultStr := r.URL.Query().Get(paramName)
	if resultStr != "" {
		res, err := strconv.Atoi(resultStr)
		if err != nil {
			return defaultValue
		}
		result = res
	}
	return result
}

func setHeaderAndResponse(w http.ResponseWriter, response Response, err error) {
	if err != nil {
		switch err.(type) {
		case *BadRequestError:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write([]byte("{\"error\": \"" + err.Error() + "\"}"))
		return
	}

	if response != nil {
		result := make(map[string]interface{})
		result["response"] = response
		jsonResult, _ := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResult)
	}
}

func setUnknownTable(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("{\"error\": \"unknown table\"}"))
}

func setBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("{\"error\": \"bad request\"}"))
}

func (explorer *DbExplorer) getAllTablesResponse() (Response, error) {
	result := make(map[string]interface{})
	tables := make([]string, 0, len(explorer.tables))
	for t, _ := range explorer.tables {
		tables = append(tables, t)
	}
	sort.Strings(tables)
	result["tables"] = tables
	return result, nil
}

func (explorer *DbExplorer) getTableRecords(table string, limit int, offset int) (Response, error) {
	records := make([]Response, 0)
	query := fmt.Sprintf("select * from %s limit %d offset %d", table, limit, offset)
	rows, _ := explorer.db.Query(query)
	defer rows.Close()
	for rows.Next() {
		record, err := readCurrentRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	response := make(map[string]interface{})
	response["records"] = records
	return response, nil
}

func (explorer *DbExplorer) getTableRecordById(table string, primaryKey string, id string) (Response, error) {
	query := fmt.Sprintf("select * from %s where %s = ?", table, primaryKey)
	rows, err := explorer.db.Query(query, id)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		record, err := readCurrentRecord(rows)
		if err != nil {
			return nil, err
		}
		response := make(map[string]interface{})
		response["record"] = record
		return response, nil
	}

	return nil, nil
}

func readCurrentRecord(rows *sql.Rows) (map[string]interface{}, error) {
	colNames, _ := rows.Columns()
	cols, _ := rows.ColumnTypes()
	columns := make([]interface{}, len(cols))

	columnPointers := make([]interface{}, len(cols))
	for i, _ := range columns {
		columnPointers[i] = &columns[i]
	}
	if err := rows.Scan(columnPointers...); err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for i, colName := range colNames {
		switch cols[i].DatabaseTypeName() {
		case "INT":
			int32Value := &sql.NullInt32{}
			int32Value.Scan(columns[i])
			if int32Value.Valid {
				result[colName] = int32Value.Int32
			} else {
				result[colName] = nil
			}
		case "VARCHAR":
			str := &sql.NullString{}
			str.Scan(columns[i])
			if str.Valid {
				result[colName] = str.String
			} else {
				result[colName] = nil
			}
		case "TEXT":
			str := &sql.NullString{}
			str.Scan(columns[i])
			if str.Valid {
				result[colName] = str.String
			} else {
				result[colName] = nil
			}
		case "FLOAT":
			fl := &sql.NullFloat64{}
			fl.Scan(columns[i])
			if fl.Valid {
				result[colName] = fl.Float64
			} else {
				result[colName] = nil
			}

		default:
			result[colName] = nil
		}

	}
	return result, nil
}

func readAllTables(db *sql.DB) map[string][]ColumnInfo {
	tables := make(map[string][]ColumnInfo)
	{
		rows, _ := db.Query("SHOW TABLES")
		defer rows.Close()
		for rows.Next() {
			table := ""
			rows.Scan(&table)
			tables[table] = make([]ColumnInfo, 0)
		}
	}
	{
		for table, _ := range tables {
			columnsInfos := make([]ColumnInfo, 0)
			columnRows, _ := db.Query("SHOW COLUMNS FROM " + table)
			defer columnRows.Close()
			for columnRows.Next() {
				column := &Column{}
				columnRows.Scan(&column.Field, &column.Type, &column.Null, &column.Key, &column.Default, &column.Extra)
				columnInfo := ColumnInfo{
					Name:         column.Field.String,
					IsPrimaryKey: column.Key.String == "PRI",
					IsAutoInc:    strings.Contains(column.Extra.String, "auto_increment"),
					HasDefault:   column.Default.Valid,
					DbType:       column.Type.String,
					IsNullable:   column.Null.String == "YES",
				}
				columnsInfos = append(columnsInfos, columnInfo)
			}
			tables[table] = columnsInfos
		}
	}
	return tables
}
