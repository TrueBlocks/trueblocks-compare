package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

const defaultDatabaseFileName = "database.sqlite"

type Database struct {
	db       *sqlx.DB
	fileName string
}

func NewDatabaseConnection(override bool, dataDir string, fileName string) (d *Database, err error) {
	d = &Database{}
	if fileName != "" {
		d.fileName = fileName
	} else {
		if !override {
			d.fileName = time.Now().Format(time.DateTime) + ".sqlite"
		} else {
			d.fileName = defaultDatabaseFileName
		}
	}
	if dataDir != "" {
		d.fileName = path.Join(dataDir, d.fileName)
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
		"INSERT INTO download_status VALUES(:address, :provider)",
		map[string]interface{}{
			"address":  address,
			"provider": provider,
		},
	)
	return
}

func (d *Database) Downloaded(address string) (providers []string, anyProvider bool, err error) {
	stmt, err := d.db.PrepareNamed(
		`SELECT provider FROM download_status WHERE address = :address`,
	)
	if err != nil {
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

type AppearanceData struct {
	types.Appearance
	BalanceChange bool
}

func (d *Database) SaveAppearance(provider string, appearance AppearanceData) (err error) {

	var appearanceId int
	err = d.db.Get(
		&appearanceId,
		`INSERT INTO appearances(address, block_number, transaction_index, provider) VALUES(?, ?, ?, ?) RETURNING id`,
		appearance.Address.String(),
		appearance.BlockNumber,
		appearance.TransactionIndex,
		provider,
	)
	if err != nil {
		return err
	}

	_, err = d.db.NamedExec(
		`INSERT INTO appearance_reasons(appearance_id, provider, reason) VALUES(:id, :provider, :reason)`,
		map[string]any{
			"id":       appearanceId,
			"provider": provider,
			"reason":   appearance.Reason,
		},
	)
	if err != nil {
		return err
	}

	if appearance.BalanceChange {
		_, err = d.db.NamedExec(
			`INSERT INTO appearance_balance_changes VALUES(:id, true)`,
			map[string]any{
				"id": appearanceId,
			},
		)
		if err != nil {
			return fmt.Errorf("inserting balance change: %w", err)
		}
	}

	return
}

func (d *Database) SaveIncompatibleAddress(address string, appearances []types.Appearance) (err error) {
	_, err = d.db.NamedExec(
		`INSERT INTO incompatible_addresses VALUES(:address, :appearanceCount)`,
		map[string]any{
			"address":         address,
			"appearanceCount": len(appearances),
		},
	)
	return
}

func (d *Database) UniqueAppearanceCount(provider string) (count int, err error) {
	err = d.db.Get(
		&count,
		`SELECT
		count(*)
		FROM view_appearances_with_providers
		WHERE EXISTS (SELECT 1 FROM json_each(providers) WHERE value = ? AND json_array_length(providers) = 1)`,
		provider,
	)

	return
}

func (d *Database) AppearanceCount(provider string) (count int, err error) {
	err = d.db.Get(
		&count,
		`SELECT
		count(*)
		FROM view_appearances_with_providers
		WHERE EXISTS (SELECT 1 FROM json_each(providers) WHERE value = ?)`,
		provider,
	)
	return
}

func (d *Database) AddressCountTotal() (count int, err error) {
	err = d.db.Get(&count, `SELECT count(*) FROM (SELECT DISTINCT address FROM appearances)`)
	return
}

func (d *Database) UniqueAddressCount(provider string) (count int, err error) {
	err = d.db.Get(
		&count,
		`SELECT
		count(*)
		FROM (
			SELECT DISTINCT
			address
			FROM
			view_appearances_with_providers
			WHERE EXISTS (SELECT 1 FROM json_each(providers) WHERE value = ? AND json_array_length(providers) = 1)
		)`,
		provider,
	)

	return
}

func (d *Database) AddressCount(provider string) (count int, err error) {
	err = d.db.Get(
		&count,
		`SELECT
		count(*)
		FROM (
			SELECT DISTINCT
			address
			FROM
			view_appearances_with_providers
			WHERE EXISTS (SELECT 1 FROM json_each(providers) WHERE value = ?)
		)`,
		provider,
	)

	return
}

type GroupedReasons struct {
	Reason string
	Count  int
}

func (d *Database) UniqueAppearancesGroupedReasons(provider string) (groupedReasons []GroupedReasons, err error) {
	err = d.db.Select(
		&groupedReasons,
		`WITH apps AS(
			SELECT * FROM view_appearances_with_providers WHERE EXISTS (SELECT 1 FROM json_each(providers) WHERE value = ? AND json_array_length(providers) = 1)
		) SELECT
		reason,
		count(*) as count
		FROM apps
		JOIN appearance_reasons r ON r.appearance_id = apps.id
		GROUP BY reason ORDER BY count`,
		provider,
	)

	return
}

func (d *Database) BalanceChangeCount(provider string) (count int, err error) {
	err = d.db.Get(
		&count,
		`WITH apps AS(
			SELECT * FROM view_appearances_with_providers WHERE EXISTS (SELECT 1 FROM json_each(providers) WHERE value = ? AND json_array_length(providers) = 1)
		) SELECT
		count(*) as count
		FROM apps
		JOIN appearance_balance_changes c ON c.appearance_id = apps.id`,
		provider,
	)

	return
}
