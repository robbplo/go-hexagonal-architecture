package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompanyMonthUsage_WithAdjustment(t *testing.T) {
	tests := []struct {
		name    string
		usage   CompanyMonthUsage
		delta   int64
		want    int64
		wantErr bool
	}{
		{
			name: "positive adjustment",
			usage: CompanyMonthUsage{
				CompanyID:    "cmp_1",
				BudgetTokens: 100,
				InputTokens:  10,
				OutputTokens: 5,
			},
			delta: 2,
			want:  17,
		},
		{
			name: "negative below zero fails",
			usage: CompanyMonthUsage{
				CompanyID:    "cmp_1",
				BudgetTokens: 100,
				InputTokens:  10,
				OutputTokens: 5,
			},
			delta:   -20,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.usage.WithAdjustment(tt.delta, time.Now())
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.EffectiveUsage())
		})
	}
}

func TestCompanyMonthUsage_Reset(t *testing.T) {
	usage := CompanyMonthUsage{
		CompanyID:    "cmp_1",
		BudgetTokens: 100,
		InputTokens:  30,
		OutputTokens: 20,
	}

	got := usage.Reset(time.Now())
	assert.Equal(t, int64(0), got.EffectiveUsage())
}
