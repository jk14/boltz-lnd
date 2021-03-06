package database

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/BoltzExchange/boltz-lnd/boltz"
	"github.com/btcsuite/btcd/btcec"
	"strconv"
	"strings"
)

type Swap struct {
	Id                  string
	Status              boltz.SwapUpdateEvent
	PrivateKey          *btcec.PrivateKey
	Preimage            []byte
	RedeemScript        []byte
	Invoice             string
	Address             string
	ExpectedAmount      int
	TimoutBlockHeight   int
	LockupTransactionId string
	RefundTransactionId string
}

type SwapSerialized struct {
	Id                  string
	Status              string
	PrivateKey          string
	Preimage            string
	RedeemScript        string
	Invoice             string
	Address             string
	ExpectedAmount      int
	TimeoutBlockHeight  int
	LockupTransactionId string
	RefundTransactionId string
}

func (swap *Swap) Serialize() SwapSerialized {
	preimage := ""

	if swap.Preimage != nil {
		preimage = hex.EncodeToString(swap.Preimage)
	}

	return SwapSerialized{
		Id:                  swap.Id,
		Status:              swap.Status.String(),
		PrivateKey:          formatPrivateKey(swap.PrivateKey),
		Preimage:            preimage,
		RedeemScript:        hex.EncodeToString(swap.RedeemScript),
		Invoice:             swap.Invoice,
		Address:             swap.Address,
		ExpectedAmount:      swap.ExpectedAmount,
		TimeoutBlockHeight:  swap.TimoutBlockHeight,
		LockupTransactionId: swap.LockupTransactionId,
		RefundTransactionId: swap.RefundTransactionId,
	}
}

func parseSwap(rows *sql.Rows) (*Swap, error) {
	var swap Swap

	var status string
	var privateKey string
	var preimage string
	var redeemScript string

	err := rows.Scan(
		&swap.Id,
		&status,
		&privateKey,
		&preimage,
		&redeemScript,
		&swap.Invoice,
		&swap.Address,
		&swap.ExpectedAmount,
		&swap.TimoutBlockHeight,
		&swap.LockupTransactionId,
		&swap.RefundTransactionId,
	)

	if err != nil {
		return nil, err
	}

	swap.Status = boltz.ParseEvent(status)

	privateKeyBytes, err := hex.DecodeString(privateKey)

	if err != nil {
		return nil, err
	}

	swap.PrivateKey, _ = parsePrivateKey(privateKeyBytes)

	if preimage != "" {
		swap.Preimage, err = hex.DecodeString(preimage)

		if err != nil {
			return nil, err
		}
	}

	swap.RedeemScript, err = hex.DecodeString(redeemScript)

	return &swap, err
}

func (database *Database) QuerySwap(id string) (swap *Swap, err error) {
	rows, err := database.db.Query("SELECT * FROM swaps WHERE id = '" + id + "'")

	if err != nil {
		return swap, err
	}

	defer rows.Close()

	if rows.Next() {
		swap, err = parseSwap(rows)

		if err != nil {
			return swap, err
		}
	} else {
		return swap, errors.New("could not find Swap " + id)
	}

	return swap, err
}

func (database *Database) querySwaps(query string) (swaps []Swap, err error) {
	rows, err := database.db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		swap, err := parseSwap(rows)

		if err != nil {
			return nil, err
		}

		swaps = append(swaps, *swap)
	}

	return swaps, err
}

func (database *Database) QuerySwaps() ([]Swap, error) {
	return database.querySwaps("SELECT * FROM swaps")
}

func (database *Database) QueryPendingSwaps() ([]Swap, error) {
	return database.querySwaps("SELECT * FROM swaps WHERE status NOT IN ('" + strings.Join(boltz.CompletedStatus, "','") + "')")
}

func (database *Database) QueryRefundableSwaps(currentBlockHeight uint32) ([]Swap, error) {
	return database.querySwaps("SELECT * FROM swaps WHERE status NOT IN ('" + strings.Join(boltz.CompletedStatus, "','") + "') AND timeoutBlockHeight <= " + strconv.FormatUint(uint64(currentBlockHeight), 10))
}

func (database *Database) CreateSwap(swap Swap) error {
	insertStatement := "INSERT INTO swaps (id, status, privateKey, preimage, redeemScript, invoice, address, expectedAmount, timeoutBlockheight, lockupTransactionId, refundTransactionId) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	statement, err := database.db.Prepare(insertStatement)

	if err != nil {
		return err
	}

	preimage := ""

	if swap.Preimage != nil {
		preimage = hex.EncodeToString(swap.Preimage)
	}

	_, err = statement.Exec(
		swap.Id,
		swap.Status.String(),
		formatPrivateKey(swap.PrivateKey),
		preimage,
		hex.EncodeToString(swap.RedeemScript),
		swap.Invoice,
		swap.Address,
		swap.ExpectedAmount,
		swap.TimoutBlockHeight,
		swap.LockupTransactionId,
		swap.RefundTransactionId,
	)

	if err != nil {
		return err
	}

	return statement.Close()
}

func (database *Database) UpdateSwapStatus(swap *Swap, status boltz.SwapUpdateEvent) error {
	swap.Status = status

	_, err := database.db.Exec("UPDATE swaps SET status = ? WHERE id = ?", status.String(), swap.Id)
	return err
}

func (database *Database) SetSwapInvoice(swap *Swap, invoice string) error {
	swap.Invoice = invoice

	_, err := database.db.Exec("UPDATE swaps SET invoice = ? WHERE id = ?", invoice, swap.Id)
	return err
}

func (database *Database) SetSwapLockupTransactionId(swap *Swap, lockupTransactionId string) error {
	swap.LockupTransactionId = lockupTransactionId

	_, err := database.db.Exec("UPDATE swaps SET lockupTransactionId = ? WHERE id = ?", lockupTransactionId, swap.Id)
	return err
}

func (database *Database) SetSwapRefundTransactionId(swap *Swap, refundTransactionId string) error {
	swap.RefundTransactionId = refundTransactionId

	_, err := database.db.Exec("UPDATE swaps SET status = ?, refundTransactionId = ? WHERE id = ?", boltz.SwapRefunded.String(), refundTransactionId, swap.Id)
	return err
}
