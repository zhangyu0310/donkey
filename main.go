package main

import (
	"flag"
	"fmt"
	"os"

	"donkey/config"
	"donkey/donkey"
)

var (
	dbType         = flag.String("db-type", "mysql", "Type of testing Database")
	host           = flag.String("host", "127.0.0.1", "Host of testing Database")
	port           = flag.Int("port", 3306, "Port of testing Database")
	user           = flag.String("user", "root", "User of testing Database")
	pass           = flag.String("password", "", "Password of testing Database")
	database       = flag.String("db", "my_donkey", "Database of testing Database")
	insertRows     = flag.Uint64("rows", 0, "Number of insert rows. (0 is infinity)")
	frontSQL       = flag.String("front-SQL", "", "SQL file of forward SQL. Running before testing")
	postSQL        = flag.String("post-SQL", "", "SQL file of post SQL. Running after testing")
	uniqueSyntax   = flag.String("unique-syntax", "", "Unique syntax for create table")
	routineNum     = flag.Uint("routine-num", 0, "Number of testing routine. (0/1 both single routine)")
	insertData     = flag.Bool("insert-data", true, "Insert test data to testing Database")
	checkData      = flag.Bool("check-data", true, "Check test data from testing Database")
	insertPackage  = flag.Uint("insert-package", 0, "Number of rows in once insert. (0/1 both single row)")
	extraColumnNum = flag.Uint("extra-column-num", 0, "Testing table extra column number")
	insertDelay    = flag.Int64("insert-delay", 0, "Insert delay. (ms)")
	timeConsume    = flag.Bool("time-consume", false, "Print time consume. (s)")
)

func cmdConfigSetToGlobal(cfg *config.Config) {
	cfg.DbType = *dbType
	cfg.Host = *host
	cfg.Port = *port
	cfg.User = *user
	cfg.Pass = *pass
	cfg.Database = *database
	cfg.InsertRows = *insertRows
	cfg.FrontSQL = *frontSQL
	cfg.PostSQL = *postSQL
	cfg.UniqueSyntax = *uniqueSyntax
	if *routineNum == 0 {
		cfg.RoutineNum = 1
	} else {
		cfg.RoutineNum = *routineNum
	}
	cfg.InsertData = *insertData
	cfg.CheckData = *checkData
	if *insertPackage == 0 {
		cfg.InsertPackage = 1
	} else {
		cfg.InsertPackage = *insertPackage
	}
	cfg.ExtraColumnNum = *extraColumnNum
	cfg.InsertDelay = *insertDelay
	cfg.TimeConsume = *timeConsume
}

func main() {
	help := flag.Bool("help", false, "show usage")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *pass == "" {
		fmt.Println("Database password must be input.")
		os.Exit(1)
	}
	config.InitializeConfig(cmdConfigSetToGlobal)

	err := donkey.Initialize()
	if err != nil {
		fmt.Println("Init donkey failed, err:", err)
		os.Exit(1)
	}
	defer donkey.Close()
	err = donkey.Run()
	if err != nil {
		fmt.Println("Run donkey failed, err:", err)
		os.Exit(1)
	}
}
