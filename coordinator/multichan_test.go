package main

import (
	"testing"
	"time"
)

// TestMultichanSendOnAbandonChan: send to abandon chan shoud return err
func TestMultichanSendOnAbandonChan(t *testing.T) {
	t.Skip()
	mc := NewMultiChan()

	if err := mc.Send("chan13", nil, 2*time.Second); err == nil {
		t.Fatalf("should be error, got nil")
	}
}

func TestMultichanSendTimeout(t *testing.T) {
	mc := NewMultiChan()

	go func() {
		if err := mc.Send("chan21", nil, 0); err != nil {
			t.Fatalf("expect no error, got %v", err)
		}
	}()
	mc.Recv("chan21", 0)
}

func TestReceiveTimeout(t *testing.T) {
	mc := NewMultiChan()
	then := time.Now()
	if _, err := mc.Recv("chan21", 1*time.Millisecond); err == nil {
		t.Fatalf("expect err, got nil")
	}
	if time.Since(then) > 1*time.Second {
		t.Fatalf("too long, expected 1 ms")
	}
}

func TestReceiveMultiple(t *testing.T) {
	mc := NewMultiChan()

	go func() {
		if err := mc.Send("chan21", nil, 0); err != nil {
			t.Fatalf("expect no error, got %v", err)
		}
		if err := mc.Send("chan21", nil, 0); err != nil {
			t.Fatalf("expect no error, got %v", err)
		}
		if err := mc.Send("chan21", nil, 0); err != nil {
			t.Fatalf("expect no error, got %v", err)
		}
	}()
	mc.Recv("chan21", 0)
	mc.Recv("chan21", 0)
	mc.Recv("chan21", 0)
}

func TestReceiveMultiple2(t *testing.T) {
	mc := NewMultiChan()

	done := make(chan bool)
	go func() {
		if err := mc.Send("chan21", nil, 0); err != nil {
			t.Fatalf("expect no error, got %v", err)
		}

		if err := mc.Send("chan21", nil, 0); err == nil {
			t.Fatalf("expect err, got nil")
		}
		<-done
	}()
	mc.Recv("chan21", 0)
	done<-true
	time.Sleep(100 * time.Millisecond)
}
