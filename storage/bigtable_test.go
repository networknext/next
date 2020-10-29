package storage_test

import (
    "context"
    "testing"
    "time"
    "os"
    "strconv"

    "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/networknext/backend/storage"

    "cloud.google.com/go/bigtable"
    "google.golang.org/api/option"
)

func checkBigtableEmulation(t *testing.T) {
	bigtableEmulatorHost := os.Getenv("BIGTABLE_EMULATOR_HOST")
	if bigtableEmulatorHost == "" {
		t.Skip("Bigtable emulator not set up, skipping bigtable test")
	}
}

func TestBigTableAdmin(t *testing.T) {
	checkBigtableEmulation(t)

	t.Parallel()

	ctx := context.Background()
	logger := log.NewNopLogger()

	t.Run("New Bigtable Admin", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	})

	t.Run("New Bigtable Admin with Opts", func (t *testing.T) {
		opts := option.WithoutAuthentication()
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger, opts)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	})

	t.Run("Get Table List", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		defer func() {
			err := btAdmin.Close()
			assert.NoError(t, err)
		}()

		_, err = btAdmin.GetTableList(ctx)
		assert.NoError(t, err)
	})

	t.Run("Create and Delete Table", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		tblName := "Test"
		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)
	})

	t.Run("Verify Table Exists True", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		tblName := "Test"
		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)

		retVal, err := btAdmin.VerifyTableExists(ctx, tblName)
		
		assert.NoError(t, err)
		assert.Equal(t, retVal, true)
	})

	t.Run("Verify Table Exists False", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		defer func() {
			err := btAdmin.Close()
			assert.NoError(t, err)
		}()

		tblName := "Test"

		retVal, err := btAdmin.VerifyTableExists(ctx, tblName)
		
		assert.NoError(t, err)
		assert.Equal(t, retVal, false)
	})

	t.Run("Set Max Age Policy", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		tblName := "Test"
		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)

		maxAge := time.Hour * time.Duration(1)
		err = btAdmin.SetMaxAgePolicy(ctx, tblName, btCfNames, maxAge)
		assert.NoError(t, err)
	})
}

