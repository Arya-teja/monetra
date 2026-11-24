package domain

type CashflowType string

const (
	CashflowIncome  CashflowType = "income"
	CashflowExpense CashflowType = "expense"
)

type Category struct {
	Name      string
	Type      CashflowType
	IsDefault bool
}
