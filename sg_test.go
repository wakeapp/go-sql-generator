package sg

import (
	"strings"
	"testing"
)

func TestGetInsertSQL(t *testing.T) {
	tests := []struct {
		name        string
		fields      []string
		valuesStack [][]string
		expected    map[string]string
	}{
		{
			fields: []string{
				"field_1",
				"field_2",
			},
			valuesStack: [][]string{
				{
					"field_1",
					"field_2",
				},
				{
					"field_1_1",
					"field_2_1",
				},
			},
		},
		{
			fields: []string{
				"field_1",
				"field_2",
			},
			valuesStack: [][]string{
				{
					"field_1_2",
					"field_2_2",
				},
				{
					"field_1_3",
					"field_2_3",
				},
				{
					"field_1_4",
					"field_2_4",
				},
				{
					"field_1_5",
					"field_2_5",
				},
			},
		},
		{
			fields: []string{
				"field_1",
				"field_2",
			},
			valuesStack: [][]string{
				{
					"field_1_2",
					"field_2_2",
				},
			},
		},
	}

	var sqlGenerator = MysqlSqlGenerator{}
	var dataInsert InsertData
	var valuesCount int

	for _, test := range tests {
		valuesCount = 0
		dataInsert = InsertData{
			TableName: "TestTable",
			Fields:    test.fields,
		}

		for _, values := range test.valuesStack {
			dataInsert.Add(values)
			valuesCount += len(values)
		}

		query, args, err := sqlGenerator.GetInsertSql(dataInsert)
		if err != nil {
			t.Fatalf("on GetInsertSql: %s", err)
		}

		if actualValuesCount := strings.Count(query, "?"); actualValuesCount != valuesCount {
			t.Fatalf("on compare values count and params: expected: %d, actual: %d", valuesCount, actualValuesCount)
		}

		if actualValuesCount := strings.Count(query, "?"); actualValuesCount != len(args) {
			t.Fatalf("on compare values count and args: expected: %d, actual: %d", actualValuesCount, len(args))
		}

		if valuesCount != len(args) {
			t.Fatalf("on compare values count and args: expected: %d, actual: %d", valuesCount, len(args))
		}
	}
}

func TestGetInsertSQLWithID(t *testing.T) {
	tests := []struct {
		name        string
		fields      []string
		valuesStack []struct {
			Values []string
			ID     string
		}
		expected map[string]string
	}{
		{
			fields: []string{
				"field_2",
			},
			valuesStack: []struct {
				Values []string
				ID     string
			}{
				{
					ID:     "2",
					Values: []string{"field_2"},
				},
				{
					ID:     "1",
					Values: []string{"field_1"},
				},
			},
		},
		{
			fields: []string{
				"field_2",
			},
			valuesStack: []struct {
				Values []string
				ID     string
			}{
				{
					ID:     "2",
					Values: []string{"field_2"},
				},
				{
					ID:     "1",
					Values: []string{"field_1"},
				},
				{
					ID:     "4",
					Values: []string{"field_4"},
				},
				{
					ID:     "3",
					Values: []string{"field_3"},
				},
			},
		},
		{
			fields: []string{
				"id",
				"field_2",
			},
			valuesStack: []struct {
				Values []string
				ID     string
			}{
				{
					ID:     "1",
					Values: []string{"field_1"},
				},
			},
		},
	}

	var sqlGenerator = MysqlSqlGenerator{}
	var dataInsert InsertData
	var valuesCount int

	for _, test := range tests {
		valuesCount = 0
		dataInsert = InsertData{
			TableName: "TestTable",
			Fields:    test.fields,
		}

		dataInsert.SetOptimize(true)

		for _, row := range test.valuesStack {
			dataInsert.AddWithID(row.ID, row.Values)
			valuesCount += len(row.Values)
		}

		query, args, err := sqlGenerator.GetInsertSql(dataInsert)
		if err != nil {
			t.Fatalf("on GetInsertSql: %s", err)
		}

		if actualValuesCount := strings.Count(query, "?"); actualValuesCount != valuesCount {
			t.Fatalf("on compare values count and params: expected: %d, actual: %d", valuesCount, actualValuesCount)
		}

		if actualValuesCount := strings.Count(query, "?"); actualValuesCount != len(args) {
			t.Fatalf("on compare values count and args: expected: %d, actual: %d", actualValuesCount, len(args))
		}

		if valuesCount != len(args) {
			t.Fatalf("on compare values count and args: expected: %d, actual: %d", valuesCount, len(args))
		}

		var previousRow rowValues

		for _, v := range dataInsert.ValuesList {
			if previousRow.ID != "" && v.ID < previousRow.ID {
				t.Fatalf("InsertData values not sorted")
			}
		}
	}
}