func TestBigTable(t *testing.T) {
	checkBigtableEmulation(t)
	
	t.Parallel()

	ctx := context.Background()
	logger := log.NewNopLogger()

	t.Run("New Bigtable No Table", func (t *testing.T) {
		os.Setenv("GOOGLE_BIGTABLE_TABLE_NAME", "")
		defer os.Unsetenv("GOOGLE_BIGTABLE_TABLE_NAME")

		t.Run("New Bigtable No Table With Opts", func (t *testing.T) {
			opts := option.WithoutAuthentication()
			btClient, err := storage.NewBigTable(ctx, "", "", logger, opts)
			assert.EqualError(t, err, "NewBigTable() GOOGLE_BIGTABLE_TABLE_NAME is not defined")
			assert.Nil(t, btClient)
		})

		t.Run("New Bigtable No Table Without Opts", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.EqualError(t, err, "NewBigTable() GOOGLE_BIGTABLE_TABLE_NAME is not defined")
			assert.Nil(t, btClient)
		})
	})

	t.Run("New Bigtable With Table", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		tblName := "Test"
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)

		btAdmin, err = storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		os.Setenv("GOOGLE_BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("GOOGLE_BIGTABLE_TABLE_NAME")

		t.Run("New Bigtable With Table With Opts", func (t *testing.T) {
			opts := option.WithoutAuthentication()
			btClient, err := storage.NewBigTable(ctx, "", "", logger, opts)
			assert.NoError(t, err)

			err = btClient.Close()
			assert.NoError(t, err)
		})

		t.Run("New Bigtable With Table Without Opts", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)

			err = btClient.Close()
			assert.NoError(t, err)
		})
		
	})

	t.Run("Get Table", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		tblName := "Test"
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)

		btAdmin, err = storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		os.Setenv("GOOGLE_BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("GOOGLE_BIGTABLE_TABLE_NAME")

		t.Run("Get Table That Exists", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			tbl := btClient.GetTable("Test")
			assert.NotEmpty(t, tbl)
		})

		t.Run("Get Table That Does Not Exist", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			// Even if a table does not exist, Bigtable API will return a Table struct with the given name
			tbl := btClient.GetTable("")

			assert.NotNil(t, tbl)
		})
	})

	t.Run("Insert Row Into Table", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		tblName := "Test"
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)

		btAdmin, err = storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		os.Setenv("GOOGLE_BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("GOOGLE_BIGTABLE_TABLE_NAME")

		t.Run("Insert Valid Row Into Table", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			metaBinary := make([]byte, 0)
			sliceBinary := make([]byte, 0)
			
			dataMap := make(map[string][]byte)
			dataMap["meta"] = metaBinary
			dataMap["slice"] = sliceBinary

			cfMap := make(map[string]string)
			cfMap["meta"] = btCfNames[0]
			cfMap["slice"] = btCfNames[0]

			err = btClient.InsertRowInTable(ctx, rowKeys, dataMap, cfMap)
			assert.NoError(t, err)
		})

		t.Run("Insert Invalid Row Into Table", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			dataMap := make(map[string][]byte)
			dataMap["meta"] = []byte{}

			cfMap := make(map[string]string)

			// Should not be able to find key "meta" in cfMap
			err = btClient.InsertRowInTable(ctx, rowKeys, dataMap, cfMap)
			assert.EqualError(t, err, "Column name meta not present in column family map")
		})

		t.Run("Insert Invalid Row Into Table 100000 Mutations", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			dataMap := make(map[string][]byte)
			dataMap["meta"] = []byte{}

			cfMap := make(map[string]string)
			cfMap["meta"] = "test"

			for i := 0; i < 100010; i++ {
				iStr := strconv.Itoa(i)
				dataMap[iStr] = []byte{}
				cfMap[iStr] = iStr
			}

			// Should not be able to handle more than 100000 mutations
			err = btClient.InsertRowInTable(ctx, rowKeys, dataMap, cfMap)
			assert.NotNil(t, err)
		})
	})

	t.Run("Insert Session Data Into Table", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		tblName := "Test"
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)

		btAdmin, err = storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		os.Setenv("GOOGLE_BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("GOOGLE_BIGTABLE_TABLE_NAME")

		t.Run("Insert Valid Session Data Into Table", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			metaBinary := make([]byte, 0)
			sliceBinary := make([]byte, 0)


			err = btClient.InsertSessionData(ctx, btCfNames, metaBinary, sliceBinary, rowKeys)
			assert.NoError(t, err)
		})

		t.Run("Insert Invalid Session Data Into Table", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			metaBinary := make([]byte, 0)
			sliceBinary := make([] byte, 0)
			
			// Should attempt to create column family map and fail
			err = btClient.InsertSessionData(ctx, make([]string, 0), metaBinary, sliceBinary, rowKeys)
			assert.EqualError(t, err, "Column family names slice is empty")
		})
	})

	t.Run("Get Rows With Prefix", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		tblName := "Test"
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)

		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		os.Setenv("GOOGLE_BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("GOOGLE_BIGTABLE_TABLE_NAME")

		t.Run("Get Rows With Prefix With Opts Success", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			metaBinary := make([]byte, 0)
			sliceBinary := make([]byte, 0)

			err = btClient.InsertSessionData(ctx, btCfNames, metaBinary, sliceBinary, rowKeys)
			assert.NoError(t, err)

			prefix := "meta"
			opts := bigtable.RowFilter(bigtable.ColumnFilter("meta"))
			_, err = btClient.GetRowsWithPrefix(ctx, prefix, opts)

			assert.NoError(t, err)
		})

		t.Run("Get Rows With Prefix Without Opts Success", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			metaBinary := make([]byte, 0)
			sliceBinary := make([]byte, 0)

			err = btClient.InsertSessionData(ctx, btCfNames, metaBinary, sliceBinary, rowKeys)
			assert.NoError(t, err)

			prefix := "meta"
			_, err = btClient.GetRowsWithPrefix(ctx, prefix)

			assert.NoError(t, err)
		})

	})

	t.Run("Get Row With Rowkey", func (t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)

		tblName := "Test"
		btCfNames := []string{"ColFamName"}
		err = btAdmin.CreateTable(ctx, tblName, btCfNames)
		assert.NoError(t, err)

		defer func() {
			err := btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		}()

		os.Setenv("GOOGLE_BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("GOOGLE_BIGTABLE_TABLE_NAME")

		t.Run("Get Row With Rowkey With Opts Success", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			metaBinary := make([]byte, 0)
			sliceBinary := make([]byte, 0)

			err = btClient.InsertSessionData(ctx, btCfNames, metaBinary, sliceBinary, rowKeys)
			assert.NoError(t, err)

			opts := bigtable.RowFilter(bigtable.ColumnFilter("meta"))
			_, err = btClient.GetRowWithRowKey(ctx, rowKeys[0], opts)

			assert.NoError(t, err)
		})

		t.Run("Get Rows With Rowkey Without Opts Success", func (t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			
			defer func () {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta", "slice"}
			
			metaBinary := make([]byte, 0)
			sliceBinary := make([]byte, 0)

			err = btClient.InsertSessionData(ctx, btCfNames, metaBinary, sliceBinary, rowKeys)
			assert.NoError(t, err)

			_, err = btClient.GetRowWithRowKey(ctx, rowKeys[0])

			assert.NoError(t, err)
		})
	})
}	
