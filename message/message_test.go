package message_test

import (
	"auditor/message"
	"bufio"
	"log"
	"os"
	"testing"
)

func TestNym(t *testing.T) {
}

func TestContract(t *testing.T) {
	lines := readFile("fixtures/contract")

	m := message.NewMessage(lines)

	if m.Type != "SIGNED CONTRACT" {
		t.Error("Expected 'SIGNED CONTRACT', got ", m.Type)
	}

	if m.Payload != "something sane" {
		t.Error("Expected 'something sane', got ", m.Payload)
	}

	if len(m.Signatures) != 1 || m.Signatures[0] != "something sane" {
		t.Error("Expected lenght 1 and content '', got ", m.Signatures)
	}
}

func TestAccount(t *testing.T) {
	lines := readFile("fixtures/account")

	m := message.NewMessage(lines)

	if m.Type != "SIGNED ACCOUNT" {
		t.Error("Expected 'SIGNED ACCOUNT', got ", m.Type)
	}
}

func TestInbox(t *testing.T) {
	lines := readFile("fixtures/inbox")

	m := message.NewMessage(lines)

	if m.Type != "SIGNED LEDGER" {
		t.Error("Expected 'SIGNED LEDGER', got ", m.Type)
	}
}

func TestReceiptSuccess(t *testing.T) {
	lines := readFile("fixtures/receipt.success")

	m := message.NewMessage(lines)

	if m.Type != "SIGNED TRANSACTION" {
		t.Error("Expected 'SIGNED TRANSACTION', got ", m.Type)
	}
}

func readFile(file string) []string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("err %+v\n", err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}
