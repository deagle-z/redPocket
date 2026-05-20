package repository

import "testing"

func TestWithdrawAccountBindingIdentitiesExtractsAccountAndNameVariants(t *testing.T) {
	accountData := `{
		"accNo": " 123456 ",
		"accNoBrlW": "111222",
		"accNoMXNW": "987654",
		"accNameMxnW": "Maria",
		"accNameBrlW": "Joao",
		"bankCode": "001"
	}`

	got := withdrawAccountBindingIdentities(accountData)
	for _, want := range []withdrawAccountBindingIdentity{
		{AccountNo: "123456"},
		{AccountNo: "111222", AccountName: "Joao"},
		{AccountNo: "987654", AccountName: "Maria"},
	} {
		if !got[want] {
			t.Fatalf("withdrawAccountBindingIdentities() missing %#v in %#v", want, got)
		}
	}
	if got[withdrawAccountBindingIdentity{AccountNo: "001"}] {
		t.Fatalf("withdrawAccountBindingIdentities() included non-account field: %#v", got)
	}
}

func TestWithdrawAccountBindingIdentitiesIgnoresInvalidJSON(t *testing.T) {
	got := withdrawAccountBindingIdentities(`{`)
	if len(got) != 0 {
		t.Fatalf("withdrawAccountBindingIdentities() = %#v, want empty map", got)
	}
}

func TestWithdrawAccountBindingIdentitiesIgnoresEmptyValues(t *testing.T) {
	got := withdrawAccountBindingIdentities(`{"accNo": null, "accNoMXNW": " ", "accNameMxnW": "Maria"}`)
	if len(got) != 0 {
		t.Fatalf("withdrawAccountBindingIdentities() = %#v, want empty map", got)
	}
}

func TestWithdrawAccountAlreadyBoundErrorUsesI18nKey(t *testing.T) {
	if withdrawAccountAlreadyBoundMessage != "withdraw_account_already_bound" {
		t.Fatalf("withdrawAccountAlreadyBoundMessage = %q, want i18n key", withdrawAccountAlreadyBoundMessage)
	}
}
