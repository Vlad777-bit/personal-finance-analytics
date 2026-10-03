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
) (ImportTransactionsResult, error) {
	result := ImportTransactionsResult{}
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return result, domain.ErrUserIDRequired
	}
	if strings.TrimSpace(input.CSVData) == "" {
		return result, fmt.Errorf("%w: data is required", domain.ErrInvalidCSV)
	}

	reader := csv.NewReader(strings.NewReader(input.CSVData))
	reader.FieldsPerRecord = transactionCSVColumns
	reader.ReuseRecord = false

	header, err := reader.Read()
	if err != nil {
		return result, fmt.Errorf("%w: read header: %w", domain.ErrInvalidCSV, err)
	}
	if !equalStrings(header, expectedTransactionCSVHeader) {
		return result, fmt.Errorf("%w: invalid header", domain.ErrInvalidCSV)
	}

	rowNumber := 1
	for {
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("import transactions: %w", err)
		}

		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			return result, nil
		}
		rowNumber++
		if readErr != nil {
			result.addError(rowNumber, fmt.Sprintf("read row: %v", readErr))

			continue
		}

		amount, parseErr := strconv.ParseInt(strings.TrimSpace(record[0]), 10, 64)
		if parseErr != nil {
			result.addError(rowNumber, fmt.Sprintf("parse amount: %v", parseErr))

			continue
		}
		occurredAt, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(record[3]))
		if parseErr != nil {
			result.addError(rowNumber, fmt.Sprintf("parse occurred_at: %v", parseErr))

			continue
		}

		transaction, parseErr := domain.NewTransaction(domain.NewTransactionParams{
			UserID:      userID,
			Amount:      amount,
			Category:    record[1],
			Description: record[2],
			OccurredAt:  occurredAt,
		})
		if parseErr != nil {
			result.addError(rowNumber, fmt.Sprintf("validate row: %v", parseErr))

			continue
		}

		if _, createErr := s.transactionRepository.CreateWithinBudget(ctx, transaction); createErr != nil {
			if contextErr := ctx.Err(); contextErr != nil {
				return result, fmt.Errorf("import transactions: %w", contextErr)
			}
			result.addError(rowNumber, createErr.Error())

			continue
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
		result.ImportedCount++
	}
}

func (result *ImportTransactionsResult) addError(row int, message string) {
	result.FailedCount++
	result.Errors = append(result.Errors, ImportTransactionsError{Row: row, Message: message})
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
