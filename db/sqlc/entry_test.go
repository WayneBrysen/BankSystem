package sqlc

import (
	"context"
	"testing"
	"time"

	"github.com/Brysen/simplebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomEntry(t *testing.T, accountID int64) Entry {
	arg := CreateEntryParams{
		AccountID: accountID,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account.ID)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry1 := createRandomEntry(t, account.ID)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt.Time, entry2.CreatedAt.Time, time.Second)
}

func TestUpdateEntry(t *testing.T) {
    // 首先创建一个随机账户和一个随机条目
    account := createRandomAccount(t)
    entry1 := createRandomEntry(t, account.ID)

    // 定义更新参数
    arg := UpdateEntryParams{
        ID:     entry1.ID,
        Amount: util.RandomMoney(), // 随机生成新的金额
    }

    // 执行更新操作
    entry2, err := testQueries.UpdateEntry(context.Background(), arg)
    require.NoError(t, err)
    require.NotEmpty(t, entry2)

    // 验证返回的数据是否按预期更新
    require.Equal(t, entry1.ID, entry2.ID)
    require.Equal(t, entry1.AccountID, entry2.AccountID)
    require.Equal(t, arg.Amount, entry2.Amount) // 验证金额是否更新
    require.WithinDuration(t, entry1.CreatedAt.Time, entry2.CreatedAt.Time, time.Second)
}

func TestDeleteEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry := createRandomEntry(t, account.ID)

	err := testQueries.DeleteEntry(context.Background(), entry.ID)
	require.NoError(t, err)

	_, err = testQueries.GetEntry(context.Background(), entry.ID)
	require.Error(t, err)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomEntry(t, account.ID)
	}

	arg := ListEntriesParams{
		Limit:  5,
		Offset: 5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}

