package main

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// It's called a model, which is a database table.
type Log struct {
	ID         uint      // PK
	Time       time.Time `gorm:"index"`
	Msg        string
	Level      int8
	LogDetails []LogDetail // one-to-many
}

type LogDetail struct {
	ID        uint // PK
	LogID     uint // FK referencing Log
	DetailMsg string
}

// Hooks - BeforeSave, BeforeCreate, AfterSave, AfterCreate.
func (u *Log) BeforeCreate(tx *gorm.DB) (err error) {
	fmt.Println("BeforeCreate", u.Msg)
	return nil
}

func main() {
	db, _ := gorm.Open(sqlite.Open("log.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	// CREATE TABLE and CREATE INDEX for each model.
	migrate := func() {
		db.AutoMigrate(&Log{}, &LogDetail{})
	}
	migrate()

	// INSERT INTO `logs` (`time`,`msg`,`level`) VALUES (...) RETURNING `id`
	insert := func() {
		log := Log{Time: time.Now(), Msg: "welcome!"}
		db.Create(&log)
	}
	insert()

	// INSERT INTO `logs` (`msg`,`level`) VALUES ("wow!",3) RETURNING `id`
	insertSelectedFields := func() {
		log := Log{Time: time.Now(), Msg: "wow!", Level: 3}
		db.
			Select("Msg", "Level").
			Create(&log)
	}
	insertSelectedFields()

	// INSERT INTO `logs` (`time`,`msg`,`level`) VALUES (...),(...) RETURNING `id`
	// INSERT INTO `logs` (`time`,`msg`,`level`) VALUES (...) RETURNING `id`
	insertInBatches := func() {
		logs := []Log{{Msg: "a"}, {Msg: "b"}, {Msg: "c"}}
		db.CreateInBatches(&logs, 2)
	}
	insertInBatches()

	// INSERT ... ON CONFLICT (`id`) DO UPDATE SET ... RETURNING `id`
	upsert := func() {
		log := Log{ID: 1, Time: time.Now(), Msg: "welcome!"}
		db.
			Clauses(clause.OnConflict{UpdateAll: true}).
			Create(&log)
	}
	upsert()

	// SELECT * FROM `logs` ORDER BY `logs`.`id` LIMIT 1
	// SELECT * FROM `logs` ORDER BY `logs`.`id` DESC LIMIT 1
	firstOrLast := func() {
		log := Log{}
		db.First(&log)
		log = Log{} // To throw away the saved PK.
		db.Last(&log)
	}
	firstOrLast()

	// SELECT * FROM `logs` ORDER BY `logs`.`id` LIMIT 1
	firstToMap := func() {
		logMap := map[string]interface{}{}
		db.
			Model(&Log{}).
			First(&logMap)
	}
	firstToMap()

	// SELECT * FROM `logs` WHERE `logs`.`id` = 100000000 ORDER BY `logs`.`id` LIMIT 1
	firstButNotFound := func() {
		log := Log{ID: 100000000}
		result := db.First(&log)
		fmt.Println(result.Error)                                    // record not found
		fmt.Println(errors.Is(result.Error, gorm.ErrRecordNotFound)) // true
	}
	firstButNotFound()

	// SELECT * FROM `logs`
	selectAll := func() {
		log := Log{}
		db.Find(&log)
	}
	selectAll()

	// SELECT * FROM `logs` LIMIT 2 OFFSET 3
	selectWithLimitAndOffset := func() {
		log := Log{}
		db.
			Limit(2).
			Offset(3).
			Find(&log)
	}
	selectWithLimitAndOffset()

	// SELECT * FROM `logs` WHERE `logs`.`id` IN (1,2,3)
	selectByPK1 := func() {
		logs := []Log{}
		db.Find(&logs, []int{1, 2, 3})
	}
	selectByPK1()

	// SELECT * FROM `logs` WHERE `logs`.`id` IN (1,2,3)
	selectByPK2 := func() {
		logs := []Log{}
		db.
			Where([]int{1, 2, 3}).
			Find(&logs)
	}
	selectByPK2()

	// SELECT * FROM `logs` WHERE msg LIKE "%wel%" AND id >= 1
	selectWithCondition := func() {
		logs := []Log{}
		db.
			Where("msg LIKE ? AND id >= ?", "%wel%", 1).
			Find(&logs)
	}
	selectWithCondition()

	// SELECT * FROM `logs` WHERE msg IN ("a","b")
	selectWithIN := func() {
		logs := []Log{}
		db.
			Where("msg IN ?", []string{"a", "b"}).
			Find(&logs)
	}
	selectWithIN()

	// SELECT * FROM `logs` WHERE `logs`.`msg` = "x"
	selectWithStruct := func() {
		logs := []Log{}
		db.
			Where(&Log{Msg: "x"}). // Zero values have no effect.
			Find(&logs)
	}
	selectWithStruct()

	// SELECT * FROM `logs` WHERE `logs`.`msg` <> "x"
	selectWithNotStruct := func() {
		logs := []Log{}
		db.
			Not(&Log{Msg: "x"}). // Zero values have no effect.
			Find(&logs)
	}
	selectWithNotStruct()

	// SELECT * FROM `logs` WHERE `msg` = "y"
	selectWithMap := func() {
		logs := []Log{}
		db.
			Where(map[string]interface{}{"msg": "y"}).
			Find(&logs)
	}
	selectWithMap()

	// SELECT * FROM `logs` WHERE id = 1 OR `logs`.`id` = 2 OR `id` = 3
	selectWithOr := func() {
		logs := []Log{}
		db.
			Where("id = ?", 1).
			Or(&Log{ID: 2}).
			Or(map[string]interface{}{"id": 3}).
			Find(&logs)
	}
	selectWithOr()

	// SELECT `msg`,`level` FROM `logs`
	selectSomeFieldsOnly := func() {
		logs := []Log{}
		db.
			Select("msg", "level").
			Find(&logs)
	}
	selectSomeFieldsOnly()

	// SELECT * FROM `logs` ORDER BY msg desc, level
	selectWithOrderBy := func() {
		logs := []Log{}
		db.
			Order("msg desc, level").
			Find(&logs)
	}
	selectWithOrderBy()

	// SELECT count(*) FROM `logs` WHERE msg LIKE "%wel%"
	count := func() {
		c := int64(0)
		db.
			Model(&Log{}).
			Where("msg LIKE ?", "%wel%").
			Count(&c)
	}
	count()

	// SELECT level as lev, cound(id) as tot FROM `logs` GROUP BY `level` HAVING lev >= 3
	groupBy := func() {
		type groupByResultRow struct {
			Lev int8
			Tot int64
		}
		groupByResultRows := []groupByResultRow{}
		db.
			Model(&Log{}).
			Select("level as lev, cound(id) as tot").
			Group("level").
			Having("lev >= ?", 3).
			Find(&groupByResultRows)
	}
	groupBy()

	// SELECT DISTINCT `msg`,`level` FROM `logs`
	distinct := func() {
		logs := []Log{}
		db.
			Distinct("msg", "level").
			Find(&logs)
	}
	distinct()

	// SELECT * FROM `log_details` WHERE `log_details`.`log_id` IN (1,2,3,4,5)
	// SELECT * FROM `logs` WHERE id <= 5
	preload := func() {
		log := Log{}
		db.First(&log)

		logDetails := []LogDetail{
			{LogID: log.ID, DetailMsg: "detail 1"},
			{LogID: log.ID, DetailMsg: "detail 2"},
		}
		db.Create(&logDetails)
		fmt.Println(len(log.LogDetails)) // Zero

		logs := []Log{}
		db.
			Where("id <= ?", 5).
			Find(&logs)
		fmt.Println(len(logs[0].LogDetails)) // Zero

		logs = []Log{}
		db.
			Preload("LogDetails").
			Where("id <= ?", 5).
			Find(&logs)
		fmt.Println(len(logs[0].LogDetails)) // Non-zero
	}
	preload()

	// SELECT log_details.id AS log_detail_id, logs.id AS log_id
	// FROM `log_details` LEFT JOIN logs ON logs.id = log_details.log_id
	join := func() {
		type joinResultRow struct {
			LogDetailID uint
			LogID       uint
		}
		joinResultRows := []joinResultRow{}
		db.
			Model(&LogDetail{}).
			Select("log_details.id AS log_detail_id, logs.id AS log_id").
			Joins("LEFT JOIN logs ON logs.id = log_details.log_id").
			Find(&joinResultRows)
	}
	join()

	// SELECT `logs`.`id`,`logs`.`msg` FROM `logs` ORDER BY `logs`.`id` LIMIT 1
	selectWithSubsetStruct := func() {
		type LogSubset struct {
			ID  uint
			Msg string
		}
		logSubset := LogSubset{}
		db.
			Model(&Log{}).
			First(&logSubset)
	}
	selectWithSubsetStruct()

	// SELECT * FROM `logs` ORDER BY `logs`.`id` LIMIT 1 FOR UPDATE
	selectForUpdate := func() {
		log := Log{}
		db.
			Clauses(clause.Locking{Strength: "UPDATE"}). // No effect on Sqlite.
			First(&log)
	}
	selectForUpdate()

	// INSERT INTO `logs` (`time`,`msg`,`level`) VALUES ("0000-00-00 00:00:00","xxx",0) RETURNING `id`
	// SELECT * FROM `logs` WHERE `logs`.`msg` = "xxx" ORDER BY `logs`.`id` LIMIT 1
	selectOrInsert := func() {
		log := Log{Msg: "xxx"}
		db.
			Where(&log).
			FirstOrCreate(&log)
	}
	selectOrInsert()

	// UPDATE `logs` SET `time`="2022-10-20 11:54:03.206",`msg`="welcome!",`level`=0
	// WHERE `id` = 1
	updateBySave := func() {
		log := Log{}
		db.First(&log)
		log.Time = time.Now()
		db.Save(&log)
	}
	updateBySave()

	// UPDATE `logs` SET `time`="2022-10-20 11:55:51.599" WHERE `logs`.`id` = 1
	updateColumn := func() {
		db.
			Model(&Log{}).
			Where(&Log{ID: 1}).
			Update("time", time.Now())
	}
	updateColumn()

	// UPDATE `logs` SET `level`=9,`time`="2022-10-20 11:58:33.06" WHERE `logs`.`id` = 1
	updateMultipleColumns := func() {
		db.
			Model(&Log{}).
			Where(&Log{ID: 1}).
			Updates(map[string]interface{}{"time": time.Now(), "level": 9})
	}
	updateMultipleColumns()

	// UPDATE `logs` SET `level`=level + 1 WHERE `logs`.`id` = 1
	updateUsingExpression := func() {
		db.
			Model(&Log{}).
			Where(&Log{ID: 1}).
			Updates(map[string]interface{}{"level": gorm.Expr("level + ?", 1)})
	}
	updateUsingExpression()

	// UPDATE `logs` SET `level`=9,`time`="2022-10-20 12:02:13.149"
	// WHERE id BETWEEN 1 AND 10 RETURNING `msg`,`level`
	updateAndReturn := func() {
		logs := []Log{}
		columnsToReturn := []clause.Column{
			{Name: "msg"},
			{Name: "level"},
		}
		db.
			Model(&logs). // The RETURNING is done thorugh logs.
			Clauses(clause.Returning{Columns: columnsToReturn}).
			Where("id BETWEEN ? AND ?", 1, 10).
			Updates(map[string]interface{}{"time": time.Now(), "level": 9})
	}
	updateAndReturn()

	// DELETE FROM `logs` WHERE `logs`.`id` = 1
	deleteButRollback := func() {
		db.Transaction(func(tx *gorm.DB) error {
			tx.Delete(&Log{ID: 1})
			return errors.New("rollback deletion") // nil to commit. (https://bityl.co/FABV)
		})
		log := Log{}
		result := db.Where(&Log{ID: 1}).First(&log)
		fmt.Println(result.RowsAffected) // 1
	}
	deleteButRollback()
}
