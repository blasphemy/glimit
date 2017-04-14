package glimit

import (
	"errors"
	"github.com/jinzhu/gorm"
	"time"
)

var (
	//ErrRateLimitExceeded Returned when a rate limit is exceeded
	ErrRateLimitExceeded = errors.New("Rate Limit Exceeded")
	//ErrInvalidID occurs when a limiter with an invalid ID is used.
	ErrInvalidID = errors.New("Invalid ID")
)

//Limiter represents the rate limiter
type Limiter struct {
	db       *gorm.DB
	ID       uint
	Times    int
	Interval time.Duration
}

//Action represents the Action being ratelimited.
type Action struct {
	Timestamp time.Time `gorm:"index"`
	ID        uint
	LimiterID uint `gorm:"index"`
}

//DoMigrations adds the models to the database. Should be run once,
// if you choose to use automigrations at all.
func DoMigrations(db *gorm.DB) {
	db.AutoMigrate(&Limiter{})
	db.AutoMigrate(&Action{})
	db.Model(&Action{}).AddForeignKey("limiter_id", "limiters", "RESTRICT", "RESTRICT")
}

//NewLimiter Creates a new limiter with the specified arguments and saves it
// to the database.
func NewLimiter(actions int, interval time.Duration, db *gorm.DB) (*Limiter, error) {
	newLimit := &Limiter{
		db:       db,
		Times:    actions,
		Interval: interval,
	}
	err := db.Save(newLimit).Error
	if err != nil {
		return nil, err
	}
	return newLimit, nil
}

/*Take attempts to do an "action". if there is an error determining if an action
   is possible or writing the action to the database, the returned action count
	 will be 0. If the rate is exceeded, the total count will be returned,
	 as well ass ErrRateLimitExceeded. If the Take is successful, it will return
	 the number of actions and a nil error.
*/
func (l *Limiter) Take() (int, error) {
	if l.ID == 0 {
		return 0, ErrInvalidID
	}
	tx := l.db.Begin()
	limiter := &Limiter{}
	err := tx.Find(limiter, l.ID).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	var count int
	calculated := time.Now().Truncate(limiter.Interval)
	err = tx.Model(&Action{}).Where("timestamp >= ? AND limiter_id = ?",
		calculated,
		limiter.ID).Count(&count).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if count >= limiter.Times {
		tx.Rollback()
		return count, ErrRateLimitExceeded
	}
	err = tx.Save(&Action{
		Timestamp: time.Now(),
		LimiterID: limiter.ID,
	}).Error
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	err = tx.Commit().Error
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

//Cleanup deletes all "expired" actions associated with a limiter
func (l *Limiter) Cleanup() error {
	if l.ID == 0 {
		return ErrInvalidID
	}
	limiter := &Limiter{}
	err := l.db.Find(limiter, l.ID).Error
	if err != nil {
		return err
	}
	calculated := time.Now().Truncate(limiter.Interval)
	err = l.db.Where("timestamp <= ? AND limiter_id = ?",
		calculated,
		limiter.ID).Delete(Action{}).Error
	return err
}

//Delete deletes the limiter as well as all associated actions
func (l *Limiter) Delete() error {
	if l.ID == 0 {
		return ErrInvalidID
	}
	err := l.db.Delete(Action{}, &Action{LimiterID: l.ID}).Error
	if err != nil {
		return err
	}
	err = l.db.Delete(Limiter{}, l.ID).Error
	return err
}

//ByID retrieves a limiter from it's ID
func ByID(ID uint, db *gorm.DB) (*Limiter, error) {
	l := &Limiter{}
	err := db.First(l, ID).Error
	if err != nil {
		return nil, err
	}
	l.db = db
	return l, nil
}

//Save allows you to update the attributes of a limiter. If you make any
// changes to a limiter, simply Save() it, and it will go back to the database
func (l *Limiter) Save() error {
	if l.ID == 0 {
		return ErrInvalidID
	}
	return l.db.Model(&Limiter{}).Where(l.ID).Update(l).Error
}

//CleanupAll removes all expired actions from all limiters in the database.
// possibly a very expensive call.
func CleanupAll(db *gorm.DB) error {
	limiters := []Limiter{}
	err := db.Find(&limiters).Error
	if err != nil {
		return err
	}
	for _, x := range limiters {
		x.db = db
		err = x.Cleanup()
		if err != nil {
			return err
		}
	}
	return nil
}
