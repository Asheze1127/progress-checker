package postgres

import (
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

func TestQuestionRepositoryImplementsInterface(t *testing.T) {
	// Compile-time check is done via var _ above, but this test
	// verifies it explicitly for clarity.
	var _ entities.QuestionRepository = (*QuestionRepository)(nil)
}

func TestParseMentorIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []entities.MentorID
	}{
		{
			name:     "empty braces",
			input:    "{}",
			expected: nil,
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single mentor",
			input:    "{mentor-1}",
			expected: []entities.MentorID{"mentor-1"},
		},
		{
			name:     "multiple mentors",
			input:    "{mentor-1,mentor-2,mentor-3}",
			expected: []entities.MentorID{"mentor-1", "mentor-2", "mentor-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMentorIDs(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d mentor IDs, got %d", len(tt.expected), len(result))
			}

			for i, id := range result {
				if id != tt.expected[i] {
					t.Errorf("expected mentor ID %q at index %d, got %q", tt.expected[i], i, id)
				}
			}
		})
	}
}
