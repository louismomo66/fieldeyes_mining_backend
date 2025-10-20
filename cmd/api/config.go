package main

import (
	"log"
	"mineral/data"
	"mineral/pkg/email"
	"sync"

	"gorm.io/gorm"
)

type Config struct {
	DB            *gorm.DB
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	Wait          *sync.WaitGroup
	Models        data.Models
	Mailer        email.Mailer
	ErrorChan     chan error
	ErrorChanDone chan bool
}
