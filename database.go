package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const defaultDatabaseFileName = "database.sqlite"

type Database struct {
	db       *sqlx.DB
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
	d.db, err = sqlx.Open("sqlite", d.fileName)
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
	_, err = d.db.NamedExec(
		"insert into download_status values(:address, :provider)",
		map[string]interface{}{
			"address":  address,
			"provider": provider,
		},
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
	stmt, err := d.db.PrepareNamed(
		`select provider from download_status where address = :address`,
	)
	if err != nil {
		// if err == sql.ErrNoRows {
		// 	err = nil
		// }
		return
	}
	err = stmt.Select(&providers, map[string]interface{}{
		"address": address,
	})
	if err != nil {
		return
	}
	if len(providers) > 0 {
		anyProvider = true
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

type dbAppearance struct {
	Address          string `db:"address"`
	BlockNumber      int32  `db:"block_number"`
	TransactionIndex int32  `db:"transaction_index"`
}

func (d *Database) AppearancesByProviders(providers []string) (appearances []types.Appearance, err error) {
	args := make([]any, 0, len(providers)+1)
	rawSql := `select
		address,
		block_number,
		transaction_index
		from view_appearances_with_providers
		where exists (select 1 from json_each(providers) where value = ?`

	for i := 0; i < len(providers); i++ {
		if i > 0 {
			rawSql += "or ?"
		}
		args = append(args, providers[i])
	}
	args = append(args, len(providers))

	rawSql += `) and json_array_length(providers) = ?;`

	log.Println(rawSql)
	log.Println(args...)

	raws := []dbAppearance{}
	err = d.db.Select(
		&raws,
		rawSql,
		args...,
	)

	for _, raw := range raws {
		appearance := types.Appearance{
			Address:          base.HexToAddress(raw.Address),
			BlockNumber:      uint32(raw.BlockNumber),
			TransactionIndex: uint32(raw.TransactionIndex),
		}
		appearances = append(appearances, appearance)
	}

	return

}
