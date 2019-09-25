Sql generator
===

## Обзор
Пакет служит для генерации сложных типовых запросов, таких как bulk insert, bulk update, bulk upsert 


## Примеры использования

```go
sqlGenerator := sg.MysqlSqlGenerator{}
```

### Select
```go
selectData := sg.SelectData{
    TableName: "SearchText",
    Fields:    []string{"id", "country", "os"},
    Where:     map[string]string{"id": "189920059115387768", "country": "IT"},
}

query, args, err := sqlGenerator.GetSelectSql(selectData)
```

### Insert
```go
dataInsert := sg.InsertData{
    TableName: "SearchText",
    IsIgnore:  false,
    Fields:    []string{"id", "country"},
}

dataInsert.Add([]string{"123", "it"})
dataInsert.Add([]string{"345", "ua"})
dataInsert.Add([]string{"456", "by"})

query, args, err = sqlGenerator.GetInsertSql(dataInsert)
```

### Update
```go
dataUpdate := sg.UpdateData{
    TableName: "SearchText",
}

dataUpdate.Add(
    map[string]string{"country": "ru", "os": "Android"},
    map[string]string{"id": "189920059115387768", "country": "IT"},
)

dataUpdate.Add(
    map[string]string{"country": "FR", "os": "ios"},
    map[string]string{"id": "456", "country": "UA"},
)

query, args, err = sqlGenerator.GetUpdateSql(dataUpdate)
```

### Upsert
```go
dataUpsert := sg.UpsertData{
    TableName: "SearchText",
    Fields:    []string{"id", "country"},
    ReplaceDataList: []sg.ReplaceData{
        {Field: "id"},
        {Field: "country", Type: sg.CONDITION, Condition: "1"},
        {Field: "frequency", Type: sg.INCREMENT},
    },
}

dataUpsert.Add([]string{"123", "it"})
dataUpsert.Add([]string{"345", "ua"})
dataUpsert.Add([]string{"456", "by"})

query, args, err = sqlGenerator.GetUpsertSql(dataUpsert)
```
