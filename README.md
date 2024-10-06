# select
method chaining query builder for sqlite with error checking inspired by supabase, and Postgrest

# goal
To create a package for quickly building queries in golang

'mydb.From("table").Select("column, column, join_table(*)").Eq("column", "value").OrderBy("column").Limit(1)'
