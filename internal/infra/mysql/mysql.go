package mysql

import (
	"fmt"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewConn() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", env.AppEnv.DBUser, env.AppEnv.DBPass, env.AppEnv.DBHost, env.AppEnv.DBPort, env.AppEnv.DBName)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
