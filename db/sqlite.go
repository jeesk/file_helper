package db

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"time"
)

var db *gorm.DB

func GetDb() *gorm.DB {
	return db
}

func init() {
	fmt.Println("MsgDb 使用的是sqlite")
}

func setUpDb(dbfilePath string) error {
	var err error
	db, err = gorm.Open(sqlite.Open(dbfilePath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(5 * time.Second)
	err = db.AutoMigrate(&Message{})
	if err != nil {
		fmt.Println("合并表失败:" + err.Error())
		return err
	}
	return nil
}

type MsgbackendSqlite struct{}

func (m MsgbackendSqlite) InitSqlite(dbPath string) {
	setUpDb(dbPath)
}

func (msg *MsgbackendSqlite) CreateMessage(meg *Message) error {
	tx := GetDb().Create(meg)
	return tx.Error
}
func (msg *MsgbackendSqlite) QueryMsg(id int) ([]Message, error) {
	var msgs []Message
	tx := GetDb().Where("id > ? ", id).Find(&msgs)
	if tx.Error != nil {
		return msgs, tx.Error
	}
	return msgs, nil
}
