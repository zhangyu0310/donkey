package donkey

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	zlog "github.com/zhangyu0310/zlogger"

	"donkey/config"
)

var (
	ErrUnknownDbType       = errors.New("unknown db type")
	ErrNotSupportDbTypeNow = errors.New("todo")
	ErrGetVariablesFailed  = errors.New("get variables failed")
)

var (
	dbs     []*sqlx.DB
	records []map[uint64]string
	counter atomic.Value
	stop    atomic.Value
	esc     chan os.Signal
)

func Initialize() error {
	cfg := config.GetGlobalConfig()
	esc = make(chan os.Signal, 1)
	signal.Notify(esc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-esc
		fmt.Println(sig)
		stop.Store(true)
	}()
	err := zlog.New("./", "donkey_result", false, zlog.LogLevelAll)
	if err != nil {
		fmt.Println("Logger init failed. err:", err)
		return err
	}
	dbType := strings.ToLower(cfg.DbType)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port)
	for i := 0; i < int(cfg.RoutineNum); i++ {
		db, err := sqlx.Open(dbType, dsn)
		if err != nil {
			fmt.Printf("Open testing database failed, err: %s\n", err)
			return err
		}
		dbs = append(dbs, db)
	}
	for i := 0; i < int(cfg.RoutineNum); i++ {
		records = append(records, make(map[uint64]string))
	}
	counterType := uint64(0)
	counter.Store(counterType)
	stop.Store(false)
	return nil
}

func Close() {
	for i := range dbs {
		_ = dbs[i].Close()
	}
}

func Run() error {
	cfg := config.GetGlobalConfig()
	err := execFrontSQL()
	if err != nil {
		return err
	}
	err = createTestingDb()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("USE %s", cfg.Database)
	for i := range dbs {
		_, err = dbs[i].Exec(sql)
		if err != nil {
			fmt.Printf("Db %d use database failed, err: %s\n", i, err)
			return err
		}
	}
	err = createTestingTable()
	if err != nil {
		return err
	}
	err = execTestingSQL()
	if err != nil {
		return err
	}
	err = checkForCorrectness()
	if err != nil {
		return err
	}
	err = execPostSQL()
	if err != nil {
		return err
	}
	return nil
}

func execSQLFile(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Open file %s failed, err: %s\n", fileName, err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Close file %s failed, err: %s\n", fileName, err)
		}
	}(file)

	tx, err := dbs[0].Begin()
	if err != nil {
		fmt.Println("Begin tx failed, err:", err)
		return err
	}
	br := bufio.NewReader(file)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				_ = tx.Rollback()
				fmt.Printf("Read file %s failed, err: %s\n", fileName, err)
				return err
			}
		}
		_, err = tx.Exec(string(line))
		if err != nil {
			_ = tx.Rollback()
			fmt.Printf("Exec file %s SQL failed, err: %s\n", fileName, err)
			return err
		}
	}
	return nil
}

func execFrontSQL() error {
	cfg := config.GetGlobalConfig()
	fileName := cfg.FrontSQL
	if fileName == "" {
		fmt.Println("Don't need exec front sql.")
		return nil
	}
	err := execSQLFile(fileName)
	if err != nil {
		fmt.Println("Exec front sql failed, err:", err)
		return err
	}
	return nil
}

func execPostSQL() error {
	cfg := config.GetGlobalConfig()
	fileName := cfg.PostSQL
	if fileName == "" {
		fmt.Println("Don't need exec post sql.")
		return nil
	}
	err := execSQLFile(fileName)
	if err != nil {
		fmt.Println("Exec post sql failed, err:", err)
		return err
	}
	return nil
}

type MySQLVariable struct {
	Name  string
	Value string
}

func createTestingDb() error {
	cfg := config.GetGlobalConfig()
	switch strings.ToLower(cfg.DbType) {
	case "mysql":
		err := createDbForMySQL()
		if err != nil {
			fmt.Println("Create testing database failed, err:", err)
			return err
		}
	case "postgres":
		// TODO:
		fmt.Println("TODO...")
		return ErrNotSupportDbTypeNow
	default:
		fmt.Println("Unknown database type:", cfg.DbType)
		return ErrUnknownDbType
	}
	return nil
}

func createTestingTable() error {
	cfg := config.GetGlobalConfig()
	switch strings.ToLower(cfg.DbType) {
	case "mysql":
		err := createTableForMySQL()
		if err != nil {
			fmt.Println("Create testing table failed, err:", err)
			return err
		}
	case "postgres":
		// TODO:
		fmt.Println("TODO...")
		return ErrNotSupportDbTypeNow
	default:
		fmt.Println("Unknown database type:", cfg.DbType)
		return ErrUnknownDbType
	}
	return nil
}

func execTestingSQL() error {
	cfg := config.GetGlobalConfig()
	wg := sync.WaitGroup{}
	wg.Add(int(cfg.RoutineNum))
	for i := 0; i < int(cfg.RoutineNum); i++ {
		go func(routineId int) {
			localCounter := uint64(0)
			for !stop.Load().(bool) {
				if counter.CompareAndSwap(localCounter, localCounter+1) {
					if cfg.InsertRows != 0 && localCounter >= cfg.InsertRows {
						stop.Store(true)
						continue
					}
					uuidStr := uuid.New().String()
					_, err := dbs[routineId].Exec(
						"INSERT INTO `donkey_test` (`id`, `uuid`) VALUES (?,?)",
						localCounter, uuidStr)
					if err != nil {
						fmt.Printf("Routine %d insert testing sql failed. err: %s\n", routineId, err)
					} else {
						records[routineId][localCounter] = uuidStr
					}
				} else {
					localCounter = counter.Load().(uint64)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return nil
}

func checkForCorrectness() error {
	fmt.Println("Checking...")
	failed := false
	for i, record := range records {
		for count, uuidStr := range record {
			var data string
			err := dbs[0].QueryRow("SELECT `uuid` FROM `donkey_test` WHERE `id`=?", count).Scan(&data)
			if err != nil {
				fmt.Printf("Check failed: Select id %d failed, err: %s\n", count, err)
				zlog.ErrorF("Check failed: Select routine [%d] id [%d] & uuid [%s] failed, err: %s",
					i, count, uuidStr, err)
				failed = true
				continue
			}
			if uuidStr != data {
				fmt.Printf("Check failed: id %d uuid different between memory & database\n.", count)
				zlog.ErrorF("Check failed: id [%d] uuid different between memory & database."+
					" Memory: [%s] Database: [%s]", count, uuidStr, data)
				failed = true
			}
		}
	}
	fmt.Println()
	if failed {
		fmt.Println("Check failed...")
	} else {
		fmt.Println("Check success!")
	}
	return nil
}
