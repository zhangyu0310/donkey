package donkey

import (
	"fmt"
	"strings"

	"donkey/config"
)

func createDbForMySQL() error {
	cfg := config.GetGlobalConfig()
	var lowerCase bool
	variable := &MySQLVariable{}
	err := dbs[0].QueryRow("SHOW VARIABLES LIKE 'lower_case_table_names'").Scan(&variable.Name, &variable.Value)
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

	rows, err := dbs[0].Query("SHOW DATABASES")
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
		_, err = dbs[0].Exec(sql)
		if err != nil {
			fmt.Println("MySQL create database failed, err:", err)
			return err
		}
	} else {
		fmt.Printf("Database %s is exist, don't need to create again.\n", dbName)
	}
	return nil
}

func createTableForMySQL() error {
	cfg := config.GetGlobalConfig()
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `donkey_test` ("+
		"`id` BIGINT NOT NULL,"+
		"`uuid` CHAR(36) NOT NULL,"+
		"PRIMARY KEY (`id`)"+
		") ENGINE=InnoDB DEFAULT CHARSET=utf8 %s", cfg.UniqueSyntax)
	_, err := dbs[0].Exec(sql)
	if err != nil {
		fmt.Println("MySQL create table failed, err:", err)
		return err
	}
	return nil
}
