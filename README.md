# Golphin - Go based db... 
> its not better than any other db; but u know what, its something I made

This is a project, where i try to build a db from scratch in goLang. my aim? to build postgreSQL; ohh u mean realistically? to reach as near as possible to sqlite. 

Benchmarks: [/dev/bench](https://parthkapoor-dev.github.io/golphin/dev/bench/)

## Whats done? 
1. Append only upserting / deletion
2. greedily searching for key in reverse
3. segmenations
4. Compaction (file replacement based)
5. CI (tests & benchmark tracking)
6. Indexing & Snapshots

## How to run this project 
- ./golphin set key value
- ./golphin get key
- ./golphin delete key  

## ROADMAP
1. dont know yet.
2. [DONE] Working on Indexing now: I'll start with creating the index on every run, and then optimize this by saving snapshot in a tmp file
3. TUI for state to live in memory; currently cli makes the memory 
4. Create a daemon (background running service), like a server maybe -- note this is postgres
5. How does sqlite works? its serverless u know.

## STORAGE ENGINE BENCH

### Hashmap Complexities:
Currently we have an append only disk storage + hashmap cache ie. 
| Operation | Time Complexite | Space Complexity | Storage type dependent |
| --------- | --------------- | ---------------- | ---------------------- |
| Insert    | O(1)            | O(1)             | DISK                   |          
| Update    | O(1)            | O(1)             | DISK                   |
| Delete    | O(1)            | O(1)             | DISK                   |
| Read      | O(1)            | O(1)             | Cache (in-memory)      |
| Range Read| Θ(n)            | O(1)             | Cache (in-memory)      |
| Compaction| O(n)            | O(1)             | DISK                   |

###  Complexities:
since this is not sorted we are now introducing Binary Search Trees
| Operation | Time Complexite | Space Complexity | Storage type dependent |
| --------- | --------------- | ---------------- | ---------------------- |
| Insert    | O(logn)         | O(1)             | DISK                   |          
| Update    | O(logn)         | O(1)             | DISK                   |
| Delete    | O(logn)         | O(1)             | DISK                   |
| Read      | O(logn)         | O(1)             | Cache (in-memory)      |
| Range Read| Θ(2 * logn)     | O(1)             | Cache (in-memory)      |
| Compaction| O(n)            | O(1)             | DISK                   |


## References
1. https://www.nan.fyi/database
2. https://github.com/golang-standards/project-layout
3. https://github.com/egregors/sortedmap

:love: from [parthkapoor](https://parthkapoor.me)
