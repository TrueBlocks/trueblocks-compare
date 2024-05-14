package main

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	_ "modernc.org/sqlite"
)

func databaseFileName(t *testing.T) string {
	t.Helper()
	filename := path.Join(
		os.TempDir(),
		"test_db.sqlite",
	)
	t.Log("Using database:", filename)
	return filename
}

func makeTestDatabase(t *testing.T) (*Database, func()) {
	t.Helper()

	fileName := databaseFileName(t)
	d := &Database{
		fileName: fileName,
	}
	if err := d.openDatabase(); err != nil {
		t.Fatal(err)
	}

	if err := d.PrepareTables(); err != nil {
		t.Fatal(err)
	}

	return d, func() {
		d.Close()
		os.Remove(fileName)
	}
}

func TestDatabase_PrepareTables(t *testing.T) {
	var err error
	fileName := databaseFileName(t)
	d := &Database{
		fileName: fileName,
	}
	if err = d.openDatabase(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.Close()
		os.Remove(fileName)
	}()

	if err := d.PrepareTables(); err != nil {
		t.Fatal(err)
	}
}

func TestDatabase_MarkAsDownloaded(t *testing.T) {
	d, cleanup := makeTestDatabase(t)
	defer cleanup()

	if err := d.MarkAsDownloaded("0x0", "key"); err != nil {
		t.Fatal(err)
	}

	var result int
	err := d.db.QueryRow("select count(*) from download_status;").Scan(&result)
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatal("wrong result:", result)
	}
}

// func TestDatabase_Downloaded(t *testing.T) {
// 	d, cleanup := makeTestDatabase(t)
// 	defer cleanup()

// 	if err := d.MarkAsDownloaded("0x0", "key"); err != nil {
// 		t.Fatal(err)
// 	}

// 	result, err := d.Downloaded("0x0", "key")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !result {
// 		t.Fatal("wrong result:", result)
// 	}
// }

func TestDatabase_Downloaded(t *testing.T) {
	d, cleanup := makeTestDatabase(t)
	defer cleanup()

	if err := d.MarkAsDownloaded("0x0", "key"); err != nil {
		t.Fatal(err)
	}

	result, ok, err := d.Downloaded("0x0")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if !reflect.DeepEqual(result, []string{"key"}) {
		t.Fatal("wrong result:", result)
	}
}

func TestDatabase_SaveAppearances(t *testing.T) {
	d, cleanup := makeTestDatabase(t)
	defer cleanup()

	appearances := []types.Appearance{
		{
			Address:          base.HexToAddress("0x0"),
			BlockNumber:      1,
			TransactionIndex: 1,
		},
		{
			Address:          base.HexToAddress("0x1"),
			BlockNumber:      2,
			TransactionIndex: 2,
		},
		{
			Address:          base.HexToAddress("0x2"),
			BlockNumber:      3,
			TransactionIndex: 3,
		},
	}

	if err := d.SaveAppearances("etherscan", appearances); err != nil {
		t.Fatal(err)
	}

	var count int
	err := d.db.QueryRow("select count(*) from appearances where provider = 'etherscan'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatal("wrong count:", count)
	}
}

func TestDatabase_SelectByProviders(t *testing.T) {
	d, cleanup := makeTestDatabase(t)
	defer cleanup()

	appearances := []types.Appearance{
		{
			Address:          base.HexToAddress("0x0"),
			BlockNumber:      1,
			TransactionIndex: 1,
		},
		{
			Address:          base.HexToAddress("0x1"),
			BlockNumber:      2,
			TransactionIndex: 2,
		},
		{
			Address:          base.HexToAddress("0x2"),
			BlockNumber:      3,
			TransactionIndex: 3,
		},
	}

	if err := d.SaveAppearances("etherscan", []types.Appearance{appearances[0]}); err != nil {
		t.Fatal(err)
	}
	if err := d.SaveAppearances("etherscan", []types.Appearance{appearances[1]}); err != nil {
		t.Fatal(err)
	}
	if err := d.SaveAppearances("alchemy", []types.Appearance{appearances[1]}); err != nil {
		t.Fatal(err)
	}
	if err := d.SaveAppearances("key", []types.Appearance{appearances[1]}); err != nil {
		t.Fatal(err)
	}

	esOnly, err := d.AppearancesByProviders([]string{"etherscan"})
	if err != nil {
		t.Fatal(err)
	}
	if l := len(esOnly); l != 1 {
		t.Fatal("wrong length", l)
	}

	alchemyOnly, err := d.AppearancesByProviders([]string{"alchemy"})
	if err != nil {
		t.Fatal(err)
	}
	if l := len(alchemyOnly); l != 0 {
		t.Fatal("wrong length", l)
	}

	etherscanAlchemy, err := d.AppearancesByProviders([]string{"etherscan", "alchemy"})
	if err != nil {
		t.Fatal(err)
	}
	if l := len(etherscanAlchemy); l != 0 {
		t.Fatal("wrong length", l)
	}

	keyOnly, err := d.AppearancesByProviders([]string{"key"})
	if err != nil {
		t.Fatal(err)
	}
	if l := len(keyOnly); l != 0 {
		t.Fatal("wrong length", l)
	}
}
