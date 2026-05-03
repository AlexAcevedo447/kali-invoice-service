package metrics

import (
	"sync"
)

// Metrics holds counters for invoice operations.
type Metrics struct {
	mu                     sync.RWMutex
	invoicesCreated        int64
	invoicesPaid           int64
	invoicesCanceled       int64
	invoiceItemsCreated    int64
	invoiceItemsUpdated    int64
	invoiceItemsDeleted    int64
	rabbitEventsPublished  int64
	rabbitEventsFailed     int64
	idempotencyHits        int64
	idempotencyMisses      int64
}

var globalMetrics = &Metrics{}

// IncrementInvoicesCreated increments the counter for created invoices.
func IncrementInvoicesCreated() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.invoicesCreated++
}

// IncrementInvoicesPaid increments the counter for paid invoices.
func IncrementInvoicesPaid() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.invoicesPaid++
}

// IncrementInvoicesCanceled increments the counter for canceled invoices.
func IncrementInvoicesCanceled() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.invoicesCanceled++
}

// IncrementInvoiceItemsCreated increments the counter for created invoice items.
func IncrementInvoiceItemsCreated() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.invoiceItemsCreated++
}

// IncrementInvoiceItemsUpdated increments the counter for updated invoice items.
func IncrementInvoiceItemsUpdated() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.invoiceItemsUpdated++
}

// IncrementInvoiceItemsDeleted increments the counter for deleted invoice items.
func IncrementInvoiceItemsDeleted() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.invoiceItemsDeleted++
}

// IncrementRabbitEventsPublished increments the counter for published events.
func IncrementRabbitEventsPublished() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.rabbitEventsPublished++
}

// IncrementRabbitEventsFailed increments the counter for failed events.
func IncrementRabbitEventsFailed() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.rabbitEventsFailed++
}

// IncrementIdempotencyHits increments the counter for idempotency cache hits.
func IncrementIdempotencyHits() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.idempotencyHits++
}

// IncrementIdempotencyMisses increments the counter for idempotency cache misses.
func IncrementIdempotencyMisses() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.idempotencyMisses++
}

// GetSnapshot returns a snapshot of current metrics.
func GetSnapshot() map[string]int64 {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	return map[string]int64{
		"invoices_created":         globalMetrics.invoicesCreated,
		"invoices_paid":            globalMetrics.invoicesPaid,
		"invoices_canceled":        globalMetrics.invoicesCanceled,
		"invoice_items_created":    globalMetrics.invoiceItemsCreated,
		"invoice_items_updated":    globalMetrics.invoiceItemsUpdated,
		"invoice_items_deleted":    globalMetrics.invoiceItemsDeleted,
		"rabbit_events_published":  globalMetrics.rabbitEventsPublished,
		"rabbit_events_failed":     globalMetrics.rabbitEventsFailed,
		"idempotency_hits":         globalMetrics.idempotencyHits,
		"idempotency_misses":       globalMetrics.idempotencyMisses,
	}
}
