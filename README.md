# Golphin - Go based db... 
> its not better than any other db; but u know what, its something I made

This is a project, where i try to build a db from scratch in goLang. my aim? to build postgreSQL; ohh u mean realistically? to reach as near as possible to sqlite. 

## Whats done? 
1. Append only upserting / deletion
2. greedily searching for key in reverse
3. segmenations
4. Compaction (file replacement based)
5. CI

## How to run this project 
- ./golphin set key value
- ./golphin get key
- ./golphin delete key  


## ROADMAP
1. dont know yet.
2. Working on Indexing now: I'll start with creating the index on every run, and then optimize this by saving snapshot in a tmp file

## References
1. https://www.nan.fyi/database
2. https://github.com/golang-standards/project-layout

:love: from [parthkapoor](https://parthkapoor.me)
