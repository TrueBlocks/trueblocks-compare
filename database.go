package main

import (
	"database/sql"
	"os"
	"strings"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	_ "modernc.org/sqlite"
)

const defaultDatabaseFileName = "database.sqlite"

type Database struct {
	db       *sql.DB
	fileName string
	readOnly bool
}

func NewDatabaseConnection(override bool) (d *Database, err error) {
	d = &Database{}
	if !override {
		d.fileName = time.Now().Format(time.DateTime) + ".sqlite"
	} else {
		d.fileName = defaultDatabaseFileName
	}
	if err = d.openDatabase(); err != nil {
		return
	}
	err = d.PrepareTables()
	return
}

func (d *Database) openDatabase() (err error) {
	d.db, err = sql.Open("sqlite", d.fileName)
	return
}

func (d *Database) PrepareTables() (err error) {
	sqlFile, err := os.ReadFile("create_tables.sql")
	if err != nil {
		return
	}

	_, err = d.db.Exec(string(sqlFile))

	return
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) MarkAsDownloaded(address string, provider string) (err error) {
	_, err = d.db.Exec(
		"insert into download_status values(@address, @provider);",
		sql.Named("address", address),
		sql.Named("provider", provider),
	)
	return
}

// func (d *Database) Downloaded(address string, provider string) (downloaded bool, err error) {
// 	err = d.db.QueryRow(
// 		`select case when exists (
// 			select * from download_status where address = @address and provider = @provider
// 		) then 'TRUE' else 'FALSE' end;`,
// 		sql.Named("address", address),
// 		sql.Named("provider", provider),
// 	).Scan(&downloaded)

// 	return
// }

func (d *Database) Downloaded(address string) (providers []string, anyProvider bool, err error) {
	var foundProviders string
	err = d.db.QueryRow(
		`select provider from download_status where address = @address;`,
		sql.Named("address", address),
	).Scan(&foundProviders)
	if len(foundProviders) > 0 {
		anyProvider = true
	}
	providers = strings.Split(foundProviders, ",")
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func (d *Database) SaveAppearances(provider string, appearances []types.Appearance) (err error) {
	dbTx, err := d.db.Begin()
	if err != nil {
		return
	}
	for _, appearance := range appearances {
		_, err = dbTx.Exec(
			`insert into appearances values(@address, @blockNumber, @txIndex, @provider);`,
			sql.Named("address", appearance.Address.String()),
			sql.Named("blockNumber", appearance.BlockNumber),
			sql.Named("txIndex", appearance.TransactionIndex),
			sql.Named("provider", provider),
		)
		if err != nil {
			return
		}
	}
	err = dbTx.Commit()
	return
}
