package db

import (
	"os"
	"testing"

	"github.com/yz4230/arpmon/pkg/testutil"
)

var tdbf *testutil.TestDBFactory

func TestMain(m *testing.M) {
	tdbf = testutil.NewTestDB()

	baseConfig, err := LoadConfigFromEnv()
	if err != nil {
		panic(err)
	}

	factoryConfig := baseConfig.Clone()
	factoryConfig.Database = ""
	if err = tdbf.Connect(factoryConfig); err != nil {
		panic(err)
	}

	code := m.Run()

	if err := tdbf.Close(); err != nil {
		panic(err)
	}

	os.Exit(code)
}
