package handler_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hibiken/asynq"

	"github.com/example/brainway/internal/handler"
	"github.com/example/brainway/internal/queue"
	"github.com/example/brainway/pb"
)

func TestIngestBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		transactions    []*pb.Transaction
		enqueueErr      error
		wantQueued      int32
		wantFailedIDs   []string
		wantFailReasons []string // partial substrings matched against each failure reason
	}{
		{
			name: "zero_amount_rejected",
			transactions: []*pb.Transaction{
				{Id: "tx1", UserId: "u1", Amount: 0, Currency: "USD"},
			},
			wantQueued:      0,
			wantFailedIDs:   []string{"tx1"},
			wantFailReasons: []string{"greater than zero"},
		},
		{
			name: "negative_amount_rejected",
			transactions: []*pb.Transaction{
				{Id: "tx2", UserId: "u1", Amount: -10, Currency: "USD"},
			},
			wantQueued:      0,
			wantFailedIDs:   []string{"tx2"},
			wantFailReasons: []string{"greater than zero"},
		},
		{
			name: "unsupported_currency_rejected",
			transactions: []*pb.Transaction{
				{Id: "tx3", UserId: "u1", Amount: 100, Currency: "GBP"},
			},
			wantQueued:      0,
			wantFailedIDs:   []string{"tx3"},
			wantFailReasons: []string{"not supported"},
		},
		{
			name: "valid_usd_transaction_queued",
			transactions: []*pb.Transaction{
				{Id: "tx4", UserId: "u2", Amount: 50.5, Currency: "USD"},
			},
			wantQueued:    1,
			wantFailedIDs: nil,
		},
		{
			name: "all_supported_currencies_queued",
			transactions: []*pb.Transaction{
				{Id: "tx5", UserId: "u3", Amount: 1, Currency: "EUR"},
				{Id: "tx6", UserId: "u3", Amount: 1, Currency: "ILS"},
				{Id: "tx7", UserId: "u3", Amount: 1, Currency: "USD"},
			},
			wantQueued:    3,
			wantFailedIDs: nil,
		},
		{
			name: "mixed_batch_partial_success",
			transactions: []*pb.Transaction{
				{Id: "good1", UserId: "u4", Amount: 99, Currency: "USD"},
				{Id: "bad1", UserId: "u4", Amount: 0, Currency: "USD"},
				{Id: "good2", UserId: "u4", Amount: 10, Currency: "ILS"},
				{Id: "bad2", UserId: "u4", Amount: 5, Currency: "JPY"},
			},
			wantQueued:      2,
			wantFailedIDs:   []string{"bad1", "bad2"},
			wantFailReasons: []string{"greater than zero", "not supported"},
		},
		{
			name: "enqueue_error_moves_to_failed",
			transactions: []*pb.Transaction{
				{Id: "tx8", UserId: "u5", Amount: 20, Currency: "EUR"},
			},
			enqueueErr:      errors.New("redis connection refused"),
			wantQueued:      0,
			wantFailedIDs:   []string{"tx8"},
			wantFailReasons: []string{"enqueue error"},
		},
		{
			name:          "empty_batch",
			transactions:  nil,
			wantQueued:    0,
			wantFailedIDs: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &queue.MockEnqueuer{}
			if tc.enqueueErr != nil {
				mock.EnqueueFunc = func(_ *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
					return nil, tc.enqueueErr
				}
			}

			h := handler.New(mock)
			resp, err := h.IngestBatch(context.Background(), &pb.BatchRequest{
				Transactions: tc.transactions,
			})
			if err != nil {
				t.Fatalf("unexpected gRPC error: %v", err)
			}

			if resp.GetQueuedCount() != tc.wantQueued {
				t.Errorf("QueuedCount: got %d, want %d", resp.GetQueuedCount(), tc.wantQueued)
			}

			gotFailed := resp.GetFailed()
			if len(gotFailed) != len(tc.wantFailedIDs) {
				t.Fatalf("failed count: got %d, want %d (failures: %v)",
					len(gotFailed), len(tc.wantFailedIDs), gotFailed)
			}
			for i, f := range gotFailed {
				if f.GetId() != tc.wantFailedIDs[i] {
					t.Errorf("failed[%d].Id: got %q, want %q", i, f.GetId(), tc.wantFailedIDs[i])
				}
				if i < len(tc.wantFailReasons) {
					if !strings.Contains(f.GetReason(), tc.wantFailReasons[i]) {
						t.Errorf("failed[%d].Reason: %q does not contain %q",
							i, f.GetReason(), tc.wantFailReasons[i])
					}
				}
			}

			// Verify the mock was called the right number of times.
			if tc.enqueueErr == nil && int(tc.wantQueued) != len(mock.Calls) {
				t.Errorf("Enqueue call count: got %d, want %d", len(mock.Calls), tc.wantQueued)
			}
		})
	}
}
