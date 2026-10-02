package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/domain"
	"github.com/Vlad777-bit/personal-finance-analytics/ledger/internal/repository"
)

func (s *service) ExportTransactions(
	ctx context.Context,
	input ExportTransactionsInput,
) (string, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return "", domain.ErrUserIDRequired
	}
	if input.From.IsZero() || input.To.IsZero() || !input.From.Before(input.To) {
		return "", domain.ErrInvalidPeriod
	}

	transactions, err := s.transactionRepository.List(ctx, repository.TransactionFilter{
		UserID:   userID,
		Category: strings.TrimSpace(input.Category),
		From:     input.From,
		To:       input.To,
	})
	if err != nil {
		return "", fmt.Errorf("list transactions for export: %w", err)
	}

	var output strings.Builder
	writer := csv.NewWriter(&output)
	if err := writer.Write(expectedTransactionCSVHeader); err != nil {
		return "", fmt.Errorf("write CSV header: %w", err)
	}
	for _, transaction := range transactions {
		if err := writer.Write([]string{
			strconv.FormatInt(transaction.Amount, 10),
			transaction.Category,
			transaction.Description,
			transaction.OccurredAt.Format(timeRFC3339Nano),
		}); err != nil {
			return "", fmt.Errorf("write transaction CSV row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush transaction CSV: %w", err)
	}

	return output.String(), nil
}

const timeRFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
