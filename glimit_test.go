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

func TestNilTake(t *testing.T) {
	l := &Limiter{}
	count, err := l.Take()
	if err == nil {
		t.Fail()
	}
	if count > 0 {
		t.Fail()
	}
}

func TestNilDelete(t *testing.T) {
	l := &Limiter{}
	err := l.Delete()
	if err == nil {
		t.Fail()
	}
}

func TestNilCleanup(t *testing.T) {
	l := &Limiter{}
	err := l.Cleanup()
	if err == nil {
		t.Fail()
	}
}

func TestNilSave(t *testing.T) {
	l := &Limiter{}
	err := l.Save()
	if err == nil {
		t.Fail()
	}
}

func TestUpdate(t *testing.T) {
	var err error
	limiter, err = NewLimiter(5, 10*time.Second, db)
	if err != nil {
		t.Fail()
	}
	limiter.Interval = 20 * time.Second
	err = limiter.Save()
	if err != nil {
		t.Fail()
	}
	limiter2 := &Limiter{}
	err = db.Find(limiter2, limiter.ID).Error
	if err != nil {
		t.Fail()
	}
	if limiter2.Interval != 20*time.Second {
		t.Fail()
	}
}

func TestMulti(t *testing.T) {
	var err error
	limiter, err = NewLimiter(5, 5*time.Minute, db)
	if err != nil {
		t.Error(err.Error())
	}
	for i := 0; i < 25; i++ {
		go func() {
			time.Sleep(time.Second * 1)
			limiter.Take()
		}()
	}
	time.Sleep(20 * time.Second)
	count, err := limiter.Take()
	if count > 5 {
		t.Errorf("count should be less than 5, is %d", count)
	}
	if err != nil {
		if err != ErrRateLimitExceeded {
			t.Fail()
		}
	}
}

func TestCleanupAll(t *testing.T) {
	err := CleanupAll(limiter.db)
	if err != nil {
		t.Fail()
	}
}
