// Package store persiste contas e lançamentos via GORM.
package store

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"financas-backend/internal/model"
)

// Store é a interface usada pelos handlers.
type Store interface {
	CreateAccount(a *model.Account) error
	ListAccounts() ([]model.Account, error)
	GetAccount(id uint) (model.Account, error)
	UpdateAccount(a *model.Account) error
	DeleteAccount(id uint) error
	CreateTransaction(t *model.Transaction) error
	ListTransactions(accountID uint) ([]model.Transaction, error)
	DeleteTransaction(id uint) error
	AllTransactions() ([]model.Transaction, error)
}

// GormStore implementa Store com GORM.
type GormStore struct {
	db *gorm.DB
}

// Open conecta no Postgres e roda AutoMigrate.
func Open(dsn string) (*GormStore, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&model.Account{}, &model.Transaction{}); err != nil {
		return nil, err
	}
	return &GormStore{db: db}, nil
}

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("não encontrada")
	}
	return err
}

func (s *GormStore) CreateAccount(a *model.Account) error {
	return s.db.Create(a).Error
}

func (s *GormStore) ListAccounts() ([]model.Account, error) {
	var out []model.Account
	err := s.db.Order("id").Find(&out).Error
	return out, err
}

func (s *GormStore) GetAccount(id uint) (model.Account, error) {
	var a model.Account
	err := s.db.First(&a, id).Error
	return a, notFound(err)
}

func (s *GormStore) UpdateAccount(a *model.Account) error {
	res := s.db.Model(&model.Account{}).Where("id = ?", a.ID).Updates(map[string]any{"name": a.Name, "currency": a.Currency})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("não encontrada")
	}
	return nil
}

func (s *GormStore) DeleteAccount(id uint) error {
	res := s.db.Delete(&model.Account{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("não encontrada")
	}
	return s.db.Where("account_id = ?", id).Delete(&model.Transaction{}).Error
}

func (s *GormStore) CreateTransaction(t *model.Transaction) error {
	return s.db.Create(t).Error
}

func (s *GormStore) ListTransactions(accountID uint) ([]model.Transaction, error) {
	var out []model.Transaction
	q := s.db.Order("date desc, id desc")
	if accountID != 0 {
		q = q.Where("account_id = ?", accountID)
	}
	err := q.Find(&out).Error
	return out, err
}

func (s *GormStore) DeleteTransaction(id uint) error {
	res := s.db.Delete(&model.Transaction{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("não encontrada")
	}
	return nil
}

func (s *GormStore) AllTransactions() ([]model.Transaction, error) {
	var out []model.Transaction
	err := s.db.Order("id").Find(&out).Error
	return out, err
}
