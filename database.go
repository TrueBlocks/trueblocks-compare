package main

import (
	"fmt"
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

func NewDatabaseConnection(override bool, fileName string) (d *Database, err error) {
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

type AppearanceData struct {
	types.Appearance
	BalanceChange bool
}

// func (d *Database) SaveAppearances(provider string, appearances []AppearanceData) (err error) {
// 	// dbTx, err := d.db.Begin()
// 	// if err != nil {
// 	// 	return
// 	// }
// 	// for _, appearance := range appearances {
// 	// 	_, err = dbTx.Exec(
// 	// 		`insert into appearances values(@address, @blockNumber, @txIndex, @provider) returning id`,
// 	// 		sql.Named("address", appearance.Address.String()),
// 	// 		sql.Named("blockNumber", appearance.BlockNumber),
// 	// 		sql.Named("txIndex", appearance.TransactionIndex),
// 	// 		sql.Named("provider", provider),
// 	// 	)
// 	// 	if err != nil {
// 	// 		return
// 	// 	}

// 	// }
// 	// err = dbTx.Commit()
// 	// return

// 	// if err != nil {
// 	// 	return
// 	// }
// 	for _, appearance := range appearances {
// 		// m := map[string]any{
// 		// 	"address":     appearance.Address.String(),
// 		// 	"blockNumber": appearance.BlockNumber,
// 		// 	"txIndex":     appearance.TransactionIndex,
// 		// 	"provider":    provider,
// 		// }
// 		// rows, err := d.db.NamedQuery(
// 		// 	`insert into appearances(address, block_number, transaction_index, provider) values(:address, :blockNumber, :txIndex, :provider) returning id`,
// 		// 	m,
// 		// )
// 		// if err != nil {
// 		// 	return err
// 		// }

// 		var appearanceId int
// 		err := d.db.Get(
// 			&appearanceId,
// 			`insert into appearances(address, block_number, transaction_index, provider) values(?, ?, ?, ?) returning id`,
// 			appearance.Address.String(),
// 			appearance.BlockNumber,
// 			appearance.TransactionIndex,
// 			provider,
// 		)
// 		if err != nil {
// 			return err
// 		}

// 		// log.Println("getting id")
// 		// var appearanceId int
// 		// if err := rows.Scan(&appearanceId); err != nil {
// 		// 	return err
// 		// }
// 		_, err = d.db.NamedExec(
// 			`insert into appearance_reasons(appearance_id, provider, reason) values(:id, :provider, :reason)`,
// 			map[string]any{
// 				"id":       appearanceId,
// 				"provider": provider,
// 				"reason":   appearance.Reason,
// 			},
// 		)
// 		if err != nil {
// 			return err
// 		}

// 		if appearance.BalanceChange {
// 			_, err = d.db.NamedExec(
// 				`insert into appearance_balance_changes values(:id, true)`,
// 				map[string]any{
// 					"id": appearanceId,
// 				},
// 			)
// 			if err != nil {
// 				return fmt.Errorf("inserting balance change: %w", err)
// 			}
// 		}
// 	}

// 	return
// }

func (d *Database) SaveAppearance(provider string, appearance AppearanceData) (err error) {

	var appearanceId int
	err = d.db.Get(
		&appearanceId,
		`insert into appearances(address, block_number, transaction_index, provider) values(?, ?, ?, ?) returning id`,
		appearance.Address.String(),
		appearance.BlockNumber,
		appearance.TransactionIndex,
		provider,
	)
	if err != nil {
		return err
	}

	_, err = d.db.NamedExec(
		`insert into appearance_reasons(appearance_id, provider, reason) values(:id, :provider, :reason)`,
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
			`insert into appearance_balance_changes values(:id, true)`,
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
		`insert into incompatible_addresses values(:address, :appearanceCount)`,
		map[string]any{
			"address":         address,
			"appearanceCount": len(appearances),
		},
	)
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

	rawSql += ` and json_array_length(providers) = ?);`

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

func (d *Database) AppearancesHavingProvider(provider string) (appearances []types.Appearance, err error) {
	raws := []dbAppearance{}
	err = d.db.Select(
		&raws,
		`select
		address,
		block_number,
		transaction_index
		from view_appearances_with_providers
		where exists (select 1 from json_each(providers) where value = ?)`,
		provider,
	)
	if err != nil {
		return
	}

	for _, dbApp := range raws {
		appearances = append(appearances, types.Appearance{
			Address:          base.HexToAddress(dbApp.Address),
			BlockNumber:      uint32(dbApp.BlockNumber),
			TransactionIndex: uint32(dbApp.TransactionIndex),
		})
	}
	return
}

func (d *Database) AddressCount() (count int, err error) {
	err = d.db.Get(&count, `SELECT count(*) FROM (SELECT DISTINCT address FROM appearances)`)
	return
}

func (d *Database) AddressCountByProviders(providers []string) (count int, err error) {
	args := make([]any, 0, len(providers)+1)
	rawSql := `select
		count(*)
		from (
			select distinct
			address
			from
			view_appearances_with_providers
			where exists (select 1 from json_each(providers) where value = ?`

	for i := 0; i < len(providers); i++ {
		if i > 0 {
			rawSql += "or ?"
		}
		args = append(args, providers[i])
	}
	args = append(args, len(providers))

	rawSql += ` and json_array_length(providers) = ?));`

	err = d.db.Get(
		&count,
		rawSql,
		args...,
	)

	return
}

func (d *Database) AddressCountHavingProvider(provider string) (count int, err error) {
	err = d.db.Get(
		&count,
		`select
		count(*)
		from (
			select distinct
			address
			from
			view_appearances_with_providers
			where exists (select 1 from json_each(providers) where value = ?)
		)`,
		provider,
	)

	return
}

func (d *Database) AppearancesWithReasonOnlyByProvider(provider string) (apps []types.Appearance, err error) {
	err = d.db.Select(
		&apps,
		`select
		id,
		address,
		block_number,
		transaction_index,
		r.reason
		from view_appearances_with_providers v
		join appearance_reasons r on r.appearance_id = v.id
		where exists (
			select 1 from json_each(v.providers) where value = ? and json_array_length(v.providers) = 1
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
		`select
		reason,
		count(*) as count
		from (
			select
			id,
			address,
			block_number,
			transaction_index,
			r.reason
			from view_appearances_with_providers v
			join appearance_reasons r on r.appearance_id = v.id
			where exists (
				select 1 from json_each(v.providers) where value = ? and json_array_length(v.providers) = 1
			)
		) group by reason order by count`,
		provider,
	)

	return
}

func (d *Database) AppearanceBalanceChangeCountOnlyByProvider(provider string) (count int, err error) {
	err = d.db.Get(
		&count,
		`select
		count(*) as count
		from view_appearances_with_providers v
		join appearance_balance_changes c on c.appearance_id = v.id
		where exists (
			select 1 from json_each(v.providers) where value = ? and json_array_length(v.providers) = 1
		)`,
		provider,
	)

	return
}
