package newrelic

import "github.com/newrelic/go-agent/v3/newrelic"

func NoticeError(err error) {
	if err == nil {
		return
	}

	txn := NRApp.StartTransaction("error")
	defer txn.End()

	txn.NoticeError(err)
}

// NoticeErrorWithTransaction reports an error to New Relic with existing transaction
func NoticeErrorWithTransaction(txn *newrelic.Transaction, err error) {
	if err == nil || txn == nil {
		return
	}

	txn.NoticeError(err)
}
