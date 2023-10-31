package operator

import (
	"donkey/pkg/config"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

var (
	ErrGetVariablesFailed = errors.New("get variables failed")
	ErrDifferentColumnNum = errors.New("column num is different from config")
)

type MySQLVariable struct {
	Name  string
	Value string
}

func CreateDbForMySQL(db *sqlx.DB) error {
	cfg := config.GetGlobalConfig()
	var lowerCase bool
	variable := MySQLVariable{}
	err := db.QueryRow("SHOW VARIABLES LIKE 'lower_case_table_names'").Scan(&variable.Name, &variable.Value)
	if err != nil {
		fmt.Println("MySQL get lower_case_table_names failed, err:", err)
		return err
	}
	if variable.Name == "lower_case_table_names" {
		if variable.Value == "0" {
			lowerCase = false
		} else if variable.Value == "1" {
			lowerCase = true
		} else {
			fmt.Println("FA SHENG SHEN MO SHI LE?")
			return ErrGetVariablesFailed
		}
	} else {
		fmt.Println("FA SHENG SHEN MO SHI LE?")
		return ErrGetVariablesFailed
	}

	var dbName string
	if lowerCase {
		dbName = strings.ToLower(cfg.Database)
	} else {
		dbName = cfg.Database
	}

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		fmt.Println("MySQL get databases failed, err:", err)
		return err
	}
	existDb := false
	for rows.Next() {
		var tmpDbName string
		if err = rows.Scan(&tmpDbName); err != nil {
			fmt.Println("MySQL scan database name failed, err:", err)
			_ = rows.Close()
			return err
		}
		if tmpDbName == dbName {
			existDb = true
			_ = rows.Close()
			break
		}
	}
	if !existDb {
		sql := fmt.Sprintf("CREATE DATABASE %s", cfg.Database)
		_, err = db.Exec(sql)
		if err != nil {
			fmt.Println("MySQL create database failed, err:", err)
			return err
		}
	} else {
		fmt.Printf("Database %s is exist, don't need to create again.\n", dbName)
	}
	return nil
}

func CreateTableForMySQL(db *sqlx.DB) error {
	cfg := config.GetGlobalConfig()
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		fmt.Println("MySQL get tables failed, err:", err)
		return err
	}

	existTable := false
	for rows.Next() {
		var tmpTableName string
		if err = rows.Scan(&tmpTableName); err != nil {
			fmt.Println("MySQL scan database name failed, err:", err)
			_ = rows.Close()
			return err
		}
		if tmpTableName == "donkey_test" {
			existTable = true
			_ = rows.Close()
			break
		}
	}

	if !existTable {
		s := "CREATE TABLE IF NOT EXISTS `donkey_test` (" +
			"`id` BIGINT NOT NULL," +
			"`uuid` CHAR(36) NOT NULL,"
		for i := uint(0); i < cfg.ExtraColumnNum; i++ {
			s += fmt.Sprintf("`uuid_extra_%d` CHAR(36) NOT NULL,", i)
		}
		s += "PRIMARY KEY (`id`)" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8 %s"
		sql := fmt.Sprintf(s, cfg.UniqueSyntax)
		_, err := db.Exec(sql)
		if err != nil {
			fmt.Println("MySQL create table failed, err:", err)
			return err
		}
	} else {
		// Check extra column num
		rows, err := db.Query("DESC `donkey_test`")
		if err != nil {
			fmt.Println("MySQL check table column failed, err:", err)
			return err
		}
		columnCount := uint(0)
		for rows.Next() {
			columnCount++
		}
		if columnCount != cfg.ExtraColumnNum+2 {
			fmt.Printf("Extra column number is different. Testing table is [%d], config is [%d]\n",
				columnCount-2, cfg.ExtraColumnNum)
			return ErrDifferentColumnNum
		}
	}
	return nil
}
