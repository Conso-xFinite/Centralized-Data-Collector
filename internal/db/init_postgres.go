package db

import (
	"Centralized-Data-Collector/pkg/logger"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

var GormDB *gorm.DB

type DB = gorm.DB

func InitGorm(dsn string) error {
	var err error
	GormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gorm_logger.New(
			// log.New(os.Stdout, "\r\n", log.LstdFlags), // 输出到控制台
			&logger.MyGormWriter{},
			gorm_logger.Config{
				SlowThreshold: 100 * time.Millisecond, // 超过此间隔的查询视为慢查询
				LogLevel:      gorm_logger.Warn,       // 只打印警告（慢查询、错误）
				Colorful:      true,                   // 彩色输出
			},
		),
	})
	if err != nil {
		return err
	}
	sqlDB, err := GormDB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(4 * time.Hour)
	return sqlDB.Ping()
}

func CloseGorm() error {
	sqlDB, err := GormDB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func isErrRecordNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound || strings.HasPrefix(err.Error(), "ERROR: operator does not exist: ")
}

// type PostgresDB struct {
// 	*sql.DB
// }

// func NewPostgresDB(dataSourceName string) (*PostgresDB, error) {
// 	db, err := sql.Open("postgres", dataSourceName)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open database: %w", err)
// 	}

// 	if err := db.Ping(); err != nil {
// 		return nil, fmt.Errorf("failed to connect to database: %w", err)
// 	}

// 	return &PostgresDB{db}, nil
// }

// func (p *PostgresDB) Close() error {
// 	return p.DB.Close()
// }

// // Example function to store data
// func (p *PostgresDB) StoreData(query string, args ...interface{}) error {
// 	_, err := p.Exec(query, args...)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute query: %w", err)
// 	}
// 	return nil
// }

// // Example function to retrieve data
// func (p *PostgresDB) RetrieveData(query string, args ...interface{}) (*sql.Rows, error) {
// 	rows, err := p.Query(query, args...)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to execute query: %w", err)
// 	}
// 	return rows, nil
// }
