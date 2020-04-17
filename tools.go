package common

import (
	"fmt"
	"io/ioutil"
	"os"
)

func CheckDir(dir string) error {
	if f, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	} else if !f.IsDir() {
		if err := os.Remove(dir); err != nil {
			return err
		}
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func TryPrintLogo(logoPath string) {
	if logoPath == "" {
		logoPath = "logo.txt"
	}
	f, err := ioutil.ReadFile(logoPath)
	if err != nil {
		return
	}
	fmt.Println(string(f))
}

type (
	Database struct {
		Addr     string
		User     string
		Password string
		Database string
	}
)

func (dbc *Database) GetConfig() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", dbc.User, dbc.Password, dbc.Addr, dbc.Database)

}
