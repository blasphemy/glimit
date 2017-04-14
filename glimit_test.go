package glimit

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"testing"
	"time"
)

var (
	limiter *Limiter
	db      *gorm.DB
)

func TestSetupDB(t *testing.T) {
	var err error
	db, err = gorm.Open("sqlite3", "testing.db")
	if err != nil {
		t.Error(err.Error())
	}
	db.LogMode(true)
	DoMigrations(db)
}

func TestSetupLimiter(t *testing.T) {
	var err error
	limiter, err = NewLimiter(2, time.Second*5, db)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestTakeOne(t *testing.T) {
	count, err := limiter.Take()
	if err != nil {
		t.Error(err.Error())
	}
	if count != 1 {
		t.Error("Expected 1 got ", count)
	}
}

func TestTakeTwo(t *testing.T) {
	count, err := limiter.Take()
	if err != nil {
		t.Error(err.Error())
	}
	if count != 2 {
		t.Error("Expected 2 got ", count)
	}
}

func TestShouldExceed(t *testing.T) {
	count, err := limiter.Take()
	if err != ErrRateLimitExceeded {
		t.Error("Was expecting error got ", err)
	}
	if count != 2 {
		t.Error("was expecting 2 got ", count)
	}
}

func TestShouldPass(t *testing.T) {
	time.Sleep(6 * time.Second)
	count, err := limiter.Take()
	if err != nil {
		t.Error(err.Error())
	}
	if count != 1 {
		t.Error("was expecting 1 got ", count)
	}
}

func TestCleanup(t *testing.T) {
	var count int
	err := limiter.Cleanup()
	if err != nil {
		t.Error(err.Error())
	}
	err = db.Model(&Action{}).Where("limiter_id = ?", limiter.ID).Count(&count).Error
	if err != nil {
		t.Error(err.Error())
	}
	if count != 1 {
		t.Error("was expecting 1 got ", count)
	}
}

func TestByID(t *testing.T) {
	l, err := ByID(limiter.ID, db)
	if err != nil {
		t.Error(err.Error())
	}
	if l.Interval != limiter.Interval {
		t.Fail()
	}
	if l.Times != limiter.Times {
		t.Fail()
	}
	if l.ID != limiter.ID {
		t.Fail()
	}
}

func TestDelete(t *testing.T) {
	err := limiter.Delete()
	if err != nil {
		t.Error(err.Error())
	}
	_, err = ByID(limiter.ID, db)
	if err == nil {
		t.Fail()
	}
}

func TestEnsureNoActions(t *testing.T) {
	var count int
	err := db.Model(&Action{}).Where(&Action{LimiterID: limiter.ID}).Count(&count).Error
	if err != nil {
		t.Error(err.Error())
	}
	if count != 0 {
		t.Error("Expected 0 got ", count)
	}
}

func TestInvalidIDDelete(t *testing.T) {
	l := &Limiter{
		ID: "",
		db: db,
	}
	err := l.Delete()
	if err != ErrInvalidID {
		t.Fail()
	}
}

func TestInvalidIDDeleteNotEmpty(t *testing.T) {
	l := &Limiter{
		ID: "TestID",
		db: db,
	}
	err := l.Delete()
	if err != ErrInvalidID {
		t.Fail()
	}
}
