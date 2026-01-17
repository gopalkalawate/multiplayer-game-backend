WE shall keep our learnings here 

For Migrations we can use goose
To create a migration file: 
goose -dir ./migrations create players_table sql
To run migrations: 
goose -dir ./migrations sqlite3 storage/storage.db up
To rollback use down