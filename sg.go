package sg

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	INCREMENT = "inc"
	DECREMENT = "dec"
	CONDITION = "cond"
	idField   = "id"
)

type MysqlSqlGenerator struct {
}

type SelectData struct {
	TableName string
	Fields    []string
	Where     map[string]string
}

func (sg MysqlSqlGenerator) GetSelectSql(data SelectData) (string, []interface{}, error) {
	params := make(map[string]interface{}, 0)
	whereParts := make([]string, 0, len(data.Where))

	index := 0

	for whereField, whereValue := range data.Where {
		index++
		namedParam := getNamedParam(whereField, index)

		wherePart := getNamedCondition(whereField, index)
		whereParts = append(whereParts, wherePart)

		params[namedParam[1:]] = whereValue
	}

	sql := fmt.Sprintf("SELECT %s FROM `%s` WHERE %s", strings.Join(data.Fields, ", "), data.TableName, strings.Join(whereParts, " AND "))

	query, args, err := sqlx.Named(sql, params)

	return query, args, err
}

type rowValues struct {
	Values []string
	ID     string
}

type rows []rowValues

var _ sort.Interface = rows{}

func (r rows) Len() int {
	return len(r)
}

func (r rows) Less(i, j int) bool {
	return r[i].ID < r[j].ID
}

func (r rows) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// InsertData - data for perform insert operation
type InsertData struct {
	TableName  string
	IsIgnore   bool
	Fields     []string
	ValuesList rows
	optimize   bool
}

// SetOptimize - set support for sql optimize
func (d *InsertData) SetOptimize(o bool) {
	d.optimize = o
}

// Optimize - return optimize state
func (d *InsertData) Optimize() bool {
	return d.optimize
}

// Add - add row to struct
func (d *InsertData) Add(values []string) {
	if len(d.ValuesList) == 0 {
		d.ValuesList = make([]rowValues, 0)
	}

	d.ValuesList = append(d.ValuesList, rowValues{Values: values})
}

// GetInsertSql - bind params and values to sql query
func (sg MysqlSqlGenerator) GetInsertSql(data InsertData) (string, []interface{}, error) {
	var params = make(map[string]interface{}, 0)
	var values = make([]string, 0, len(data.ValuesList))

	var namedParam, value, field, ignore string
	var key, valuesIndex, index int
	var namedParams []string
	var valuesData rowValues

	if len(data.ValuesList) > 0 {
		namedParams = make([]string, 0, len(data.ValuesList[0].Values))
	}

	for valuesIndex, valuesData = range data.ValuesList {
		for key, value = range valuesData.Values {
			index++
			field = data.Fields[key]

			if data.Optimize() && strings.ToLower(field) == idField {
				data.ValuesList[valuesIndex].ID = value
			}

			namedParam = getNamedParam(field, index)
			namedParams = append(namedParams, namedParam)

			params[namedParam[1:]] = value
		}

		values = append(values, "("+strings.Join(namedParams, ", ")+")")

		namedParams = namedParams[:0]
	}

	if data.IsIgnore {
		ignore = "IGNORE"
	}

	if data.optimize {
		sort.Sort(data.ValuesList)
	}

	var sql = fmt.Sprintf(
		"INSERT %s INTO %s (%s) VALUES %s",
		ignore,
		data.TableName,
		strings.Join(data.Fields, ", "),
		strings.Join(values, ", "),
	)

	return sqlx.Named(sql, params)
}

func getNamedParam(field string, index int) string {
	return fmt.Sprintf(":%s_%d", field, index)
}

func getNamedCondition(field string, index int) string {
	namedParam := getNamedParam(field, index)
	namedCondition := fmt.Sprintf("%s = %s", field, namedParam)

	return namedCondition
}

type updateDataList struct {
	Set   map[string]string
	Where map[string]string
}

type UpdateData struct {
	TableName string
	List      []updateDataList
}

