package accountsecurity

import (
	"context"
	"errors"
	"testing"

	"wecheckin/backend/pkg/passwordutil"
)

type fakePasswordStore struct {
	credential Credential
	history    []string
	updates    []PasswordUpdate
	err        error
}

func (f *fakePasswordStore) WithinTransaction(ctx context.Context, fn func(PasswordStore) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(f)
}

func (f *fakePasswordStore) CredentialForUpdate(context.Context, uint) (Credential, error) {
	return f.credential, f.err
}

func (f *fakePasswordStore) RecentPasswordHashes(context.Context, uint, int) ([]string, error) {
	return append([]string(nil), f.history...), f.err
}

func (f *fakePasswordStore) ApplyPasswordUpdate(_ context.Context, update PasswordUpdate) error {
	if f.err != nil {
		return f.err
	}
	f.updates = append(f.updates, update)
	return nil
}

func passwordHash(t *testing.T, plain string) string {
	t.Helper()
	hash, err := passwordutil.Hash(plain)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func enabledPolicy() Policy {
	policy := DefaultPolicy()
	policy.Enabled = true
	policy.EffectiveAt = 1
	return policy
}

func TestChangeOwnPasswordEnforcesStrengthHistoryAndCurrentPassword(t *testing.T) {
	store := &fakePasswordStore{credential: Credential{UserID: 7, PasswordHash: passwordHash(t, "Old12345")}}
	service := NewPasswordService(store, func(context.Context) (Policy, error) { return enabledPolicy(), nil })
	if err := service.ChangeOwnPasswordContext(context.Background(), 7, "wrong", "New12345"); !errors.Is(err, ErrCurrentPasswordInvalid) {
		t.Fatalf("wrong current password err=%v", err)
	}
	if err := service.ChangeOwnPasswordContext(context.Background(), 7, "Old12345", "weak"); err == nil {
		t.Fatal("weak permanent password accepted")
	}
	store.history = []string{passwordHash(t, "Used12345")}
	if err := service.ChangeOwnPasswordContext(context.Background(), 7, "Old12345", "Used12345"); !errors.Is(err, ErrPasswordReused) {
		t.Fatalf("history reuse err=%v", err)
	}
	if err := service.ChangeOwnPasswordContext(context.Background(), 7, "Old12345", "Fresh12345"); err != nil {
		t.Fatal(err)
	}
	if len(store.updates) != 1 || store.updates[0].MustChange || store.updates[0].ChangeSource != PasswordChangeSourceSelf {
		t.Fatalf("password update = %#v", store.updates)
	}
}

func TestTemporaryPasswordsMayBypassStrengthButNotHistory(t *testing.T) {
	policy := enabledPolicy()
	store := &fakePasswordStore{credential: Credential{UserID: 9, PasswordHash: passwordHash(t, "Existing123")}}
	service := NewPasswordService(store, func(context.Context) (Policy, error) { return policy, nil })
	if err := service.ResetTemporaryPasswordContext(context.Background(), 9, "123456"); err != nil {
		t.Fatalf("temporary password rejected: %v", err)
	}
	if len(store.updates) != 1 || !store.updates[0].MustChange || store.updates[0].ChangeSource != PasswordChangeSourceAdminReset {
		t.Fatalf("temporary update = %#v", store.updates)
	}
	store.updates = nil
	store.history = []string{passwordHash(t, "123456")}
	if err := service.ResetTemporaryPasswordContext(context.Background(), 9, "123456"); !errors.Is(err, ErrPasswordReused) {
		t.Fatalf("temporary history reuse err=%v", err)
	}
}
