package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/example/brainway/internal/queue"
	"github.com/example/brainway/pb"
)

var validCurrencies = map[string]struct{}{
	"USD": {},
	"EUR": {},
	"ILS": {},
}

// TransactionPayload is the JSON body stored inside each asynq task.
type TransactionPayload struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Handler implements pb.TransactionServiceServer.
type Handler struct {
	pb.UnimplementedTransactionServiceServer
	enqueuer queue.Enqueuer
}

// New constructs a Handler. Accepts any Enqueuer to allow test injection.
func New(e queue.Enqueuer) *Handler {
	return &Handler{enqueuer: e}
}

// IngestBatch validates each transaction, enqueues valid ones, and collects failures.
func (h *Handler) IngestBatch(
	ctx context.Context,
	req *pb.BatchRequest,
) (*pb.BatchResponse, error) {
	var queued int32
	var failed []*pb.FailedTransaction

	for _, tx := range req.GetTransactions() {
		if reason := validate(tx); reason != "" {
			failed = append(failed, &pb.FailedTransaction{Id: tx.GetId(), Reason: reason})
			continue
		}

		payload, err := json.Marshal(TransactionPayload{
			ID:       tx.GetId(),
			UserID:   tx.GetUserId(),
			Amount:   tx.GetAmount(),
			Currency: tx.GetCurrency(),
		})
		if err != nil {
			failed = append(failed, &pb.FailedTransaction{
				Id:     tx.GetId(),
				Reason: fmt.Sprintf("marshal error: %v", err),
			})
			continue
		}

		task := asynq.NewTask(queue.TaskTypeTransaction, payload)
		if _, err := h.enqueuer.Enqueue(task); err != nil {
			failed = append(failed, &pb.FailedTransaction{
				Id:     tx.GetId(),
				Reason: fmt.Sprintf("enqueue error: %v", err),
			})
			continue
		}

		queued++
	}

	return &pb.BatchResponse{QueuedCount: queued, Failed: failed}, nil
}

// validate returns an empty string when the transaction is valid, or the rejection reason.
func validate(tx *pb.Transaction) string {
	if tx.GetAmount() <= 0 {
		return "amount must be greater than zero"
	}
	if _, ok := validCurrencies[tx.GetCurrency()]; !ok {
		return fmt.Sprintf("currency %q is not supported; must be one of USD, EUR, ILS", tx.GetCurrency())
	}
	return ""
}
