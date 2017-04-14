package glimit

import (
	"errors"
	"github.com/google/uuid"
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
	ID       string
	Times    int
	Interval time.Duration
}

//Action represents the Action being ratelimited.
type Action struct {
	Timestamp time.Time `gorm:"index"`
	ID        string
	LimiterID string `gorm:"index"`
}

//DoMigrations adds the models to the database. Should be run once,
// if you choose to use automigrations at all.
func DoMigrations(db *gorm.DB) {
	db.AutoMigrate(&Limiter{})
	db.AutoMigrate(&Action{})
}

//NewLimiter Creates a new limiter with the specified arguments and saves it
// to the database.
func NewLimiter(actions int, interval time.Duration, db *gorm.DB) (*Limiter, error) {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	newLimit := &Limiter{
		db:       db,
		ID:       newUUID.String(),
		Times:    actions,
		Interval: interval,
	}
	err = db.Save(newLimit).Error
	if err != nil {
		return nil, err
	}
	return newLimit, nil
}

//Take attempts to do an "action". if there is an error determining if an action is possible
// or writing the action to the database, the returned action count will be 0.
// If the rate is exceeded, the total count will be returned, as well as ErrRateLimitExceeded.
// If the Take is successful, it will return the amount of actions within the period,
// and no errors.
func (l *Limiter) Take() (int, error) {
	_, err := uuid.Parse(l.ID)
	if err != nil {
		return 0, ErrInvalidID
	}
	var count int
	calculated := time.Now().Truncate(l.Interval)
	err = l.db.Model(&Action{}).Where("timestamp >= ? AND limiter_id = ?", calculated, l.ID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	if count >= l.Times {
		return count, ErrRateLimitExceeded
	}
	newActionID, err := uuid.NewUUID()
	if err != nil {
		return 0, err
	}
	newAction := &Action{
		ID:        newActionID.String(),
		Timestamp: time.Now(),
		LimiterID: l.ID,
	}
	err = l.db.Save(&newAction).Error
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

//Cleanup deletes all "expired" actions associated with a limiter
func (l *Limiter) Cleanup() error {
	_, err := uuid.Parse(l.ID)
	if err != nil {
		return ErrInvalidID
	}
	calculated := time.Now().Truncate(l.Interval)
	err = l.db.Where("timestamp <= ? AND limiter_id = ?", calculated, l.ID).Delete(Action{}).Error
	return err
}

//Delete deletes the limiter as well as all associated actions
func (l *Limiter) Delete() error {
	_, err := uuid.Parse(l.ID)
	if err != nil {
		return ErrInvalidID
	}
	err = l.db.Delete(Action{}, &Action{LimiterID: l.ID}).Error
	if err != nil {
		return err
	}
	err = l.db.Delete(Limiter{}, &Limiter{ID: l.ID}).Error
	return err
}

//ByID retrieves a limiter from it's ID
func ByID(ID string, db *gorm.DB) (*Limiter, error) {
	_, err := uuid.Parse(ID)
	if err != nil {
		return nil, ErrInvalidID
	}
	l := &Limiter{}
	err = db.First(l, &Limiter{ID: ID}).Error
	if err != nil {
		return nil, err
	}
	l.db = db
	return l, nil
}
