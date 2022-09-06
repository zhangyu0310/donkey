package donkey

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
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
	ErrUnknownDbType          = errors.New("unknown db type")
	ErrNotSupportDbTypeNow    = errors.New("todo")
	ErrGetVariablesFailed     = errors.New("get variables failed")
	ErrDifferentRoutineNum    = errors.New("different routine number")
	ErrEntryNumFileIncomplete = errors.New("entry number file is incomplete")
	ErrEntryNumFileLost       = errors.New("entry number file is lost")
	ErrDatabaseDataLost       = errors.New("database data is lost")
)

var (
	dbs      []*sqlx.DB
	archives []*Archive
	// Go version is too low, not support atomic.Value
	// counter atomic.Value
	counter uint64
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
		archive, err := NewArchive(i)
		if err != nil {
			fmt.Println("Get new archive failed, err:", err)
			return err
		}
		archives = append(archives, archive)
	}
	atomic.StoreUint64(&counter, 0)
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
	sqlStr := fmt.Sprintf("USE %s", cfg.Database)
	for i := range dbs {
		_, err = dbs[i].Exec(sqlStr)
		if err != nil {
			fmt.Printf("Db %d use database failed, err: %s\n", i, err)
			return err
		}
	}
	err = createTestingTable()
	if err != nil {
		return err
	}
	if cfg.InsertData {
		err = execTestingSQL()
		if err != nil {
			return err
		}
	}
	if cfg.CheckData {
		err = checkForCorrectness()
		if err != nil {
			return err
		}
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

func storgeEntryNum(nums []uint64) error {
	f, err := os.OpenFile("./entry_num", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("Open entry number file failed, err:", err)
		return err
	}
	data := make([]byte, 0, 128)
	sizeOfEntries := EncodeFixedUint64(uint64(len(nums)))
	data = append(data, sizeOfEntries[:]...)
	for _, num := range nums {
		entryNum := EncodeFixedUint64(num)
		data = append(data, entryNum[:]...)
	}
	_, err = f.Write(data)
	if err != nil {
		fmt.Println("Write entry number file failed, err:", err)
		return err
	}
	_ = f.Sync()
	return nil
}

func readEntryNum(routineNum uint, quiet bool) ([]uint64, error) {
	f, err := os.OpenFile("./entry_num", os.O_RDONLY, 0666)
	if err != nil {
		if !quiet {
			fmt.Println("Open entry number file failed, err:", err)
		}
		return nil, err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		if !quiet {
			fmt.Println("Read entry number file failed, err:", err)
		}
		return nil, err
	}
	entryNumVec := make([]uint64, 0, routineNum)
	index := 0
	if (len(data) - index) < 8 {
		if !quiet {
			fmt.Println("Entry number file is incomplete")
		}
		return nil, ErrEntryNumFileIncomplete
	}
	numOfEntries := DecodeFixedUint64(GetFixedUint64(data, index))
	if numOfEntries != uint64(routineNum) {
		if !quiet {
			fmt.Println("Routine number is different in two tasks.")
		}
		return nil, ErrDifferentRoutineNum
	}
	index += 8
	for i := 0; i < int(routineNum); i++ {
		if (len(data) - index) < 8 {
			if !quiet {
				fmt.Println("Entry number file is incomplete")
			}
			return nil, ErrEntryNumFileIncomplete
		}
		fixedNum := GetFixedUint64(data, index)
		index += 8
		entryNum := DecodeFixedUint64(fixedNum)
		entryNumVec = append(entryNumVec, entryNum)
	}
	return entryNumVec, nil
}

func execTestingSQL() error {
	cfg := config.GetGlobalConfig()
	// Get max id in testing table (if exist)
	maxId := uint64(0)
	err := dbs[0].QueryRow("SELECT `id` FROM `donkey_test` ORDER BY `id` DESC LIMIT 1").Scan(&maxId)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("Get max insert id from testing table failed, err:", err)
		}
		maxId = 0
	} else {
		maxId++
	}
	atomic.StoreUint64(&counter, maxId)
	// Read entry num file
	originEntryNums, err := readEntryNum(cfg.RoutineNum, true)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if maxId != 0 {
				fmt.Println("Panic: Entry number file is lost!")
				return ErrEntryNumFileLost
			}
		} else {
			fmt.Println("Read entry number file failed, err:", err)
			return err
		}
	} else {
		if maxId == 0 {
			fmt.Println("Panic: Testing database data is lost!")
			return ErrDatabaseDataLost
		}
	}
	// Get update percent
	wg := sync.WaitGroup{}
	wg.Add(int(cfg.RoutineNum))
	tenPercentRowNum := cfg.InsertRows / 10
	// Seek archive for append entries
	for _, archive := range archives {
		err := archive.SeekForAppend()
		if err != nil {
			fmt.Println("Seek for append entries failed, err:", err)
			return err
		}
	}
	// Insert test data to testing database
	for i := 0; i < int(cfg.RoutineNum); i++ {
		go func(routineId int) {
			localCounter := uint64(0)
			for !stop.Load().(bool) {
				if atomic.CompareAndSwapUint64(&counter, localCounter, localCounter+1) {
					insertCount := localCounter - maxId
					if cfg.InsertRows == 0 {
						if insertCount%10000 == 0 {
							fmt.Printf("Insert count: (%d/♾️)\n", insertCount)
						}
					} else {
						if insertCount%tenPercentRowNum == 0 {
							fmt.Printf("Insert progress: %d%% - (%d/%d)\n",
								insertCount/tenPercentRowNum*10, insertCount, cfg.InsertRows)
						}
					}
					if cfg.InsertRows != 0 && insertCount >= cfg.InsertRows {
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
						err = archives[routineId].AppendEntry(&Entry{
							Id:   localCounter,
							Uuid: uuidStr,
						})
						if err != nil {
							fmt.Printf("id: %d, uuid: %s insert success, but append to archive failed\n",
								localCounter, uuidStr)
						}
					}
				} else {
					localCounter = atomic.LoadUint64(&counter)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	// Record number of insert entries.
	entryNumVec := make([]uint64, cfg.RoutineNum)
	for i, archive := range archives {
		entryNumVec[i] = archive.EntityNum
		if originEntryNums != nil {
			entryNumVec[i] += originEntryNums[i]
		}
		archive.Flush()
	}
	err = storgeEntryNum(entryNumVec)
	if err != nil {
		fmt.Println("Entity number storge failed, err:", err)
	}
	return nil
}

func checkForCorrectness() error {
	fmt.Println("Checking...")
	cfg := config.GetGlobalConfig()
	failed := false
	totalRows := uint64(0)
	nowRow := uint64(0)
	entryNums, err := readEntryNum(cfg.RoutineNum, false)
	if err != nil {
		if err == ErrDifferentRoutineNum {
			fmt.Println("Panic: Use different routine num of two tasks. err:", err)
			return err
		} else {
			fmt.Println("Can't get total entry number, will not print progress rate.")
		}
	} else {
		for _, num := range entryNums {
			totalRows += num
		}
	}
	tenPercentRowNum := totalRows / 10

	wg := sync.WaitGroup{}
	wg.Add(int(cfg.RoutineNum))
	for i := 0; i < int(cfg.RoutineNum); i++ {
		go func(routineId int) {
			for {
				localNowRow := atomic.AddUint64(&nowRow, 1)
				if tenPercentRowNum != 0 && localNowRow%tenPercentRowNum == 0 {
					fmt.Printf("Check progress: %d%% - (%d/%d)\n",
						localNowRow/tenPercentRowNum*10, localNowRow, totalRows)
				}
				entry, err := archives[routineId].GetOneEntry()
				if err != nil {
					if err == ErrReadEndOfFile {
						wg.Done()
						break
					} else {
						failed = true
						fmt.Printf("Read archive failed: "+
							"Check routine [%d] read archive file failed, err: %s\n", routineId, err)
						zlog.ErrorF("Read archive failed: "+
							"Check routine [%d] read archive file failed, err: %s", routineId, err)
						wg.Done()
						break
					}
				}
				var data string
				err = dbs[routineId].QueryRow("SELECT `uuid` FROM `donkey_test` WHERE `id`=?", entry.Id).
					Scan(&data)
				if err != nil {
					fmt.Printf("Check failed: Select id %d failed, err: %s\n", entry.Id, err)
					zlog.ErrorF("Check failed: Select routine [%d] id [%d] & uuid [%s] failed, err: %s",
						routineId, entry.Id, entry.Uuid, err)
					failed = true
					continue
				}
				if entry.Uuid != data {
					fmt.Printf("Check failed: id %d uuid different between archive & database\n.", entry.Id)
					zlog.ErrorF("Check failed: id [%d] uuid different between archive & database."+
						" Archive: [%s] Database: [%s]", entry.Id, entry.Uuid, data)
					failed = true
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println()
	if failed {
		fmt.Println("Check failed...")
	} else {
		fmt.Println("Check success!")
	}
	return nil
}
