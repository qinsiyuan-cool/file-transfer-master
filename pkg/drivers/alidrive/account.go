package alidrive

type Account struct {
	RefreshToken string
}

func (Account) Get(feild string) string {
	return Account{}.RefreshToken
}
