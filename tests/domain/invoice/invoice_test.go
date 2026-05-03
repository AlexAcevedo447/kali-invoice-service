package invoice_test

import (
	"testing"
	"time"

	"github.com/AlexAcevedo447/kali-invoice-service/internal/domain/invoice"
)

var baseItems = []invoice.ItemInput{
	{ItemID: "item-1", Quantity: 2, UnitPrice: 50.0},
	{ItemID: "item-2", Quantity: 1, UnitPrice: 100.0},
}

func newTestInvoice(t *testing.T) *invoice.Invoice {
	t.Helper()
	inv, err := invoice.New("customer-1", time.Now(), time.Now().AddDate(0, 0, 30), baseItems)
	if err != nil {
		t.Fatalf("unexpected error creating invoice: %v", err)
	}
	return inv
}

// --- New ---

func TestNew_HappyPath(t *testing.T) {
	inv := newTestInvoice(t)

	if inv.ID == "" {
		t.Error("expected non-empty ID")
	}
	if inv.CustomerID != "customer-1" {
		t.Errorf("expected customer-1, got %s", inv.CustomerID)
	}
	if inv.Status != invoice.StatusPending {
		t.Errorf("expected PENDING, got %s", inv.Status)
	}
	if len(inv.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(inv.Items))
	}
	// 2*50 + 1*100 = 200
	if inv.Total != 200.0 {
		t.Errorf("expected total 200, got %f", inv.Total)
	}
}

func TestNew_MissingCustomerID(t *testing.T) {
	_, err := invoice.New("", time.Now(), time.Now().AddDate(0, 0, 30), baseItems)
	if err == nil {
		t.Error("expected error for missing customer_id")
	}
}

func TestNew_NoItems(t *testing.T) {
	_, err := invoice.New("customer-1", time.Now(), time.Now().AddDate(0, 0, 30), nil)
	if err == nil {
		t.Error("expected error for empty items")
	}
}

func TestNew_InvalidItemQuantity(t *testing.T) {
	items := []invoice.ItemInput{{ItemID: "item-1", Quantity: 0, UnitPrice: 10.0}}
	_, err := invoice.New("customer-1", time.Now(), time.Now().AddDate(0, 0, 30), items)
	if err == nil {
		t.Error("expected error for quantity <= 0")
	}
}

func TestNew_InvalidItemPrice(t *testing.T) {
	items := []invoice.ItemInput{{ItemID: "item-1", Quantity: 1, UnitPrice: -5.0}}
	_, err := invoice.New("customer-1", time.Now(), time.Now().AddDate(0, 0, 30), items)
	if err == nil {
		t.Error("expected error for unit_price <= 0")
	}
}

// --- MarkAsPaid ---

func TestMarkAsPaid_HappyPath(t *testing.T) {
	inv := newTestInvoice(t)
	if err := inv.MarkAsPaid(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if inv.Status != invoice.StatusPaid {
		t.Errorf("expected PAID, got %s", inv.Status)
	}
}

func TestMarkAsPaid_AlreadyPaid(t *testing.T) {
	inv := newTestInvoice(t)
	_ = inv.MarkAsPaid()
	err := inv.MarkAsPaid()
	if err != invoice.ErrAlreadyPaid {
		t.Errorf("expected ErrAlreadyPaid, got %v", err)
	}
}

func TestMarkAsPaid_AlreadyCanceled(t *testing.T) {
	inv := newTestInvoice(t)
	_ = inv.Cancel()
	err := inv.MarkAsPaid()
	if err != invoice.ErrAlreadyCanceled {
		t.Errorf("expected ErrAlreadyCanceled, got %v", err)
	}
}

// --- Cancel ---

func TestCancel_HappyPath(t *testing.T) {
	inv := newTestInvoice(t)
	if err := inv.Cancel(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if inv.Status != invoice.StatusCanceled {
		t.Errorf("expected CANCELED, got %s", inv.Status)
	}
}

func TestCancel_AlreadyCanceled(t *testing.T) {
	inv := newTestInvoice(t)
	_ = inv.Cancel()
	err := inv.Cancel()
	if err != invoice.ErrAlreadyCanceled {
		t.Errorf("expected ErrAlreadyCanceled, got %v", err)
	}
}

func TestCancel_AlreadyPaid(t *testing.T) {
	inv := newTestInvoice(t)
	_ = inv.MarkAsPaid()
	err := inv.Cancel()
	if err != invoice.ErrAlreadyPaid {
		t.Errorf("expected ErrAlreadyPaid, got %v", err)
	}
}