package qb

import (
  "fmt"
  "strings"
)

type SelectBuilder struct {
  columns []string
  table   string
  where   []string
  args    []any
  orderBy string
  limit   int
  hasLimit bool
}

func Select(columns ...string) *SelectBuilder {
  return &SelectBuilder{columns: columns}
}

func (b *SelectBuilder) From(table string) *SelectBuilder {
  b.table = table
  return b
}

func (b *SelectBuilder) Where(condition string, args ...any) *SelectBuilder {
  clause, nextArgs := replacePlaceholders(condition, b.args, args)
  b.where = append(b.where, clause)
  b.args = append(b.args, nextArgs...)
  return b
}

func (b *SelectBuilder) WhereEq(column string, value any) *SelectBuilder {
  return b.Where(column+" = ?", value)
}

func (b *SelectBuilder) OrderBy(order string) *SelectBuilder {
  b.orderBy = order
  return b
}

func (b *SelectBuilder) Limit(limit int) *SelectBuilder {
  b.limit = limit
  b.hasLimit = true
  return b
}

func (b *SelectBuilder) Build() (string, []any) {
  query := "select " + strings.Join(b.columns, ", ") + " from " + b.table
  if len(b.where) > 0 {
    query += " where " + strings.Join(b.where, " and ")
  }
  if b.orderBy != "" {
    query += " order by " + b.orderBy
  }
  if b.hasLimit {
    query += fmt.Sprintf(" limit %d", b.limit)
  }
  return query, b.args
}

type InsertBuilder struct {
  table     string
  columns   []string
  values    []any
  returning []string
}

func Insert(table string) *InsertBuilder {
  return &InsertBuilder{table: table}
}

func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
  b.columns = columns
  return b
}

func (b *InsertBuilder) Values(values ...any) *InsertBuilder {
  b.values = values
  return b
}

func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
  b.returning = columns
  return b
}

func (b *InsertBuilder) Build() (string, []any) {
  placeholders := make([]string, len(b.values))
  for i := range b.values {
    placeholders[i] = fmt.Sprintf("$%d", i+1)
  }

  query := "insert into " + b.table +
    " (" + strings.Join(b.columns, ", ") + ") values (" + strings.Join(placeholders, ", ") + ")"

  if len(b.returning) > 0 {
    query += " returning " + strings.Join(b.returning, ", ")
  }

  return query, b.values
}

type UpdateBuilder struct {
  table     string
  sets      []string
  args      []any
  where     []string
  returning []string
}

func Update(table string) *UpdateBuilder {
  return &UpdateBuilder{table: table}
}

func (b *UpdateBuilder) Set(column string, value any) *UpdateBuilder {
  placeholder := fmt.Sprintf("$%d", len(b.args)+1)
  b.sets = append(b.sets, column+" = "+placeholder)
  b.args = append(b.args, value)
  return b
}

func (b *UpdateBuilder) Where(condition string, args ...any) *UpdateBuilder {
  clause, nextArgs := replacePlaceholders(condition, b.args, args)
  b.where = append(b.where, clause)
  b.args = append(b.args, nextArgs...)
  return b
}

func (b *UpdateBuilder) WhereEq(column string, value any) *UpdateBuilder {
  return b.Where(column+" = ?", value)
}

func (b *UpdateBuilder) Returning(columns ...string) *UpdateBuilder {
  b.returning = columns
  return b
}

func (b *UpdateBuilder) Build() (string, []any) {
  if len(b.where) == 0 {
    panic("qb: update missing where clause")
  }
  query := "update " + b.table + " set " + strings.Join(b.sets, ", ")
  if len(b.where) > 0 {
    query += " where " + strings.Join(b.where, " and ")
  }
  if len(b.returning) > 0 {
    query += " returning " + strings.Join(b.returning, ", ")
  }
  return query, b.args
}

type DeleteBuilder struct {
  table     string
  where     []string
  args      []any
  returning []string
}

func Delete(table string) *DeleteBuilder {
  return &DeleteBuilder{table: table}
}

func (b *DeleteBuilder) Where(condition string, args ...any) *DeleteBuilder {
  clause, nextArgs := replacePlaceholders(condition, b.args, args)
  b.where = append(b.where, clause)
  b.args = append(b.args, nextArgs...)
  return b
}

func (b *DeleteBuilder) WhereEq(column string, value any) *DeleteBuilder {
  return b.Where(column+" = ?", value)
}

func (b *DeleteBuilder) Returning(columns ...string) *DeleteBuilder {
  b.returning = columns
  return b
}

func (b *DeleteBuilder) Build() (string, []any) {
  if len(b.where) == 0 {
    panic("qb: delete missing where clause")
  }
  query := "delete from " + b.table
  if len(b.where) > 0 {
    query += " where " + strings.Join(b.where, " and ")
  }
  if len(b.returning) > 0 {
    query += " returning " + strings.Join(b.returning, ", ")
  }
  return query, b.args
}

func replacePlaceholders(condition string, existingArgs []any, newArgs []any) (string, []any) {
  if len(newArgs) == 0 {
    return condition, nil
  }

  placeholders := strings.Count(condition, "?")
  if placeholders != len(newArgs) {
    panic(fmt.Sprintf("qb: placeholder count %d does not match args %d", placeholders, len(newArgs)))
  }

  var b strings.Builder
  argIndex := len(existingArgs) + 1

  for i := 0; i < len(condition); i++ {
    if condition[i] == '?' {
      b.WriteString(fmt.Sprintf("$%d", argIndex))
      argIndex++
      continue
    }
    b.WriteByte(condition[i])
  }

  return b.String(), newArgs
}
