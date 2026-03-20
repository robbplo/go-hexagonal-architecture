package model

import (
	"testing"

	domainerrors "github.com/linkai/go-chatbot-api/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateKnowledgeTokenBudget(t *testing.T) {
	tests := []struct {
		name       string
		total      int
		max        int
		wantErr    bool
		expectedIs error
	}{
		{
			name:  "within budget",
			total: 10,
			max:   20,
		},
		{
			name:       "exceeds budget",
			total:      30,
			max:        20,
			wantErr:    true,
			expectedIs: domainerrors.ErrKnowledgeLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKnowledgeTokenBudget(tt.total, tt.max)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedIs)
				return
			}
			require.NoError(t, err)
		})
	}
}