func (d *UpdateData) Add(set map[string]string, where map[string]string) {
	if len(d.List) == 0 {
		d.List = make([]updateDataList, 0)
	}

	d.List = append(d.List, updateDataList{Set: set, Where: where})
}

func (sg MysqlSqlGenerator) GetUpdateSql(data UpdateData) (string, []interface{}, error) {
	whereParts := make([]string, 0, len(data.List))

	values := make(map[string]map[string]string, 0)
	params := make(map[string]interface{}, 0)

	index := 0

	for _, updateData := range data.List {
		whenParts := make([]string, 0, len(updateData.Where))
		thenParts := make([]string, 0, len(updateData.Where))

		for whereField, whereValue := range updateData.Where {
			index++

			namedParam := getNamedParam(whereField, index)
			whenPart := getNamedCondition(whereField, index)

			whenParts = append(whenParts, whenPart)
			params[namedParam[1:]] = whereValue
		}

		when := "(" + strings.Join(whenParts, " AND ") + ")"

		whereParts = append(whereParts, when)

		for setField, setValue := range updateData.Set {
			index++

			namedParam := getNamedParam(setField, index)
			thenPart := getNamedCondition(setField, index)

			thenParts = append(thenParts, thenPart)
			params[namedParam[1:]] = setValue

			if _, ok := values[setField]; !ok {
				values[setField] = make(map[string]string, 0)
			}

			values[setField][when] = namedParam
		}
	}

	setParts := make([]string, 0, len(values))

	for setField, data := range values {
		conditionParts := make([]string, 0, len(data))

		for when, namedParam := range data {
			conditionPart := fmt.Sprintf("WHEN %s THEN %s", when, namedParam)
			conditionParts = append(conditionParts, conditionPart)
		}

		setPart := fmt.Sprintf("%[1]s = CASE %s ELSE %[1]s END", setField, strings.Join(conditionParts, " "))

		setParts = append(setParts, setPart)
	}

	sql := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", data.TableName, strings.Join(setParts, ", "), strings.Join(whereParts, " OR "))

	query, args, err := sqlx.Named(sql, params)

	return query, args, err
}

type UpsertData struct {
	TableName       string
	Fields          []string
	ValuesList      []rowValues
	ReplaceDataList []ReplaceData
}

type ReplaceData struct {
	Field     string
	Type      string
	Condition string
}

func (d *UpsertData) Add(values []string) {
	if len(d.ValuesList) == 0 {
		d.ValuesList = make([]rowValues, 0)
	}

	d.ValuesList = append(d.ValuesList, rowValues{Values: values})
}

func (sg MysqlSqlGenerator) GetUpsertSql(data UpsertData) (string, []interface{}, error) {
	InsertBulkData := InsertData{
		TableName:  data.TableName,
		Fields:     data.Fields,
		ValuesList: data.ValuesList,
		IsIgnore:   false,
	}

	query, args, err := sg.GetInsertSql(InsertBulkData)

	query += " ON DUPLICATE KEY UPDATE "

	sqlReplaceParts := make([]string, 0, len(data.ReplaceDataList))

	for _, replaceData := range data.ReplaceDataList {

		switch replaceData.Type {
		case INCREMENT:
			sqlReplaceParts = append(sqlReplaceParts, fmt.Sprintf("%[1]s = %[1]s + VALUES(%[1]s)", replaceData.Field))
		case DECREMENT:
			sqlReplaceParts = append(sqlReplaceParts, fmt.Sprintf("%[1]s = %[1]s - VALUES(%[1]s)", replaceData.Field))
		case CONDITION:
			sqlReplaceParts = append(sqlReplaceParts, fmt.Sprintf("%s = %s", replaceData.Field, replaceData.Condition))
		default:
			sqlReplaceParts = append(sqlReplaceParts, fmt.Sprintf("%[1]s = VALUES(%[1]s)", replaceData.Field))
		}
	}

	query += strings.Join(sqlReplaceParts, ", ")

	return query, args, err
}
