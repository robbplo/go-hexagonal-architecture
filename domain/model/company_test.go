package model

import (
	"testing"
	"time"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCompanyDeletion(t *testing.T) {
	tests := []struct {
		name        string
		activeUsers int
		wantErr     error
	}{
		{
			name:        "allows delete with no active users",
			activeUsers: 0,
		},
		{
			name:        "blocks delete with active users",
			activeUsers: 2,
			wantErr:     domainerrors.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCompanyDeletion(tt.activeUsers)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestNewCompany_InvalidBudget(t *testing.T) {
	_, err := NewCompany("cmp_1", "Acme", -1, time.Now())
	require.Error(t, err)
	assert.ErrorContains(t, err, "monthly_token_budget")
}
