# donkey

A database test tool. Insert data and check correctness.

## Usage

### Params

| Name          | Default   | Description                                         |
|---------------|-----------|-----------------------------------------------------|
| help          | false     | Show usage                                          |
| host          | 127.0.0.1 | Host of testing database                            |
| port          | 3306      | Port of testing database                            |
| user          | root      | User of testing Database                            |
| password      | nil       | Password of testing user                            |
| db            | my_donkey | Database of testing database                        |
| db-type       | mysql     | Type of testing Database                            |
| routine-num   | 0         | Number of testing routine (0/1 both single routine) |
| rows          | 0         | Number of insert rows (0 is infinity)               |
| insert-data   | true      | Insert test data to testing Database                |
| check-data    | true      | Check test data from testing Database               |
| front-SQL     | ""        | SQL file of forward SQL. Running before testing     |
| post-SQL      | ""        | SQL file of post SQL. Running after testing         |
| unique-syntax | ""        | Unique syntax for create table                      |

### Example

```shell
./donkey -host='127.0.0.1' -port=3306 -user='poppinzhang' -password='123456' -routine-num=10 -rows=10000
```
