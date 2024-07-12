package metadata

import "testing"

type testData struct{}

func (t *testData) SetTaskID(taskID string) {}
func (t *testData) TableName() string       { return "" }

func TestFSM_Draw(t *testing.T) {
	var (
		New      = State[*testData]{Name: "New"}
		Frozen   = State[*testData]{Name: "Frozen"}
		Audit    = State[*testData]{Name: "Audit"}
		Approved = State[*testData]{Name: "Approved"}
		Rejected = State[*testData]{Name: "Rejected", IsFinal: true}
		Pay      = State[*testData]{Name: "Pay"}
		PaySucc  = State[*testData]{Name: "PaySucc", IsFinal: true}
		PayFail  = State[*testData]{Name: "PayFail", IsFinal: true}
	)
	fsm := GenFSM("Audits", New)
	fsm.RegisterTransition(
		GenTransition(New, Frozen),
		GenTransition(Frozen, Audit),
		GenTransition(Audit, Approved),
		GenTransition(Audit, Rejected),
		GenTransition(Approved, Pay),
		GenTransition(Pay, PaySucc),
		GenTransition(Pay, PayFail),
	)
	if err := fsm.Draw("audits.svg"); err != nil {
		t.Fatal(err)
	}
}
