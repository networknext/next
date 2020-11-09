package storage_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"

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

	ctx := context.Background()
	logger := log.NewNopLogger()

	t.Run("New Bigtable Admin", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

		err = btAdmin.Close()
		assert.NoError(t, err)
	})

	t.Run("New Bigtable Admin with Opts", func(t *testing.T) {
		opts := option.WithoutAuthentication()
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger, opts)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

		err = btAdmin.Close()
		assert.NoError(t, err)
	})

	t.Run("Get Table List", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

		defer func() {
			err := btAdmin.Close()
			assert.NoError(t, err)
		}()

		_, err = btAdmin.GetTableList(ctx)
		assert.NoError(t, err)
	})

	t.Run("Create and Delete Table", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)
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

	t.Run("Create Table That Already Exists", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)
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

		err = btAdmin.CreateTable(ctx, tblName, btCfNames)

		errorStr := fmt.Sprintf("rpc error: code = AlreadyExists desc = table \"projects//instances//tables/%s\" already exists", tblName)
		assert.EqualError(t, err, errorStr)
	})

	t.Run("Verify Table Exists True", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)
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

	t.Run("Verify Table Exists False", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

		defer func() {
			err := btAdmin.Close()
			assert.NoError(t, err)
		}()

		tblName := "Test"

		retVal, err := btAdmin.VerifyTableExists(ctx, tblName)

		assert.NoError(t, err)
		assert.Equal(t, retVal, false)
	})

	t.Run("Set Max Age Policy", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)
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

	t.Run("New Bigtable No Table", func(t *testing.T) {
		os.Setenv("BIGTABLE_TABLE_NAME", "")
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")

		t.Run("New Bigtable No Table With Opts", func(t *testing.T) {
			opts := option.WithoutAuthentication()
			btClient, err := storage.NewBigTable(ctx, "", "", logger, opts)
			assert.EqualError(t, err, "NewBigTable() BIGTABLE_TABLE_NAME is not defined")
			assert.Nil(t, btClient)
		})

		t.Run("New Bigtable No Table Without Opts", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.EqualError(t, err, "NewBigTable() BIGTABLE_TABLE_NAME is not defined")
			assert.Nil(t, btClient)
		})
	})

	t.Run("New Bigtable With Table", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

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

		os.Setenv("BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")

		t.Run("New Bigtable With Table With Opts", func(t *testing.T) {
			opts := option.WithoutAuthentication()
			btClient, err := storage.NewBigTable(ctx, "", "", logger, opts)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			err = btClient.Close()
			assert.NoError(t, err)
		})

		t.Run("New Bigtable With Table Without Opts", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			err = btClient.Close()
			assert.NoError(t, err)
		})
	})

	t.Run("Get Table", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

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

		os.Setenv("BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")

		t.Run("Get Table That Exists", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			tbl := btClient.GetTable("Test")
			assert.NotEmpty(t, tbl)
		})

		t.Run("Get Table That Does Not Exist", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			// Even if a table does not exist, Bigtable API will return a Table struct with the given name
			tbl := btClient.GetTable("")

			assert.NotNil(t, tbl)
		})
	})

	t.Run("Insert Row Into Table", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

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

		os.Setenv("BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")

		t.Run("Insert Valid Meta Row Into Table", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta"}

			metaBinary, err := transport.SessionMeta{}.MarshalBinary()
			assert.NoError(t, err)

			dataMap := make(map[string][]byte)
			dataMap["meta"] = metaBinary

			cfMap := make(map[string]string)
			cfMap["meta"] = btCfNames[0]

			err = btClient.InsertRowInTable(ctx, rowKeys, dataMap, cfMap)
			assert.NoError(t, err)
		})

		t.Run("Insert Valid Slice Row Into Table", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"slices"}

			sliceBinary, err := transport.SessionSlice{}.MarshalBinary()
			assert.NoError(t, err)

			dataMap := make(map[string][]byte)
			dataMap["slices"] = sliceBinary

			cfMap := make(map[string]string)
			cfMap["slices"] = btCfNames[0]

			err = btClient.InsertRowInTable(ctx, rowKeys, dataMap, cfMap)
			assert.NoError(t, err)
		})

		t.Run("Insert Invalid Row Into Table", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta"}

			dataMap := make(map[string][]byte)
			dataMap["meta"] = []byte{}

			cfMap := make(map[string]string)

			// Should not be able to find key "meta" in cfMap
			err = btClient.InsertRowInTable(ctx, rowKeys, dataMap, cfMap)
			assert.EqualError(t, err, "InsertRowInTable() Column name meta not present in column family map")
		})

		t.Run("Insert Invalid Row Into Table 100000 Mutations", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta"}

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

	t.Run("Insert Session Data Into Table", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

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

		os.Setenv("BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")

		t.Run("Insert Valid Session Meta Data Into Table", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta"}

			metaBinary, err := transport.SessionMeta{}.MarshalBinary()
			assert.NoError(t, err)

			err = btClient.InsertSessionMetaData(ctx, btCfNames, metaBinary, rowKeys)
			assert.NoError(t, err)
		})

		t.Run("Insert Valid Session Slice Data Into Table", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"slices"}

			sliceBinary, err := transport.SessionSlice{}.MarshalBinary()
			assert.NoError(t, err)

			err = btClient.InsertSessionSliceData(ctx, btCfNames, sliceBinary, rowKeys)
			assert.NoError(t, err)
		})

		t.Run("Insert Invalid Session Meta Data Into Table", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"meta"}

			metaBinary := make([]byte, 0)

			// Should attempt to create column family map and fail
			err = btClient.InsertSessionMetaData(ctx, make([]string, 0), metaBinary, rowKeys)
			assert.EqualError(t, err, "InsertSessionMetaData() Column family names slice is empty")
		})

		t.Run("Insert Invalid Session Slice Data Into Table", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"slices"}

			sliceBinary := make([]byte, 0)

			// Should attempt to create column family map and fail
			err = btClient.InsertSessionSliceData(ctx, make([]string, 0), sliceBinary, rowKeys)
			assert.EqualError(t, err, "InsertSessionSliceData() Column family names slice is empty")
		})
	})

	t.Run("Insert Session Meta Data Into Nonexistent Table", func(t *testing.T) {
		tblName, exists := os.LookupEnv("BIGTABLE_TABLE_NAME")
		if exists {
			// Delete the table that exists
			btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btAdmin)

			err = btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		} else {
			os.Setenv("BIGTABLE_TABLE_NAME", "Test")
			defer os.Unsetenv("BIGTABLE_TABLE_NAME")
		}

		btClient, err := storage.NewBigTable(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btClient)

		defer func() {
			err = btClient.Close()
			assert.NoError(t, err)
		}()

		btCfNames := []string{"ColFamName"}
		rowKeys := []string{"meta"}

		metaBinary := make([]byte, 0)

		err = btClient.InsertSessionMetaData(ctx, btCfNames, metaBinary, rowKeys)
		assert.NotNil(t, err)
	})

	t.Run("Insert Session Slice Data Into Nonexistent Table", func(t *testing.T) {
		tblName, exists := os.LookupEnv("BIGTABLE_TABLE_NAME")
		if exists {
			// Delete the table that exists
			btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btAdmin)

			err = btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		} else {
			os.Setenv("BIGTABLE_TABLE_NAME", "Test")
			defer os.Unsetenv("BIGTABLE_TABLE_NAME")
		}

		btClient, err := storage.NewBigTable(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btClient)

		defer func() {
			err = btClient.Close()
			assert.NoError(t, err)
		}()

		btCfNames := []string{"ColFamName"}
		rowKeys := []string{"slices"}

		sliceBinary := make([]byte, 0)

		err = btClient.InsertSessionSliceData(ctx, btCfNames, sliceBinary, rowKeys)
		assert.NotNil(t, err)
	})

	t.Run("Get Rows With Prefix", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

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

		os.Setenv("BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")

		t.Run("Get Meta Rows With Prefix With Opts Success", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"metaRow"}

			metaBinary := make([]byte, 0)

			err = btClient.InsertSessionMetaData(ctx, btCfNames, metaBinary, rowKeys)
			assert.NoError(t, err)

			prefix := "meta"
			opts := bigtable.RowFilter(bigtable.ColumnFilter("meta"))
			rows, err := btClient.GetRowsWithPrefix(ctx, prefix, opts)

			assert.NoError(t, err)
			assert.NotNil(t, rows)
		})

		t.Run("Get Slice Rows With Prefix With Opts Success", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"sliceRow"}

			sliceBinary := make([]byte, 0)

			err = btClient.InsertSessionSliceData(ctx, btCfNames, sliceBinary, rowKeys)
			assert.NoError(t, err)

			prefix := "slice"
			opts := bigtable.RowFilter(bigtable.ColumnFilter("slices"))
			rows, err := btClient.GetRowsWithPrefix(ctx, prefix, opts)

			assert.NoError(t, err)
			assert.NotNil(t, rows)
		})

		t.Run("Get Meta Rows With Prefix Without Opts Success", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"metaRow"}

			metaBinary := make([]byte, 0)

			err = btClient.InsertSessionMetaData(ctx, btCfNames, metaBinary, rowKeys)
			assert.NoError(t, err)

			prefix := "meta"
			rows, err := btClient.GetRowsWithPrefix(ctx, prefix)

			assert.NoError(t, err)
			assert.NotNil(t, rows)
		})

		t.Run("Get Slice Rows With Prefix Without Opts Success", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"sliceRow"}

			sliceBinary := make([]byte, 0)

			err = btClient.InsertSessionSliceData(ctx, btCfNames, sliceBinary, rowKeys)
			assert.NoError(t, err)

			prefix := "slice"
			rows, err := btClient.GetRowsWithPrefix(ctx, prefix)

			assert.NoError(t, err)
			assert.NotNil(t, rows)
		})
	})

	t.Run("Get Rows With Prefix From Nonexistent Table", func(t *testing.T) {
		tblName, exists := os.LookupEnv("BIGTABLE_TABLE_NAME")
		if tblName == "" {
			tblName = "Test"
		}
		if exists {
			// Delete the table that exists
			btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btAdmin)

			err = btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		} else {
			os.Setenv("BIGTABLE_TABLE_NAME", tblName)
			defer os.Unsetenv("BIGTABLE_TABLE_NAME")
		}

		btClient, err := storage.NewBigTable(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btClient)

		defer func() {
			err = btClient.Close()
			assert.NoError(t, err)
		}()

		prefix := "meta"
		_, err = btClient.GetRowsWithPrefix(ctx, prefix)

		errorStr := fmt.Sprintf("GetRowsWithPrefix() Could not get rows with prefix %s: rpc error: code = NotFound desc = table \"projects//instances//tables/%s\" not found", prefix, tblName)
		assert.EqualError(t, err, errorStr)
	})

	t.Run("Get Row With Rowkey", func(t *testing.T) {
		btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btAdmin)

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

		os.Setenv("BIGTABLE_TABLE_NAME", "Test")
		defer os.Unsetenv("BIGTABLE_TABLE_NAME")

		t.Run("Get Meta Row With Rowkey With Opts Success", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"metaRow"}

			metaBinary := make([]byte, 0)

			err = btClient.InsertSessionMetaData(ctx, btCfNames, metaBinary, rowKeys)
			assert.NoError(t, err)

			opts := bigtable.RowFilter(bigtable.ColumnFilter("meta"))
			row, err := btClient.GetRowWithRowKey(ctx, rowKeys[0], opts)

			assert.NoError(t, err)
			assert.NotEmpty(t, row)
		})

		t.Run("Get Slice Row With Rowkey With Opts Success", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"slices"}

			sliceBinary, err := transport.SessionSlice{}.MarshalBinary()
			assert.NoError(t, err)

			err = btClient.InsertSessionSliceData(ctx, btCfNames, sliceBinary, rowKeys)
			assert.NoError(t, err)

			opts := bigtable.RowFilter(bigtable.ColumnFilter("slices"))
			row, err := btClient.GetRowWithRowKey(ctx, rowKeys[0], opts)

			assert.NoError(t, err)
			assert.NotEmpty(t, row)
		})

		t.Run("Get Meta Row With Rowkey Without Opts Success", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"metaRow"}

			metaBinary := make([]byte, 0)

			err = btClient.InsertSessionMetaData(ctx, btCfNames, metaBinary, rowKeys)
			assert.NoError(t, err)

			row, err := btClient.GetRowWithRowKey(ctx, rowKeys[0])

			assert.NoError(t, err)
			assert.NotEmpty(t, row)
		})

		t.Run("Get Slice Row With Rowkey Without Opts Success", func(t *testing.T) {
			btClient, err := storage.NewBigTable(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btClient)

			defer func() {
				err = btClient.Close()
				assert.NoError(t, err)
			}()

			rowKeys := []string{"sliceRow"}

			sliceBinary := make([]byte, 0)

			err = btClient.InsertSessionSliceData(ctx, btCfNames, sliceBinary, rowKeys)
			assert.NoError(t, err)

			row, err := btClient.GetRowWithRowKey(ctx, rowKeys[0])

			assert.NoError(t, err)
			assert.NotEmpty(t, row)
		})
	})

	t.Run("Get Rows With Rowkey From Nonexistent Table", func(t *testing.T) {
		tblName, exists := os.LookupEnv("BIGTABLE_TABLE_NAME")
		if tblName == "" {
			tblName = "Test"
		}

		if exists {
			// Delete the table that exists
			btAdmin, err := storage.NewBigTableAdmin(ctx, "", "", logger)
			assert.NoError(t, err)
			assert.NotNil(t, btAdmin)

			err = btAdmin.DeleteTable(ctx, tblName)
			assert.NoError(t, err)

			err = btAdmin.Close()
			assert.NoError(t, err)
		} else {
			os.Setenv("BIGTABLE_TABLE_NAME", tblName)
			defer os.Unsetenv("BIGTABLE_TABLE_NAME")
		}

		btClient, err := storage.NewBigTable(ctx, "", "", logger)
		assert.NoError(t, err)
		assert.NotNil(t, btClient)

		defer func() {
			err = btClient.Close()
			assert.NoError(t, err)
		}()

		rowKeys := []string{"meta"}

		_, err = btClient.GetRowWithRowKey(ctx, rowKeys[0])

		errorStr := fmt.Sprintf("rpc error: code = NotFound desc = table \"projects//instances//tables/%s\" not found", tblName)
		assert.EqualError(t, err, errorStr)
	})
}
