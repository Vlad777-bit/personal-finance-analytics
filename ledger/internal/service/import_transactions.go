package service

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
)

const transactionCSVColumns = 4

var expectedTransactionCSVHeader = []string{
	"amount",
	"category",
	"description",
	"occurred_at",
}

func (s *service) ImportTransactions(
	ctx context.Context,
	input ImportTransactionsInput,
) (int, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return 0, domain.ErrUserIDRequired
	}
	if strings.TrimSpace(input.CSVData) == "" {
		return 0, fmt.Errorf("%w: data is required", domain.ErrInvalidCSV)
	}

	reader := csv.NewReader(strings.NewReader(input.CSVData))
	reader.FieldsPerRecord = transactionCSVColumns
	reader.ReuseRecord = false

	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("%w: read header: %w", domain.ErrInvalidCSV, err)
	}
	if !equalStrings(header, expectedTransactionCSVHeader) {
		return 0, fmt.Errorf("%w: invalid header", domain.ErrInvalidCSV)
	}

	imported := 0
	for {
		if err := ctx.Err(); err != nil {
			return imported, fmt.Errorf("import transactions: %w", err)
		}

		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			return imported, nil
		}
		if readErr != nil {
			return imported, fmt.Errorf(
				"%w: read row %d: %w",
				domain.ErrInvalidCSV,
				imported+2,
				readErr,
			)
		}

		amount, parseErr := strconv.ParseInt(strings.TrimSpace(record[0]), 10, 64)
		if parseErr != nil {
			return imported, fmt.Errorf(
				"%w: parse row %d amount: %w",
				domain.ErrInvalidCSV,
				imported+2,
				parseErr,
			)
		}
		occurredAt, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(record[3]))
		if parseErr != nil {
			return imported, fmt.Errorf(
				"%w: parse row %d occurred_at: %w",
				domain.ErrInvalidCSV,
				imported+2,
				parseErr,
			)
		}

		transaction, parseErr := domain.NewTransaction(domain.NewTransactionParams{
			UserID:      userID,
			Amount:      amount,
			Category:    record[1],
			Description: record[2],
			OccurredAt:  occurredAt,
		})
		if parseErr != nil {
			return imported, fmt.Errorf(
				"%w: validate row %d: %w",
				domain.ErrInvalidCSV,
				imported+2,
				parseErr,
			)
		}

		if _, createErr := s.transactionRepository.CreateWithinBudget(ctx, transaction); createErr != nil {
			return imported, fmt.Errorf("create CSV row %d: %w", imported+2, createErr)
		}
		if invalidateErr := s.summaryCache.InvalidateUser(ctx, userID); invalidateErr != nil {
			s.logger.Warn(
				"invalidate summary cache after CSV transaction",
				"user_id",
				userID,
				"error",
				invalidateErr,
			)
		}
		imported++
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if strings.TrimSpace(strings.ToLower(left[index])) != right[index] {
			return false
		}
	}

	return true
}
