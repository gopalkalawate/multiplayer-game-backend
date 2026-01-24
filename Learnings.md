WE shall keep our learnings here 

For Migrations we can use goose
To create a migration file: 
goose -dir ./migrations create players_table sql
To run migrations: 
goose -dir ./migrations sqlite3 storage/storage.db up
To rollback use down

Patterns we can use for real time applications:
1. Hub Pattern for socket connections
2. Reactor Pattern for event handling
3. PubSub for MatchMaking
4. How does a authoritative server work? : [link](https://medium.com/wearemighty/what-are-server-authoritative-realtime-games-e2463db534d1)
